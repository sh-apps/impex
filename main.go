package main

import (
	"flag"
	"fmt"
	"os"

	"log/slog"

	"github.com/example-pipeline/impex/cmd/container"
	"github.com/example-pipeline/impex/cmd/git"
	"github.com/example-pipeline/impex/cmd/npm"
	"github.com/example-pipeline/impex/cmd/vsix"
)

var help = `Usage: impex [cmd]

  npm
	vsix
	container
	git

Examples:

  impex npm -lock-file=/package-lock.json
  impex vsix -file=./vsix.txt
  impex container -file=./containers.txt
  impex git -file=./git.txt -accessToken=ghp_fdsfdsfd
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
	case "vsix":
		vsixCmd(args[1:])
	case "container":
		containerCmd(args[1:])
	case "containers":
		containerCmd(args[1:])
	case "git":
		gitCmd(args[1:])
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

func vsixCmd(args []string) {
	cmd := flag.NewFlagSet("vsix", flag.ExitOnError)
	fileName := cmd.String("file", "", "Path to the list of packages to download.")
	helpFlag := cmd.Bool("help", false, "Print help and exit.")
	err := cmd.Parse(args)
	if err != nil || *helpFlag || fileName == nil || *fileName == "" {
		cmd.PrintDefaults()
		return
	}
	err = vsix.Run(vsix.Arguments{
		FileName: *fileName,
	})
	if err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}
}

func containerCmd(args []string) {
	cmd := flag.NewFlagSet("container", flag.ExitOnError)
	fileName := cmd.String("file", "", "Path to the list of containers to download.")
	helpFlag := cmd.Bool("help", false, "Print help and exit.")
	err := cmd.Parse(args)
	if err != nil || *helpFlag || fileName == nil || *fileName == "" {
		cmd.PrintDefaults()
		return
	}
	err = container.Run(container.Arguments{
		FileName: *fileName,
	})
	if err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}
}

func gitCmd(args []string) {
	cmd := flag.NewFlagSet("git", flag.ExitOnError)
	fileName := cmd.String("file", "", "Path to the list of git repositories to download.")
	accessToken := cmd.String("accessToken", "", "Github access token, or password.")
	helpFlag := cmd.Bool("help", false, "Print help and exit.")
	err := cmd.Parse(args)
	if err != nil || *helpFlag || fileName == nil || *fileName == "" || accessToken == nil || *accessToken == "" {
		cmd.PrintDefaults()
		return
	}
	err = git.Run(git.Arguments{
		FileName:    *fileName,
		AccessToken: *accessToken,
	})
	if err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}
}
