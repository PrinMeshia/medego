package main

import (
	"embed"
	"errors"
	"os"
)

//go:embed templates
var templateFS embed.FS

func copyFilefromTemplate(templatePath, targetFile string) error {
	//todo: check if file exists
	if fileExists(targetFile) {
		return errors.New(targetFile + " already exists")
	}

	data, err := templateFS.ReadFile(templatePath)
	if err != nil {
		exitGracefully(err)
	}

	if err = copyDataToFile(data, targetFile); err != nil {
		exitGracefully(err)
	}

	return nil
}

func copyDataToFile(data []byte, targetFile string) error {
	if err := os.WriteFile(targetFile, data, 0644); err != nil {
		return err
	}
	return nil
}

func fileExists(path string) bool {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return false
	}
	return true
}