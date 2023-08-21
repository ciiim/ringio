package database

import (
	"database/sql"

	_ "github.com/go-sql-driver/mysql"
)

const (
	dataSource = ""
)

func NewMysql() *sql.DB {
	db, err := sql.Open("mysql", dataSource)
	if err != nil {
		return nil
	}
	return db
}
