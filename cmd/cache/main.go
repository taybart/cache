package main

import (
	"fmt"
	"os"

	"github.com/taybart/args"
)

var (
	app = args.App{
		Name:    "Cache",
		Version: "v0.0.1",
		Author:  "TayBart <taybart@email.com>",
		About:   "cache server",
		Args: map[string]*args.Arg{
			"port": {
				Short:   "p",
				Help:    "Port to listen on",
				Default: 50400,
			},
		},
	}
)

func main() {
	if err := run(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
func run() error {

	return nil
}
