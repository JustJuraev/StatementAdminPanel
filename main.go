package main

import (
	"database/sql"
	"html/template"
	"net/http"
	"time"

	_ "github.com/lib/pq"
)

type StatementStruct struct {
	Id             int
	Name           string
	LastName       string
	Date           string
	Status         int
	Statement      string
	PassportSeries string
	Time           time.Time
}

var statements = []StatementStruct{}

func GetStatements(page http.ResponseWriter, r *http.Request) {
	connStr := "user=postgres password=123456 dbname=mygovdb sslmode=disable"
	db, err := sql.Open("postgres", connStr)

	if err != nil {
		panic(err)
	}

	defer db.Close()

	row, err2 := db.Query("SELECT * FROM public.statements")

	if err2 != nil {
		panic(err2)
	}

	defer row.Close()

	statements := []StatementStruct{}
	for row.Next() {
		st := StatementStruct{}
		err3 := row.Scan(&st.Id, &st.Name, &st.LastName, &st.Date, &st.Status, &st.Statement, &st.PassportSeries, &st.Time)
		if err3 != nil {
			panic(err3)
		}
		statements = append(statements, st)
	}

	tmpl, err := template.ParseFiles("html_files/statements.html", "html_files/zagolovok.html")
	if err != nil {
		panic(err)
	}
	tmpl.ExecuteTemplate(page, "statements", statements)
}

func main() {
	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("./static/"))))

	http.HandleFunc("/", GetStatements)
	http.ListenAndServe(":8082", nil)
}
