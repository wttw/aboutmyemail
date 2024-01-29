package main

import (
	"embed"
	"fmt"
	"github.com/alecthomas/kong"
	"github.com/carlmjohnson/versioninfo"
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

type CLI struct {
	Globals

	Init    InitCmd    `cmd:"" help:"Initialize a directory with default content"`
	Stage   StageCmd   `cmd:"" help:"Upload markdown and json files to set branding on staging server"`
	Publish PublishCmd `cmd:"" help:"Publish branding from staging server to production"`
	Lint    LintCmd    `cmd:"" help:"Check for basic errors in files"`
	Dns     DnsCmd     `cmd:"" help:"Check DNS setup for a hostname"`
}

func main() {
	cli := CLI{}
	ctx := kong.Parse(&cli,
		kong.Name("amemanage"),
		kong.Description("Tool to manage white label branding for aboutmy.email"),
		kong.UsageOnError(),
		kong.ConfigureHelp(kong.HelpOptions{Compact: true}),
		kong.Vars{
			"version": versioninfo.Short(),
		})
	err := ctx.Run(&cli.Globals)
	ctx.FatalIfErrorf(err)
}
