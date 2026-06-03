package main

import (
	"encoding/json"
	"fmt"
	"n8go-docs/diagnostics"
	"os"

	"github.com/alecthomas/kong"
)

// BuildCommand builds the documentation once and exits.
//
// Examples:
//
//	n8go-docs build
//	n8go-docs build --config ./docs/n8go-docs.yaml
//	n8go-docs build --config ./docs/n8go-docs.yaml --verbose
type BuildCommand struct {
	Verbose bool `short:"v" long:"verbose" help:"Print detailed progress to stderr"`
}

func (cmd *BuildCommand) Run(cli *CLI) error {
	diagnostics.SetVerbose(cmd.Verbose)
	err := runGenerator(cli.Config)
	if err != nil {
		if cli.JSON {
			printJSON(map[string]string{"status": "error", "message": err.Error()})
		}
		return err
	}
	if cli.JSON {
		printJSON(map[string]string{"status": "ok"})
	}
	return nil
}

// ServeCommand starts a live-reload development server.
//
// Examples:
//
//	n8go-docs serve
//	n8go-docs serve --port 8080
//	n8go-docs serve --config ./docs/n8go-docs.yaml --port 8080
type ServeCommand struct {
	Port    uint16 `short:"p" long:"port"    help:"Port to listen on" default:"9080" placeholder:"<port>"`
	Verbose bool   `short:"v" long:"verbose" help:"Print detailed progress to stderr"`
}

func (cmd *ServeCommand) Run(cli *CLI) error {
	diagnostics.SetVerbose(cmd.Verbose)
	return runServer(cli.Config, int(cmd.Port))
}

// VersionCommand prints the application version.
//
// Examples:
//
//	n8go-docs version
//	n8go-docs version --json
type VersionCommand struct{}

func (cmd *VersionCommand) Run(cli *CLI) error {
	if cli.JSON {
		printJSON(map[string]string{"version": appVersion})
		return nil
	}
	fmt.Fprintf(os.Stdout, "n8go-docs %s\n", appVersion)
	return nil
}

const appVersion = "0.1.0"
const defaultServePort = 9080

// CLI is the root command.
type CLI struct {
	Config string `short:"c" long:"config"  help:"Path to config file" default:"n8go-docs.yaml" type:"path" placeholder:"<path>"`
	JSON   bool   `          long:"json"    help:"Output machine-readable JSON (where supported)"`

	Build   BuildCommand   `cmd:"" name:"build"   help:"Build the documentation site" aliases:"generate"`
	Serve   ServeCommand   `cmd:"" name:"serve"   help:"Start live-reload development server"`
	Version VersionCommand `cmd:"" name:"version" help:"Print version and exit"`
}

var cli CLI

func main() {
	ctx := kong.Parse(&cli,
		kong.Name("n8go-docs"),
		kong.Description("Static documentation generator.\n\nExamples:\n  n8go-docs build\n  n8go-docs serve --port 8080\n  n8go-docs build --config <path> --verbose"),
		kong.UsageOnError(),
		kong.ConfigureHelp(kong.HelpOptions{
			Compact:             false,
			Summary:             false,
			FlagsLast:           true,
			NoExpandSubcommands: false,
		}),
	)
	err := ctx.Run(&cli)
	if err != nil {
		diagnostics.HandleError(err)
	}
	os.Exit(0)
}

func printJSON(v any) {
	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	_ = enc.Encode(v)
}
