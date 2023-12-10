package main

import (
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"runtime"
	"strings"

	"github.com/fatih/color"
	"github.com/go-git/go-git/v5"
)

var appURL string

func copyMakefile(appName, sourceFile string) {
	source, err := os.Open(sourceFile)
	if err != nil {
		exitGracefully(err)
	}
	defer source.Close()

	dest, err := os.Create(fmt.Sprintf("./%s/Makefile", appName))
	if err != nil {
		exitGracefully(err)
	}
	defer dest.Close()

	_, err = io.Copy(dest, source)
	if err != nil {
		exitGracefully(err)
	}
}

func doNew(appName string) {
	appName = strings.ToLower(appName)
	appURL = appName

	//sanitize application name
	if strings.Contains(appName, "/") {
		exploded := strings.SplitAfter(appName, "/")
		appName = exploded[(len(exploded) - 1)]
	}

	log.Println("App name: ", appName)

	//clone skeleton files
	color.Green("\tCloning repository...")
	if _, err := git.PlainClone("./"+appName, false, &git.CloneOptions{
		URL:      "https://github.com/PrinMeshia/medego-skeleton",
		Progress: os.Stdout,
		Depth:    1,
	}); err != nil {
		exitGracefully(err)
	}

	if err := os.RemoveAll(fmt.Sprintf("./%s/.git", appName)); err != nil {
		exitGracefully(err)
	}

	//create .env file
	color.Yellow("\tCreating .env file...")
	data, err := templateFS.ReadFile("templates/env.txt")
	if err != nil {
		exitGracefully(err)
	}

	env := string(data)
	env = strings.ReplaceAll(env, "${APP_NAME}", appName)
	env = strings.ReplaceAll(env, "${KEY}", core.RandomString(32))

	err = copyDataToFile([]byte(env), fmt.Sprintf("./%s/.env", appName))
	if err != nil {
		exitGracefully(err)
	}

	//Makefile
	if runtime.GOOS == "windows" {
		copyMakefile(appName, fmt.Sprintf("./%s/MakefileWin", appName))
	} else {
		copyMakefile(appName, fmt.Sprintf("./%s/MakefileUnix", appName))
	}
	_ = os.Remove("./" + appName + "/MakefileWin")
	_ = os.Remove("./" + appName + "/MakefileUnix")

	//update go.mod
	color.Yellow("\tCreating go.mod file...")
	_ = os.Remove("./" + appName + "/go.mod")

	data, err = templateFS.ReadFile("templates/go.mod.txt")
	if err != nil {
		exitGracefully(err)
	}

	mod := string(data)
	mod = strings.ReplaceAll(mod, "${APP_NAME}", appURL)

	err = copyDataToFile([]byte(mod), "./"+appName+"/go.mod")
	if err != nil {
		exitGracefully(err)
	}

	// update existing .go files with correct name/imports
	color.Yellow("\tUpdating source files...")
	os.Chdir("./" + appName)
	updateSource()

	// run go mod tidy in the project directory
	color.Yellow("\tRunning go mod tidy...")
	cmd := exec.Command("go", "mod", "tidy")
	err = cmd.Start()
	if err != nil {
		exitGracefully(err)
	}

	color.Green("Done building " + appURL)
	color.Green("Go build something awesome")

}
