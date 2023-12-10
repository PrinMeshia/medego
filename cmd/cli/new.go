package main

import (
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/fatih/color"
	"github.com/go-git/go-git/v5"
)

const (
	skeletonRepoURL = "https://github.com/PrinMeshia/medego-skeleton"
)

var templateFiles = map[string]string{
	"env":    "templates/env.txt",
	"go.mod": "templates/go.mod.txt",
}

var appURL string

func copyMakefile(appName, sourceFile string) {
	source, err := os.Open(sourceFile)
	checkError(err, "Opening source file")
	defer source.Close()

	dest, err := os.Create(filepath.Join(".", appName, "Makefile"))
	checkError(err, "Creating destination file")
	defer dest.Close()

	_, err = io.Copy(dest, source)
	checkError(err, "Copying content")
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
	if _, err := git.PlainClone(filepath.Join(".", appName), false, &git.CloneOptions{
		URL:      skeletonRepoURL,
		Progress: os.Stdout,
		Depth:    1,
	}); err != nil {
		exitGracefully(err)
	}

	checkError(os.RemoveAll(filepath.Join(".", appName, ".git")), "Removing .git directory")

	//create .env file
	color.Yellow("\tCreating .env file...")
	envTemplate, err := templateFS.ReadFile(templateFiles["env"])
	checkError(err, "Reading .env template")

	env := string(envTemplate)
	env = strings.ReplaceAll(env, "${APP_NAME}", appName)
	env = strings.ReplaceAll(env, "${KEY}", core.RandomString(32))

	checkError(copyDataToFile([]byte(env), filepath.Join(".", appName, ".env")), "Creating .env file")

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
	checkError(os.Remove(filepath.Join(".", appName, "go.mod")), "Removing go.mod file")

	goModTemplate, err := templateFS.ReadFile(templateFiles["go.mod"])
	checkError(err, "Reading go.mod template")

	mod := string(goModTemplate)
	mod = strings.ReplaceAll(mod, "${APP_NAME}", appURL)

	checkError(copyDataToFile([]byte(mod), filepath.Join(".", appName, "go.mod")), "Creating go.mod file")

	// update existing .go files with correct name/imports
	color.Yellow("\tUpdating source files...")
	os.Chdir("./" + appName)
	updateSource()

	// run go mod tidy in the project directory
	color.Yellow("\tRunning go mod tidy...")
	cmd := exec.Command("go", "mod", "tidy")
	checkError(cmd.Start(), "Starting go mod tidy")
	checkError(cmd.Wait(), "Waiting for go mod tidy to complete")
	color.Green("Done building " + appURL)
	color.Green("Go build something awesome")

}
