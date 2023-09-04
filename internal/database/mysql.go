package database

import (
	"database/sql"
	"fmt"
	"reflect"
	"strings"

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
	err = db.Ping()
	if err != nil {
		return err
	}
	MysqlDB = db
	return nil
}

/*
传入带有sql标签的结构体，将会自动生成对应sql语句
*/
func SelectStringMysql(v any, where string) string {
	vr := reflect.ValueOf(v)
	if vr.Kind() == reflect.Ptr {
		vr = vr.Elem()
	}
	if vr.Kind() != reflect.Struct {
		return ""
	}
	vt := vr.Type()
	var sql string
	for i := 0; i < vt.NumField(); i++ {
		if tag := vt.Field(i).Tag.Get("db"); tag != "" {
			sql += tag + ","
		}
	}
	tableName := lowerTableName(vt.Name())
	sql = sql[:len(sql)-1]
	sql = "SELECT " + sql + " FROM " + tableName + " WHERE " + where
	return sql
}

/*
传入带有sql标签的结构体，将会自动生成对应sql语句
*/
func InsertStringMysql(v any) string {
	vr := reflect.ValueOf(v)
	if vr.Kind() == reflect.Ptr {
		vr = vr.Elem()
	}
	if vr.Kind() != reflect.Struct {
		return ""
	}
	vt := vr.Type()
	var sql string
	var values string
	for i := 0; i < vt.NumField(); i++ {
		if tag := vt.Field(i).Tag.Get("db"); tag != "" {
			sql += tag + ","
			values += "?,"
		}
	}
	tableName := lowerTableName(vt.Name())
	sql = sql[:len(sql)-1]
	values = values[:len(values)-1]
	sql = "INSERT INTO " + tableName + "(" + sql + ") VALUES(" + values + ")"
	return sql
}

/*
传入带有sql标签的结构体，将会自动生成对应sql语句
*/
func UpdateStringMysql(v any, where string) string {
	vr := reflect.ValueOf(v)
	if vr.Kind() == reflect.Ptr {
		vr = vr.Elem()
	}
	if vr.Kind() != reflect.Struct {
		return ""
	}
	vt := vr.Type()
	var sql string
	for i := 0; i < vt.NumField(); i++ {
		if tag := vt.Field(i).Tag.Get("db"); tag != "" {
			sql += tag + "=?,"
		}
	}
	tableName := lowerTableName(vt.Name())
	sql = sql[:len(sql)-1]
	sql = "UPDATE " + tableName + " SET " + sql + " WHERE " + where
	fmt.Println(sql)
	return sql
}

/*
传入唯一标识符，将会自动生成对应sql语句
*/
func DeleteStringMysql(table, unique string) string {
	if table == "" || unique == "" {
		return ""
	}
	return "DELETE FROM " + table + " WHERE " + unique + "=?"
}

func lowerTableName(name string) string {
	return strings.ToLower(name)
}
