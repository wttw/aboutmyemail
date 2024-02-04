package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/alecthomas/kong"
	"github.com/carlmjohnson/versioninfo"
	"github.com/fatih/color"
	"github.com/toqueteos/webbrowser"
	"github.com/wttw/aboutmyemail"
	"io"
	"net"
	"net/http"
	"net/mail"
	"os"
	"sync"
	"time"
)

type CLI struct {
	Server    string `env:"MYEMAIL_SERVER" help:"The api endpoint to use" default:"https://api.aboutmy.email/api/v1"`
	ApiKey    string `env:"MYEMAIL_APIKEY" help:"The api key to use for authorization"`
	Email     []byte `arg:"" help:"File containing raw email" type:"filecontent"`
	From      string `help:"Email address for return path" placeholder:"email@address"`
	To        string `help:"Email address for recipient" placeholder:"email@address"`
	Ip        string `help:"IP address of mailserver" placeholder:"dotted-quad"`
	Helo      string `help:"Value for mailserver HELO" placeholder:"host.name"`
	Ascii     bool   `help:"Disable internationalization"`
	Quiet     bool   `help:"Don't display parameters or progress"`
	Staged    bool   `help:"Display result using staged whitelabel configuration"`
	Open      bool   `help:"Open result in browser"`
	Callbacks string `help:"Start local webserver for callbacks" placeholder:"address:port"`
}

func main() {
	cli := CLI{}
	_ = kong.Parse(&cli,
		kong.Name("aboutmyemail"),
		kong.Description("Tool to submit messages via the aboutmy.email API"),
		kong.UsageOnError(),
		kong.ConfigureHelp(kong.HelpOptions{Compact: true}),
		kong.Vars{
			"version": versioninfo.Short(),
		})

	var waiter sync.WaitGroup

	if cli.From == "" || cli.To == "" {
		msg, err := mail.ReadMessage(bytes.NewReader(cli.Email))
		if err == nil {
			if cli.From == "" {
				returnPath, err := msg.Header.AddressList("ReturnPath")
				if err == nil && len(returnPath) > 0 {
					cli.From = returnPath[0].Address
				} else {
					from, err := msg.Header.AddressList("From")
					if err == nil && len(from) > 0 {
						cli.From = from[0].Address
					}
				}
			}
			if cli.To == "" {
				to, err := msg.Header.AddressList("To")
				if err == nil && len(to) > 0 {
					cli.To = to[0].Address
				}
			}
		}
	}
	if cli.Ip == "" {
		conn, err := net.Dial("udp", "8.8.8.8:80")
		if err == nil {

			cli.Ip = conn.LocalAddr().(*net.UDPAddr).IP.String()
		}
		_ = conn.Close()
	}
	if cli.Helo == "" {
		cli.Helo, _ = os.Hostname()
	}
	if !cli.Quiet {
		blue := color.New(color.FgHiBlue).SprintFunc()
		_, _ = fmt.Fprintf(color.Output, "From:    %s\n", blue(cli.From))
		_, _ = fmt.Fprintf(color.Output, "To:      %s\n", blue(cli.To))
		_, _ = fmt.Fprintf(color.Output, "IP:      %s\n", blue(cli.Ip))
		_, _ = fmt.Fprintf(color.Output, "Helo:    %s\n", blue(cli.Helo))
		_, _ = fmt.Fprintf(color.Output, "Payload: %s\n", blue(fmt.Sprintf("%d bytes", len(cli.Email))))
	}

	client, err := aboutmyemail.New(aboutmyemail.WithServer(cli.Server), aboutmyemail.WithApiKey(cli.ApiKey))
	if err != nil {
		fatal("Failed to create client: %s", err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	if cli.Callbacks != "" {
		listener, err := net.Listen("tcp", cli.Callbacks)
		if err != nil {
			fatal("Failed to start webserver on %s: %s", cli.Callbacks, err)
		}
		waiter.Add(1)
		go func() {
			callbackForResults(ctx, cli, listener)
			waiter.Done()
		}()
	}

	smtputf8 := !cli.Ascii
	var options string
	if cli.Staged {
		options = "stage"
	}

	request := aboutmyemail.EmailJSONRequestBody{
		From:     cli.From,
		Ip:       cli.Ip,
		Payload:  string(cli.Email),
		Smtputf8: &smtputf8,
		To:       cli.To,
		Options:  &options,
	}

	if cli.Callbacks != "" {
		url := fmt.Sprintf("http://%s/callback", cli.Callbacks)
		request.ProgressUrl = &url
		request.FinishedUrl = &url
	}

	response, err := client.EmailWithResponse(ctx, request)
	if err != nil {
		fatal("Failed to submit email: %s", err)
	}

	if response.StatusCode() != http.StatusOK {
		printError("Server rejected request: %s", response.HTTPResponse.Status)
		printResponse(response.Body, response.JSON500, response.JSON400)
		os.Exit(1)
	}

	if response.JSON200 == nil {
		fatal("Unexpected nil result in response")
	}

	id := response.JSON200.Id

	cyan := color.New(color.FgCyan).SprintFunc()
	if !cli.Quiet {
		_, _ = fmt.Fprintf(color.Output, "Processing %s ...\n", cyan(id))
	}

	if cli.Callbacks == "" {
		pollForResults(ctx, id, cli, client)
	}

	waiter.Wait()
}

// callbackForResults starts a local webserver and prints status updates it receives
func callbackForResults(ctx context.Context, cli CLI, listener net.Listener) {
	cyan := color.New(color.FgCyan).SprintFunc()
	green := color.New(color.FgGreen).SprintFunc()
	counter := 1
	var counterMtx sync.Mutex
	s := http.Server{}
	http.HandleFunc("/callback", func(w http.ResponseWriter, r *http.Request) {
		counterMtx.Lock()
		cnt := counter
		counter++
		counterMtx.Unlock()
		defer func(Body io.ReadCloser) {
			_ = Body.Close()
		}(r.Body)
		if r.Method != http.MethodPost {
			printError("Expected POST to callback, not %s", r.Method)
			http.Error(w, "Expected POST", http.StatusBadRequest)
			return
		}
		ct := r.Header.Get("Content-Type")
		if ct != "application/json" {
			printError("Expected application/json callback, not %s", ct)
			http.Error(w, "Expected application/json", http.StatusBadRequest)
			return
		}
		w.WriteHeader(http.StatusOK)
		var result aboutmyemail.StatusResult
		dec := json.NewDecoder(r.Body)
		err := dec.Decode(&result)
		if err != nil {
			printError("Failed to unmarshal callback: %s", err)
			return
		}
		if result.Messages != nil && !cli.Quiet {
			col := cyan
			if result.Url != nil && *result.Url != "" {
				col = green
			}
			for _, msg := range *result.Messages {
				_, _ = fmt.Fprintf(color.Output, "%d:  %s\n", cnt, col(msg))
			}
		}
		if result.Url != nil && *result.Url != "" {
			url := *result.Url
			if !cli.Quiet || !cli.Open {
				fmt.Printf("%s\n", url)
			}
			if cli.Open {
				err := webbrowser.Open(url)
				if err != nil {
					fatal("Failed to open browser: %s", err)
				}
			}
			_ = s.Shutdown(context.Background())
			return
		}
	})
	go func() {
		<-ctx.Done()
		_ = s.Shutdown(context.Background())
	}()
	_ = s.Serve(listener)
}

func pollForResults(ctx context.Context, id string, cli CLI, client *aboutmyemail.ClientWithResponses) {
	cyan := color.New(color.FgCyan).SprintFunc()
	for {
		response, err := client.EmailStatusWithResponse(ctx, id)
		if err != nil {
			fatal("While polling for result: %s", err)
		}
		if response.StatusCode() == http.StatusTooManyRequests {
			yellow := color.New(color.FgYellow).SprintFunc()
			_, _ = fmt.Fprintf(color.Output, "%s\n", yellow("throttled, sleeping"))
			time.Sleep(200 * time.Millisecond)
			continue
		}
		if response.StatusCode() != http.StatusOK {
			printError("Server rejected request: %s", response.HTTPResponse.Status)
			printResponse(response.Body, response.JSON500, response.JSON404)
			os.Exit(1)
		}
		if response.JSON200 == nil {
			fatal("Unexpected nil result in response")
		}
		if !cli.Quiet && response.JSON200.Messages != nil {
			for _, msg := range *response.JSON200.Messages {
				_, _ = fmt.Fprintf(color.Output, "  %s\n", cyan(msg))
			}
		}
		if response.JSON200.Url != nil && *response.JSON200.Url != "" {
			url := *response.JSON200.Url
			if !cli.Quiet || !cli.Open {
				fmt.Printf("%s\n", url)
			}
			if cli.Open {
				err := webbrowser.Open(url)
				if err != nil {
					fatal("Failed to open browser: %s", err)
				}
			}
			os.Exit(0)
		}
		time.Sleep(time.Second)
	}
}

func printError(format string, args ...any) {
	red := color.New(color.FgHiRed).SprintFunc()
	_, _ = fmt.Fprintf(color.Output, "%s: %s\n", red("ERROR"), fmt.Sprintf(format, args...))
}

func fatal(format string, args ...any) {
	printError(format, args...)
	os.Exit(1)
}

func printResponse(raw []byte, structureds ...any) {
	for _, s := range structureds {
		if s != nil {
			encoder := json.NewEncoder(os.Stdout)
			encoder.SetEscapeHTML(false)
			encoder.SetIndent("    ", "  ")
			_ = encoder.Encode(s)
			fmt.Printf("\n")
			return
		}
	}
	fmt.Printf("%s\n", string(raw))
}
