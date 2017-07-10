package main

import (
	"github.com/jawher/mow.cli"
	"github.com/rnbdev/figo/core"
	"os"
)

func main() {
	app := cli.App("figor", "catch file from figos")
	app.Spec = "NAME [OUT]"

	name := app.String(cli.StringArg{
		Name:      "NAME",
		Desc:      "filename or nickname to catch",
		HideValue: true,
	})

	out := app.String(cli.StringArg{
		Name:      "OUT",
		Desc:      "save as `OUT`",
		HideValue: true,
	})

	app.Action = func() {
		core.Figor(*name, *out)
	}

	app.Run(os.Args)
	// todo: timeout, quiet
}
