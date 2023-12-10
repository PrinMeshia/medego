package main

import (
	"errors"
	"os"

	"github.com/PrinMeshia/medego"
	"github.com/fatih/color"
)

const version = "1.0.0"

var core medego.Medego

func main() {
	var message string
	arg1, arg2, arg3, err := validateInput()
	if err != nil {
		exitGracefully(err)
	}

	setup(arg1, arg2)

	switch arg1 {
	case "help":
		showHelp()
	case "new":
		if arg2 == "" {
			exitGracefully(errors.New("Application name required"))
		}
		doNew(arg2)
	case "version":
		color.Yellow("Application version:" + version)

	case "migrate":
		if arg2 == "" {
			arg2 = "up"
		}
		if err = doMigrate(arg2, arg3); err != nil {
			exitGracefully(err)
		}
		message = "Migrations completed successfully"

	case "make":
		if arg2 == "" {
			exitGracefully(errors.New("Make requires a subcommand: (migration|model|handler)"))
		}
		if err = doMake(arg2, arg3); err != nil {
			exitGracefully(err)
		}

	default:
		showHelp()
	}
	exitGracefully(nil, message)
}

func validateInput() (string, string, string, error) {
	if len(os.Args) < 2 {
		color.Red("Error: command required")
		showHelp()
		return "", "", "", errors.New("command required")
	}

	arg1 := os.Args[1]
	arg2, arg3 := "", ""

	if len(os.Args) >= 3 {
		arg2 = os.Args[2]
	}

	if len(os.Args) >= 4 {
		arg3 = os.Args[3]
	}

	return arg1, arg2, arg3, nil
}
