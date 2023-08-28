package database

import (
	"database/sql"

	_ "github.com/go-sql-driver/mysql"
)

var (
	MysqlDB *sql.DB
)

func InitMysql(dataSource string) error {
	db, err := sql.Open("mysql", dataSource)
	if err != nil {
		return err
	}
	MysqlDB = db
	return nil
}
