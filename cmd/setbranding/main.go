package main

import (
	"bytes"
	"context"
	"embed"
	"encoding/json"
	"fmt"
	"github.com/alecthomas/kong"
	"github.com/carlmjohnson/versioninfo"
	"github.com/fatih/color"
	"github.com/wttw/aboutmyemail"
	"golang.org/x/text/language/display"
	"io"
	"io/fs"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"slices"
	"strings"
	"text/template"
	"time"
)

//go:embed templates
var templateFS embed.FS

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
	Force bool     `help:"Upload files even if tests fail"`
}

type PublishCmd struct {
}

type InitCmd struct {
	Hostname  string `help:"Generate sample content for a site at this hostname" required:""`
	Directory string `help:"Generate files in this directory" required:""`
}

type LintCmd struct {
	Files []string `arg:"" type:"existingfile"`
}

type CLI struct {
	Globals

	Init    InitCmd    `cmd:"" help:"Initialize a directory with default content"`
	Stage   StageCmd   `cmd:"" help:"Upload markdown and json files to set branding on staging server"`
	Publish PublishCmd `cmd:"" help:"Publish branding from staging server to production"`
	Lint    LintCmd    `cmd:"" help:"Check for basic errors in files"`
}

func main() {
	cli := CLI{}
	ctx := kong.Parse(&cli,
		kong.Name("setbranding"),
		kong.Description("Tool to set white label branding for aboutmy.email"),
		kong.UsageOnError(),
		kong.ConfigureHelp(kong.HelpOptions{Compact: true}),
		kong.Vars{
			"version": versioninfo.Short(),
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

func printSuccess(globals *Globals, msg string, args ...any) {
	if !globals.Quiet {
		green := color.New(color.FgHiGreen).SprintFunc()
		_, _ = fmt.Fprintf(color.Output, "%s\n", green(fmt.Sprintf(msg, args...)))
	}
}

func fatal(format string, args ...any) {
	printError(format, args...)
	os.Exit(1)
}

func (a *StageCmd) Run(globals *Globals) error {
	if lint(a.Files, globals) {
		if !a.Force {
			fatal("Tests failed, not uploading. Use --force to override")
		}
		printWarning("Tests failed, continuing anyway")
	}
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
		printSuccess(globals, "Uploaded OK")
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
		printSuccess(globals, "Published OK")
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

var hostRe = regexp.MustCompile(`(?:https?://)?([a-zA-Z0-9-]+)\.([a-zA-Z0-9-]+)\.([a-zA-Z0-9-]+)`)

func (a *InitCmd) Run(globals *Globals) error {
	matches := hostRe.FindStringSubmatch(a.Hostname)
	var baseURL, brand string
	if matches == nil {
		baseURL = "https://e.g.example.net"
		brand = "Your Brand"
	} else {
		baseURL = fmt.Sprintf("https://%s.%s", matches[2], matches[3])
		brand = strings.ToLower(matches[2])
	}
	err := os.MkdirAll(a.Directory, 0755)
	if err != nil {
		return err
	}
	err = fs.WalkDir(templateFS, "templates", func(path string, d fs.DirEntry, e error) error {
		if e != nil {
			return e
		}
		if d.IsDir() {
			return nil
		}
		tpl, err := template.New(d.Name()).ParseFS(templateFS, path)
		if err != nil {
			return err
		}
		f, err := os.Create(filepath.Join(a.Directory, strings.TrimSuffix(d.Name(), ".tpl")))
		if err != nil {
			return err
		}
		err = tpl.Execute(f, struct {
			BaseURL   string
			BrandName string
		}{
			BaseURL:   baseURL,
			BrandName: brand,
		})
		if err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		fatal("Failed to generate content: %s", err)
	}
	printSuccess(globals, "Initialized OK")
	return nil
}

func (a LintCmd) Run(globals *Globals) error {
	if !lint(a.Files, globals) {
		printSuccess(globals, "Linted OK")
	}
	return nil
}

func lint(files []string, _ *Globals) bool {
	failed := false
	expected := map[string]int{}
	err := fs.WalkDir(templateFS, "templates", func(path string, d fs.DirEntry, e error) error {
		if e != nil {
			return e
		}
		if !d.IsDir() {
			expected[strings.TrimSuffix(d.Name(), ".tpl")] = 1
		}
		return nil
	})
	if err != nil {
		fatal("%s", err)
	}
	splitRe := regexp.MustCompile(`^([a-zA-Z0-9]+)\.([a-zA-Z0-9_-]+)\.md$`)
	languages := map[string]struct{}{}
	for _, file := range files {
		base := filepath.Base(file)
		ext := filepath.Ext(file)
		switch ext {
		case ".json":
			_, ok := expected[base]
			if !ok {
				printWarning("Unexpected file: %s", base)
				failed = true
				continue
			}
			expected[base]++
			if checkJson(file) {
				failed = true
			}
		case ".md":
			baseFile := base
			matches := splitRe.FindStringSubmatch(base)
			if matches != nil {
				baseFile = matches[1] + ".md"
			}
			_, ok := expected[baseFile]
			if !ok {
				printWarning("Unexpected file: %s", base)
				failed = true
				continue
			}
			if matches != nil {
				languages[matches[2]] = struct{}{}
			}
			expected[baseFile]++
		default:
			printWarning("Unexpected file: %s", base)
			failed = true
		}
	}
	supported := map[string]struct{}{}
	for _, tag := range display.Self.Supported.Tags() {
		supported[tag.String()] = struct{}{}
	}
	additionalLanguages := []string{"x-piglatin"}
	for lang := range languages {
		if slices.Contains(additionalLanguages, lang) {
			continue
		}
		_, ok := supported[lang]
		if !ok {
			printWarning("Unexpected language: '%s'", lang)
			failed = true
		}
	}
	return failed
}

func checkJson(file string) bool {
	failed := false
	f, err := os.Open(file)
	if err != nil {
		printError("Failed to open %s: %s", file, err)
		failed = true
	}
	j := map[string]string{}
	dec := json.NewDecoder(f)
	err = dec.Decode(&j)
	if err != nil {
		printError("Failed to parse %s: %s", file, err)
		failed = true
	}
	expectedKeys := []string{"HomeLink", "LogoHTML", "MobileLogoHTML"}
	for key := range j {
		if !slices.Contains(expectedKeys, key) {
			printWarning("Unexpected key in %s: %s", file, key)
			failed = true
		}
	}
	return failed
}
