package main

import (
	"log"
	"os"
	"strings"

	"github.com/fatih/color"
	"github.com/go-git/go-git/v5"
)

func doNew(appName string) {
	appName = strings.ToLower(appName)

	//sanitize application name
	if strings.Contains(appName, "/") {
		exploded := strings.SplitAfter(appName, "/")
		appName = exploded[(len(exploded)-1)]
	}

	log.Println("App name: ", appName)


	//clone skeleton files
	color.Green("\tCloning repository...")
	if _, err := git.PlainClone("./"+appName,false, &git.CloneOptions{
		Progress: os.Stdout,
		Depth: 1,

	}); err != nil {
		exitGracefully(err)
	}
	

}
