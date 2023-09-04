package database

import (
	"fmt"
	"testing"

	_ "github.com/go-sql-driver/mysql"
)

type TestSql struct {
	Name string `db:"name"`
	Age  int    `db:"age"`
}

func TestSelectMysql(t *testing.T) {
	fmt.Printf("SelectStringMysql(&TestSql{}, \"name = ciiim\"): %v\n", SelectStringMysql(&TestSql{}, "name = ciiim"))
}
