package main

import (
	"github.com/jawher/mow.cli"
	"github.com/rnbdev/figo/core"
	"os"
)

func main() {
	app := cli.App("figos", "throw file to figor")
	app.Spec = "NAME [NICK]"

	name := app.String(cli.StringArg{
		Name:      "NAME",
		Desc:      "filename to throw",
		HideValue: true,
	})

	nick := app.String(cli.StringArg{
		Name:      "NICK",
		Desc:      "nickname from figor",
		HideValue: true,
	})

	app.Action = func() {
		core.Figos(*name, *nick)
	}

	app.Run(os.Args)
	// todo: interface, address, port, timeout, quiet
}
