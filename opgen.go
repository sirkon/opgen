package main

import (
	"fmt"
	"os"
	"runtime/debug"

	"github.com/alecthomas/kong"
	"github.com/sirkon/message"
	"github.com/sirkon/opgen/internal/app"
)

func main() {
	for _, v := range os.Args[1:] {
		switch v {
		case "-v", "--version":
			info, ok := debug.ReadBuildInfo()
			var version string
			if !ok || info.Main.Version == "" {
				version = "(devel)"
			} else {
				version = info.Main.Version
			}
			fmt.Println(app.Name, "version", version)

			os.Exit(0)
		}
	}

	var cli cliDefinition
	parser := kong.Must(
		&cli,
		kong.Name(app.Name),
		kong.Description("Functional options builders generator"),
		kong.ConfigureHelp(kong.HelpOptions{
			Summary: true,
			Compact: true,
		}),
		kong.UsageOnError(),
	)

	if _, err := parser.Parse(os.Args[1:]); err != nil {
		parser.FatalIfErrorf(err)
	}

	var types []string
	for _, typeName := range cli.Types {
		types = append(types, string(typeName))
	}

	if err := generate(cli.OptionsSourcePackage, string(cli.Dest), types); err != nil {
		message.Error(err)
		os.Exit(1)
	}
}
