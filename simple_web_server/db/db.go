package db

import (
	"database/sql"
	"fmt"
	"log"

	"../configuration"

	_ "github.com/go-sql-driver/mysql"
)

type DB struct {
	*sql.DB
}

const dsnTemplate = "%v:%v@tcp(%v:%v)/%v?charset=utf8"

func Init(config *configuration.Config) (*DB, error) {
	dsn := fmt.Sprintf(dsnTemplate,
		config.DBUser, config.DBPass, config.DBHost, config.DBPort, config.DBName,
	)

	log.Println("Connecting to db", dsn)
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		return nil, err
	}
	err = db.Ping()
	if err != nil {
		return nil, err
	}

	return &DB{db}, nil
}
