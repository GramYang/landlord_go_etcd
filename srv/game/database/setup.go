package database

import (
	"database/sql"
	_ "github.com/go-sql-driver/mysql"
	l "landlord_go/util/log"
)
var db *sql.DB
var log = l.L

func init() {
	var err error
	db, err = sql.Open("mysql", "gram:yangshu88@tcp(127.0.0.1:3306)/gram_landlord")
	if err != nil {
		log.Errorln(err)
	}
}