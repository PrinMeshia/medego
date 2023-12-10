package main

import (
	"fmt"
	"log"
	"time"

	"github.com/fatih/color"
)

func doAuth() error {
	//migrations
	dbType := cel.DB.DataType
	fileName := fmt.Sprintf("%d_create_auth_tables", time.Now().UnixMicro())
	upFile := cel.RootPath + "/migrations/" + fileName + ".up.sql"
	downFile := cel.RootPath + "/migrations/" + fileName + ".down.sql"

	log.Println(dbType, upFile, downFile)

	if err := copyFilefromTemplate("templates/migrations/"+dbType+"/auth_tables.sql", upFile); err != nil {
		exitGracefully(err)
	}
	if err := copyDataToFile([]byte("drop table if exists users cascade; drop table if exists tokens cascade; drop table if exists remember_tokens;"), downFile); err != nil {
		exitGracefully(err)
	}

	//run migrations
	if err := doMigrate("up", ""); err != nil {
		exitGracefully(err)
	}

	//copy files over
	if err := copyFilefromTemplate("templates/data/user.go.txt", cel.RootPath+"/src/data/user.go"); err != nil {
		exitGracefully(err)
	}

	if err := copyFilefromTemplate("templates/data/token.go.txt", cel.RootPath+"/src/data/token.go"); err != nil {
		exitGracefully(err)
	}
	if err := copyFilefromTemplate("templates/data/remember_token.go.txt", cel.RootPath+"/src/data/token.go"); err != nil {
		exitGracefully(err)
	}

	//copy middleware
	if err := copyFilefromTemplate("templates/middleware/auth.go.txt", cel.RootPath+"/src/middleware/auth.go"); err != nil {
		exitGracefully(err)
	}

	if err := copyFilefromTemplate("templates/middleware/auth-token.go.txt", cel.RootPath+"/src/middleware/auth-token.go"); err != nil {
		exitGracefully(err)
	}
	if err := copyFilefromTemplate("templates/middleware/remember.go.txt", cel.RootPath+"/src/middleware/auth-token.go"); err != nil {
		exitGracefully(err)
	}
	//copy handler
	if err := copyFilefromTemplate("templates/handlers/auth-handlers.go.txt", cel.RootPath+"/src/handlers/auth-handlers.go"); err != nil {
		exitGracefully(err)
	}

	//copy templates
	if err := copyFilefromTemplate("templates/views/login.jet", cel.RootPath+"/templates/auth/login.jet"); err != nil {
		exitGracefully(err)
	}
	if err := copyFilefromTemplate("templates/views/forgot.jet", cel.RootPath+"/templates/auth/forgot.jet"); err != nil {
		exitGracefully(err)
	}
	if err := copyFilefromTemplate("templates/views/reset-password.jet", cel.RootPath+"/templates/auth/reset-password.jet"); err != nil {
		exitGracefully(err)
	}

	//copy mailer
	if err := copyFilefromTemplate("templates/mailer/password-reset.html.tmpl", cel.RootPath+"/mail/password-reset.html.tmpl"); err != nil {
		exitGracefully(err)
	}
	if err := copyFilefromTemplate("templates/mailer/password-reset.plain.tmpl", cel.RootPath+"/mail/password-reset.plain.tmpl"); err != nil {
		exitGracefully(err)
	}

	color.Yellow(" - users, tokens, and remeber_tokens migrations created and executed")
	color.Yellow(" - user and token models created")
	color.Yellow(" - auth middleware created")
	color.Yellow("")
	color.Yellow("Don't forget to add user and token models in data/models.go")
	color.Yellow("And to add appropriate middleware to your routes!")

	return nil
}
