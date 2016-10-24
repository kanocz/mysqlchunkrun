package main

import (
	"database/sql"
	"flag"
	"fmt"
	"log"
	"os"

	_ "github.com/go-sql-driver/mysql"
)

var (
	dbHost = flag.String("host", "localhost", "MySQL host to connect")
	dbPort = flag.Int("port", 3306, "MySQL port to connect")
	dbUser = flag.String("user", "root", "MySQL user to use")
	dbPass = flag.String("password", "", "MySQL password to use")
	dbDb   = flag.String("database", "", "MySQL database to select before run")
)

func main() {

	flag.Parse()

	database, err := sql.Open("mysql", fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8mb4&parseTime=true",
		*dbUser, *dbPass, *dbHost, *dbPort, *dbDb))

	if nil != err {
		log.Fatalln("Error connecting to database: ", err)
	}

	query := flag.Arg(0)
	if "" == query {
		fmt.Println("Please specify query as last argument")
		return
	}

	stmt, err := database.Prepare(query)
	if nil != err {
		log.Fatalln("Error preparing query: ", err)
	}

	for {
		res, err := stmt.Exec()
		if nil != err {
			log.Fatalln("Error executing query: ", err)
		}

		if rows, err := res.RowsAffected(); 0 == rows {
			if nil != err {
				log.Fatalln("Error on RowsAffected(): ", err)
			}
			fmt.Println("0 affected rows, probably done")
			break
		}

		os.Stdout.Write([]byte("."))
	}
}
