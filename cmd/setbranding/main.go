package main

import (
	"bytes"
	"context"
	"fmt"
	"github.com/alecthomas/kong"
	"github.com/fatih/color"
	"github.com/wttw/aboutmyemail"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"time"
)

type Globals struct {
	Server  string      `env:"MYEMAIL_SERVER" help:"The api endpoint to use" default:"https://whitelabel.aboutmy.email/api/v1"`
	ApiKey  string      `env:"MYEMAIL_APIKEY" help:"The api key to use for authorization"`
	Quiet   bool        `help:"Don't display parameters or progress"`
	Version VersionFlag `name:"version" help:"Print version information and quit"`
}

type VersionFlag string

func (v VersionFlag) Decode(_ *kong.DecodeContext) error { return nil }
func (v VersionFlag) IsBool() bool                       { return true }
func (v VersionFlag) BeforeApply(app *kong.Kong, vars kong.Vars) error {
	fmt.Println(vars["version"])
	app.Exit(0)
	return nil
}

type StageCmd struct {
	Files []string `arg:"" type:"existingfile"`
}

type PublishCmd struct {
}

type CLI struct {
	Globals

	Stage   StageCmd   `cmd:"" help:"Upload markdown and json files to set branding on staging server"`
	Publish PublishCmd `cmd:"" help:"Publish branding from staging server to production"`
}

func main() {
	cli := CLI{}
	ctx := kong.Parse(&cli,
		kong.Name("setbranding"),
		kong.Description("Tool to set white label branding for aboutmy.email"),
		kong.UsageOnError(),
		kong.ConfigureHelp(kong.HelpOptions{Compact: true}),
		kong.Vars{
			"version": "0.1",
		})
	err := ctx.Run(&cli.Globals)
	ctx.FatalIfErrorf(err)
}

func printError(format string, args ...any) {
	red := color.New(color.FgHiRed).SprintFunc()
	_, _ = fmt.Fprintf(color.Output, "%s: %s\n", red("ERROR"), fmt.Sprintf(format, args...))
}

func printWarning(format string, args ...any) {
	yellow := color.New(color.FgHiYellow).SprintFunc()
	_, _ = fmt.Fprintf(color.Output, "%s: %s\n", yellow("WARN"), fmt.Sprintf(format, args...))
}

func fatal(format string, args ...any) {
	printError(format, args...)
	os.Exit(1)
}

func (a *StageCmd) Run(globals *Globals) {
	var buff bytes.Buffer
	mpw := multipart.NewWriter(&buff)
	for _, file := range a.Files {
		ext := filepath.Ext(file)
		if ext != ".md" && ext != ".json" {
			printWarning("Skipping non-markdown, non-json file: %s", file)
			continue
		}
		f, err := os.Open(file)
		if err != nil {
			printError("Failed to open %s, skipping: %s", file, err)
			continue
		}
		if !globals.Quiet {
			blue := color.New(color.FgHiBlue).SprintFunc()
			_, _ = fmt.Fprintf(color.Output, "Uploading %s as %s\n", file, blue(filepath.Base(file)))
		}
		writer, err := mpw.CreateFormFile("filename", filepath.Base(file))
		if err != nil {
			fatal("Failed to create upload: %s", err)
		}
		_, err = io.Copy(writer, f)
		if err != nil {
			fatal("Failed to create upload: %s", err)
		}
	}

	err := mpw.Close()
	if err != nil {
		fatal("Failed to create upload: %s", err)
	}
	client, err := aboutmyemail.New(aboutmyemail.WithServer(globals.Server), aboutmyemail.WithApiKey(globals.ApiKey))
	if err != nil {
		fatal("Failed to create client: %s", err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 300*time.Second)
	defer cancel()

	response, err := client.ContentPostWithBodyWithResponse(ctx, mpw.FormDataContentType(), bytes.NewReader(buff.Bytes()))
	if err != nil {
		fatal("Upload failed: %s", err)
	}

	if response.StatusCode() == http.StatusOK {
		if !globals.Quiet {
			green := color.New(color.FgHiGreen).SprintFunc()
			_, _ = fmt.Fprintf(color.Output, "%s\n", green("Uploaded OK"))
		}
	} else {
		printWarning("Server responded with %s", response.HTTPResponse.Status)
		if response.JSON400 != nil {
			printWarning("%s", response.JSON400.Message)
		} else if response.JSON500 != nil {
			printError("%s", response.JSON500.Message)
		} else {
			printError("%s", response.Body)
		}
	}
}

func (a *PublishCmd) Run(globals *Globals) error {
	client, err := aboutmyemail.New(aboutmyemail.WithServer(globals.Server), aboutmyemail.WithApiKey(globals.ApiKey))
	if err != nil {
		fatal("Failed to create client: %s", err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	response, err := client.StylePublishWithResponse(ctx)
	if err != nil {
		fatal("Failed to publish: %s", err)
	}
	if response.StatusCode() == http.StatusOK {
		if !globals.Quiet {
			green := color.New(color.FgHiGreen).SprintFunc()
			_, _ = fmt.Fprintf(color.Output, "%s\n", green("Published OK"))
		}
	} else {
		printWarning("Server responded with %s", response.HTTPResponse.Status)
		if response.JSON400 != nil {
			printWarning("%s", response.JSON400.Message)
		} else if response.JSON500 != nil {
			printError("%s", response.JSON500.Message)
		} else {
			printError("%s", response.Body)
		}
	}
	return nil
}
