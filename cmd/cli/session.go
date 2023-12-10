package main

import (
	"fmt"
	"time"
)

func doSessionTable() error {

	dbType := cel.DB.DataType

	if dbType == "mariadb" {
		dbType = "mysql"
	}
	if dbType == "postgresql" || dbType == "pgx" {
		dbType = "postgres"
	}
	fileName := fmt.Sprintf("%d_create_sesssions_table", time.Now().UnixMicro())

	upFile := cel.RootPath + "/migrations/" + fileName + "." + dbType + ".up.sql"
	downFile := cel.RootPath + "/migrations/" + fileName + "." + dbType + ".down.sql"

	if err := copyFilefromTemplate("templates/migrations/"+dbType+"/session.sql", upFile); err != nil {
		exitGracefully(err)
	}
	if err := copyDataToFile([]byte("drop table sessions"), downFile); err != nil {
		exitGracefully(err)
	}

	if err := doMigrate("up", ""); err != nil {
		exitGracefully(err)
	}
	return nil
}
