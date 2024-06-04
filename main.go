package main

import (
	"bytes"
	"flag"
	"fmt"
	"os"

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

  impex npm export -lock-file=/package-lock.json
  impex vsix export -file=./vsix.txt
  impex container export -file=./containers.txt
  impex git export -file=./git.txt -accessToken=ghp_fdsfdsfd
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
		err = npmCmd(args[1:])
	case "vsix":
		err = vsixCmd(args[1:])
	case "container":
		err = containerCmd(args[1:])
	case "containers":
		err = containerCmd(args[1:])
	case "git":
		err = gitCmd(args[1:])
	default:
		fmt.Println(help)
	}
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func subCommand(inArgs []string) (cmd string, args []string) {
	if len(inArgs) == 0 {
		return "", inArgs
	}
	return inArgs[0], inArgs[1:]
}

func npmCmd(args []string) error {
	cmd, args := subCommand(args)
	switch cmd {
	case "export":
		return npmExportCmd(args)
	case "import":
		return fmt.Errorf("not yet implemented")
	default:
		return fmt.Errorf("impex npm subcommand missing, expected export or import")
	}
}

func ErrInvalidArgs(cmd *flag.FlagSet) error {
	b := new(bytes.Buffer)
	cmd.SetOutput(b)
	cmd.PrintDefaults()
	return fmt.Errorf(b.String())
}

func npmExportCmd(args []string) error {
	cmd := flag.NewFlagSet("export", flag.ExitOnError)
	fileName := cmd.String("lock-file", "", "Path to the lock file.")
	helpFlag := cmd.Bool("help", false, "Print help and exit.")
	err := cmd.Parse(args)
	if err != nil || *helpFlag || fileName == nil || *fileName == "" {
		return ErrInvalidArgs(cmd)
	}
	return npm.Run(npm.Arguments{
		FileName: *fileName,
	})
}

func vsixCmd(args []string) error {
	cmd, args := subCommand(args)
	switch cmd {
	case "export":
		return vsixExportCmd(args)
	case "import":
		return fmt.Errorf("not yet implemented")
	default:
		return fmt.Errorf("impex vsix subcommand missing, expected export or import")
	}
}

func vsixExportCmd(args []string) error {
	cmd := flag.NewFlagSet("vsix", flag.ExitOnError)
	fileName := cmd.String("file", "", "Path to the list of packages to download.")
	helpFlag := cmd.Bool("help", false, "Print help and exit.")
	err := cmd.Parse(args)
	if err != nil || *helpFlag || fileName == nil || *fileName == "" {
		return ErrInvalidArgs(cmd)
	}
	return vsix.Run(vsix.Arguments{
		FileName: *fileName,
	})
}

func containerCmd(args []string) error {
	cmd, args := subCommand(args)
	switch cmd {
	case "export":
		return containerExportCmd(args)
	case "import":
		return fmt.Errorf("not yet implemented")
	default:
		return fmt.Errorf("impex container subcommand missing, expected export or import")
	}
}

func containerExportCmd(args []string) error {
	cmd := flag.NewFlagSet("container", flag.ExitOnError)
	fileName := cmd.String("file", "", "Path to the list of containers to download.")
	helpFlag := cmd.Bool("help", false, "Print help and exit.")
	err := cmd.Parse(args)
	if err != nil || *helpFlag || fileName == nil || *fileName == "" {
		return ErrInvalidArgs(cmd)
	}
	return container.Run(container.Arguments{
		FileName: *fileName,
	})
}

func gitCmd(args []string) error {
	cmd, args := subCommand(args)
	switch cmd {
	case "export":
		return gitExportCmd(args)
	case "import":
		return fmt.Errorf("not yet implemented")
	default:
		return fmt.Errorf("impex git subcommand missing, expected export or import")
	}
}

func gitExportCmd(args []string) error {
	cmd := flag.NewFlagSet("git", flag.ExitOnError)
	fileName := cmd.String("file", "", "Path to the list of git repositories to download.")
	accessToken := cmd.String("accessToken", "", "Github access token, or password.")
	helpFlag := cmd.Bool("help", false, "Print help and exit.")
	err := cmd.Parse(args)
	if err != nil || *helpFlag || fileName == nil || *fileName == "" || accessToken == nil || *accessToken == "" {
		return ErrInvalidArgs(cmd)
	}
	return git.Export(git.Arguments{
		FileName:    *fileName,
		AccessToken: *accessToken,
	})
}
