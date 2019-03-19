package main

import (
	"crypto/sha256"
	"database/sql"
	"fmt"
	"net/http"

	"github.com/go-chi/chi"
	_ "github.com/go-sql-driver/mysql"
)

var db *sql.DB

func init() {
	const (
		dbName = "db01"
		dbUser = "app01"
		dbPass = "app01"
		dbHost = "127.0.0.1"
		dbPort = "3306"
	)

	dbSource := fmt.Sprintf(
		"%s:%s@tcp(%s:%s)/%s?charset=utf8",
		dbUser,
		dbPass,
		dbHost,
		dbPort,
		dbName)

	var err error
	db, err = sql.Open("mysql", dbSource)
	if err != nil {
		panic(err)
	}
}

func validateUser(name string, pass string) bool {
	hash := fmt.Sprintf("%x", sha256.Sum256([]byte(name+pass)))
	rows, err := db.Query("select * from users where hash = '" + hash + "';")
	defer rows.Close()
	if err != nil {
		panic(err.Error())
	}

	columns, err := rows.Columns()
	if err != nil {
		panic(err.Error())
	}

	values := make([]sql.RawBytes, len(columns))
	scanArgs := make([]interface{}, len(values))
	for i := range values {
		scanArgs[i] = &values[i]
	}

	if rows.Next() {
		return true
	}
	return false
}

func main() {

	r := chi.NewRouter()

	r.Post("/auth", func(w http.ResponseWriter, r *http.Request) {
		r.ParseForm()

		q := r.URL.Query()
		if q == nil {
			w.Write([]byte("parameter is nothing...\n"))
			return
		}

		name, password := "", ""
		for k, v := range q {
			if k == "name" {
				name = v[0]
			} else if k == "password" {
				password = v[0]
			}
		}

		if name == "" || password == "" {
			w.Write([]byte("either name or password parameter is missing...\n"))
			return
		}

		if validateUser(name, password) {
			w.Write([]byte("OK\n"))
			return
		}

		w.Write([]byte("NG\n"))
		return
	})

	r.Post("/", func(w http.ResponseWriter, r *http.Request) {
		r.ParseForm()

		q := r.URL.Query()
		if q == nil {
			w.Write([]byte("parameter is nothing...\n"))
			return
		}

		name, password := "", ""
		for k, v := range q {
			if k == "name" {
				name = v[0]
			} else if k == "password" {
				password = v[0]
			}
		}

		if name == "" || password == "" {
			w.Write([]byte("either name or password parameter is missing...\n"))
			return
		}

		sql := "insert into users(name, password) values('" + name + "','" + password + "');"
		_, err := db.Query(sql)
		if err != nil {
			w.Write([]byte("internal db error happened...\n"))
			return
		}

		w.Write([]byte("register success.\n"))
		return
	})

	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		rows, err := db.Query("select * from users")
		defer rows.Close()
		if err != nil {
			panic(err.Error())
		}

		columns, err := rows.Columns()
		if err != nil {
			panic(err.Error())
		}

		values := make([]sql.RawBytes, len(columns))
		scanArgs := make([]interface{}, len(values))
		for i := range values {
			scanArgs[i] = &values[i]
		}

		for rows.Next() {
			err = rows.Scan(scanArgs...)
			if err != nil {
				panic(err.Error())
			}

			var value string
			for i, col := range values {
				if col == nil {
					value = "NULL"
				} else {
					value = string(col)
				}
				w.Write([]byte(columns[i] + ": " + value + "\n"))
			}
			w.Write([]byte("-----------------------------------\n"))
		}
	})

	http.ListenAndServe(":9999", r)

}
