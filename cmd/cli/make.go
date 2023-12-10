package main

import (
	"errors"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/fatih/color"
	"github.com/gertd/go-pluralize"
	"github.com/iancoleman/strcase"
)

func doMake(arg2, arg3 string) error {

	switch arg2 {
	case "key":
		rnd := cel.RandomString(32)
		color.Yellow("32 characters encryption key: %s", rnd)
	case "migration":
		dbType := cel.DB.DataType
		if arg3 == "" {
			exitGracefully(errors.New("you must give the migration a name"))
		}

		fileName := fmt.Sprintf("%d_%s", time.Now().UnixMicro(), arg3)

		upFile := cel.RootPath + "/migrations/" + fileName + "." + dbType + ".up.sql"
		downFile := cel.RootPath + "/migrations/" + fileName + "." + dbType + ".down.sql"

		if err := copyFilefromTemplate("templates/migrations/"+dbType+"/migration.up.sql", upFile); err != nil {
			exitGracefully(err)
		}

		if err := copyFilefromTemplate("templates/migrations/"+dbType+"/migration.down.sql", downFile); err != nil {
			exitGracefully(err)
		}
	case "auth":
		if err := doAuth(); err != nil {
			exitGracefully(err)
		}
	case "handler":
		if arg3 == "" {
			exitGracefully(errors.New("you must give the handler a name"))
		}
		fileName := cel.RootPath + "/src/handlers/" + strings.ToLower(arg3) + ".go"
		if fileExists(fileName) {
			exitGracefully(errors.New(fileName + " already exists!"))
		}

		data, err := templateFS.ReadFile("templates/handlers/handler.go.txt")
		if err != nil {
			exitGracefully(err)
		}
		handler := string(data)
		handler = strings.ReplaceAll(handler, "$HANDLERNAME$", strcase.ToCamel(arg3))

		if err = os.WriteFile(fileName, []byte(handler), 0644); err != nil {
			exitGracefully(err)
		}
	case "model":
		if arg3 == "" {
			exitGracefully(errors.New("you must give the model a name"))
		}

		data, err := templateFS.ReadFile("templates/data/model.go.txt")
		if err != nil {
			exitGracefully(err)
		}
		model := string(data)

		plur := pluralize.NewClient()

		var modelName = arg3
		var tableName = arg3

		if plur.IsPlural(arg3) {
			modelName = plur.Singular(arg3)
			tableName = strings.ToLower(tableName)
		} else {
			tableName = strings.ToLower(plur.Plural(arg3))
		}

		fileName := cel.RootPath + "/src/data/" + strings.ToLower(modelName) + ".go"
		if fileExists(fileName) {
			exitGracefully(errors.New(fileName + " already exists!"))
		}

		model = strings.ReplaceAll(model, "$MODELNAME$", strcase.ToCamel(modelName))
		model = strings.ReplaceAll(model, "$TABLENAME$", tableName)

		if err = os.WriteFile(fileName, []byte(model), 0644); err != nil {
			exitGracefully(err)
		}
	case "mail":
		if arg3 == "" {
			exitGracefully(errors.New("template name required"))
		}
		htmlMail := cel.RootPath + "/mail/" + strings.ToLower(arg3) + ".html.tmpl"
		plainMail := cel.RootPath + "/mail/" + strings.ToLower(arg3) + ".plain.tmpl"

		if err := copyFilefromTemplate("templates/mailer/mail.html.tmpl",htmlMail); err != nil {
			exitGracefully(err)
		}
		if err := copyFilefromTemplate("templates/mailer/mail.plain.tmpl",plainMail); err != nil {
			exitGracefully(err)
		}

	case "session":
		if err := doSessionTable(); err != nil {
			exitGracefully(err)
		}
	
	

	}

	return nil
}
