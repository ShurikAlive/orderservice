package transport

import (
	"database/sql"
	_ "github.com/go-sql-driver/mysql"
	log "github.com/sirupsen/logrus"
	"sync"
)

var instance *sql.DB
var once sync.Once

func ConnectDB() *sql.DB {

	//a, err := config.parseEnv()
	db, err := sql.Open("mysql", "root:Future1994!)@/cafe_test")//MySQL80
	if err != nil {
		log.Fatal(err)
		 db = nil
	}
	//defer db.Close()

	if err := db.Ping(); err != nil {
		log.Fatal(err)
		db = nil
	}

	return db
}

func GetDBInstance() *sql.DB {
	once.Do(func() {
		instance = ConnectDB()
	})
	return instance
}
