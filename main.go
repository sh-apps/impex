package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/example-pipeline/impex/cmd/npm"
	"golang.org/x/exp/slog"
)

var help = `Usage: impex [cmd]

  npm

Examples:

  impex npm /path/to/package-lock.json
`

func main() {
	args := os.Args[1:]
	var cmd string
	if len(args) > 0 {
		cmd = args[0]
	}

	var err error
	switch cmd {
	case "npm":
		npmCmd(args[1:])
	default:
		fmt.Println(help)
	}
	if err != nil {
		slog.Error("command failed", slog.Any("error", err))
	}
}

func npmCmd(args []string) {
	cmd := flag.NewFlagSet("npm", flag.ExitOnError)
	fileName := cmd.String("lock-file", "", "Path to the lock file.")
	helpFlag := cmd.Bool("help", false, "Print help and exit.")
	err := cmd.Parse(args)
	if err != nil || *helpFlag || fileName == nil || *fileName == "" {
		cmd.PrintDefaults()
		return
	}
	err = npm.Run(npm.Arguments{
		FileName: *fileName,
	})
	if err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}
}
