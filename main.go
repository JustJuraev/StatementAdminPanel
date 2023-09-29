package main

import (
	"crypto/md5"
	"database/sql"
	"encoding/hex"
	"fmt"
	"html/template"
	"net/http"
	"time"

	"github.com/gorilla/mux"
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
	UserId         int
	OrgId          int
}

type User struct {
	Id       int    `json:"id"`
	Login    string `json:"login"`
	Password string `json:"password"`
	Token    string `json:"token"`
	Name     string `json:"name"`
	LastName string `json:"lastname"`
	Role     int    `json:"role"`
	OrgId    int    `json:"orgid"`
}

type StatementHistory struct {
	Id          int
	UserId      int
	StatementId int
	Time        time.Time
	Action      string
	OrgId       int
}

type StH struct {
	UserId      int
	StatementId int
	Time        time.Time
	Action      string
	OrgId       int
	Name        string
	LastName    string
}

type Organization struct {
	Id   int
	Name string
}

var statements = []StatementStruct{}
var organizations = []Organization{}
var sthistory = []StH{}

func GetStatements(page http.ResponseWriter, r *http.Request) {
	connStr := "user=postgres password=123456 dbname=mygovdb sslmode=disable"
	db, err := sql.Open("postgres", connStr)

	if err != nil {
		panic(err)
	}

	defer db.Close()

	row, err2 := db.Query("SELECT * FROM public.statements WHERE status=$1", 100)

	if err2 != nil {
		panic(err2)
	}

	defer row.Close()

	statements := []StatementStruct{}
	for row.Next() {
		st := StatementStruct{}
		err3 := row.Scan(&st.Id, &st.Name, &st.LastName, &st.Date, &st.Status, &st.Statement, &st.PassportSeries, &st.Time, &st.UserId, &st.OrgId)
		if err3 != nil {
			fmt.Println(err3)
		}
		statements = append(statements, st)
	}

	tmpl, err := template.ParseFiles("html_files/statements.html", "html_files/zagolovok.html")
	if err != nil {
		panic(err)
	}
	tmpl.ExecuteTemplate(page, "statements", statements)
}

func OrgStatements(page http.ResponseWriter, r *http.Request) {
	connStr := "user=postgres password=123456 dbname=mygovdb sslmode=disable"
	db, err := sql.Open("postgres", connStr)

	if err != nil {
		panic(err)
	}

	defer db.Close()

	row, err2 := db.Query("SELECT * FROM public.statements WHERE status=$1", 110)

	if err2 != nil {
		panic(err2)
	}

	defer row.Close()

	statements := []StatementStruct{}
	for row.Next() {
		st := StatementStruct{}
		err3 := row.Scan(&st.Id, &st.Name, &st.LastName, &st.Date, &st.Status, &st.Statement, &st.PassportSeries, &st.Time, &st.UserId, &st.OrgId)
		if err3 != nil {
			fmt.Println(err3)
		}
		statements = append(statements, st)
	}

	tmpl, err := template.ParseFiles("html_files/orgstatements.html", "html_files/zagolovok.html")
	if err != nil {
		panic(err)
	}
	tmpl.ExecuteTemplate(page, "orgstatements", statements)
}

func GetStatementText(page http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	connStr := "user=postgres password=123456 dbname=mygovdb sslmode=disable"
	db, err := sql.Open("postgres", connStr)

	if err != nil {
		panic(err)
	}

	defer db.Close()

	row := db.QueryRow("SELECT * FROM public.statements WHERE id = $1", id)
	st := StatementStruct{}
	err2 := row.Scan(&st.Id, &st.Name, &st.LastName, &st.Date, &st.Status, &st.Statement, &st.PassportSeries, &st.Time, &st.UserId, &st.OrgId)
	if err2 != nil {
		fmt.Println(err2)
	}

	row2, err2 := db.Query("SELECT s.Id, s.userid, s.statementid, s.date, s.Action, s.orgid, u.name, u.lastname FROM public.statementshistory AS s INNER JOIN public.users AS u on s.userid = u.id WHERE statementid = $1", st.Id)

	if err2 != nil {
		panic(err2)
	}

	defer row2.Close()

	sthistory = []StH{}
	for row2.Next() {
		stt := StatementHistory{}
		ur := User{}
		err3 := row2.Scan(&stt.Id, &stt.UserId, &stt.StatementId, &stt.Time, &stt.Action, &stt.OrgId, &ur.Name, &ur.LastName)
		if err3 != nil {
			fmt.Println(err3)
		}

		sth := StH{
			UserId:      stt.UserId,
			StatementId: stt.StatementId,
			Time:        stt.Time,
			Action:      stt.Action,
			OrgId:       stt.OrgId,
			Name:        ur.Name,
			LastName:    ur.LastName,
		}
		sthistory = append(sthistory, sth)

	}

	data := struct {
		St    StatementStruct
		Array []StH
	}{
		St:    st,
		Array: sthistory,
	}

	tmpl, err := template.ParseFiles("html_files/getsttext.html", "html_files/zagolovok.html")
	if err != nil {
		panic(err)
	}
	tmpl.ExecuteTemplate(page, "sttext", data)
}

func Login(page http.ResponseWriter, r *http.Request) {
	tmpl, err := template.ParseFiles("html_files/login.html")
	if err != nil {
		panic(err)
	}
	tmpl.ExecuteTemplate(page, "login", nil)
}

func LoginPost(page http.ResponseWriter, r *http.Request) {
	login := r.FormValue("login")
	password := r.FormValue("password")

	if login == "" || password == "" {
		tmpl, err := template.ParseFiles("html_files/login.html")
		if err != nil {
			panic(err)
		}

		tmpl.ExecuteTemplate(page, "login", "Все поля должны быть заполнеными")
	}
	connStr := "user=postgres password=123456 dbname=mygovdb sslmode=disable"
	db, err := sql.Open("postgres", connStr)

	if err != nil {
		panic(err)
	}

	defer db.Close()

	hash := md5.Sum([]byte(password))
	hashedPass := hex.EncodeToString(hash[:])

	res := db.QueryRow("SELECT * FROM public.users WHERE login = $1 AND password = $2", login, hashedPass)
	user := User{}

	err3 := res.Scan(&user.Id, &user.Login, &user.Password, &user.Name, &user.LastName, &user.Role, &user.OrgId)

	if err3 != nil {

		if user.OrgId == 0 && user.Role == 1 {
			http.Redirect(page, r, "/statements", http.StatusSeeOther)
			return
		}
		tmpl, err2 := template.ParseFiles("html_files/login.html")
		if err2 != nil {
			panic(err2)
		}

		tmpl.ExecuteTemplate(page, "login", "Неправильный логин или пароль")

	} else {

		if user.Role == 1 {
			http.Redirect(page, r, "/statements", http.StatusSeeOther)
		}
	}
}

func AddOrg(page http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	connStr := "user=postgres password=123456 dbname=mygovdb sslmode=disable"
	db, err := sql.Open("postgres", connStr)

	if err != nil {
		panic(err)
	}

	defer db.Close()

	row := db.QueryRow("SELECT * FROM public.statements WHERE id = $1", id)
	st := StatementStruct{}

	err2 := row.Scan(&st.Id, &st.Name, &st.LastName, &st.Date, &st.Status, &st.Statement, &st.PassportSeries, &st.Time, &st.UserId, &st.OrgId)
	if err2 != nil {
		fmt.Println(err2)
	}

	row2, err3 := db.Query("SELECT * FROM public.organizations")

	if err3 != nil {
		panic(err3)
	}

	defer row2.Close()

	organizations = []Organization{}
	for row2.Next() {
		org := Organization{}
		err4 := row2.Scan(&org.Id, &org.Name)
		if err4 != nil {
			panic(err4)
		}
		organizations = append(organizations, org)
	}

	data := struct {
		Array []Organization
		St    StatementStruct
	}{
		Array: organizations,
		St:    st,
	}

	tmpl, err := template.ParseFiles("html_files/addorg.html", "html_files/zagolovok.html")
	if err != nil {
		panic(err)
	}
	tmpl.ExecuteTemplate(page, "addorg", data)
}

func AddOrgPost(page http.ResponseWriter, r *http.Request) {
	sid := r.FormValue("statementid")
	uid := r.FormValue("userid")
	orgid := r.FormValue("orgid")
	name := r.FormValue("name")
	lastname := r.FormValue("lastname")
	date := r.FormValue("date")
	status := r.FormValue("status")
	statement := r.FormValue("statement")
	ps := r.FormValue("ps")
	time := time.Now()

	connStr := "user=postgres password=123456 dbname=mygovdb sslmode=disable"
	db, err := sql.Open("postgres", connStr)

	if err != nil {
		panic(err)
	}

	defer db.Close()

	_, err2 := db.Exec("UPDATE public.statements set name=$1, lastname=$2, date=$3, status=$4, statement=$5, passportseries=$6, userid=$7, orgid=$8 where id=$9", name, lastname, date, status, statement, ps, uid, orgid, sid)
	if err2 != nil {
		panic(err2)
	}

	_, err4 := db.Exec("INSERT INTO public.statementshistory (userid, statementid, date, action, orgid) VALUES ($1, $2, $3, $4, $5)", uid, sid, time, "sent to organization", orgid)
	if err4 != nil {
		panic(err2)
	}
	http.Redirect(page, r, "/statements", http.StatusSeeOther)
}

func RejectStatement(page http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	connStr := "user=postgres password=123456 dbname=mygovdb sslmode=disable"
	db, err := sql.Open("postgres", connStr)

	if err != nil {
		panic(err)
	}

	defer db.Close()

	row := db.QueryRow("SELECT * FROM public.statements WHERE id = $1", id)
	st := StatementStruct{}

	err2 := row.Scan(&st.Id, &st.Name, &st.LastName, &st.Date, &st.Status, &st.Statement, &st.PassportSeries, &st.Time, &st.UserId, &st.OrgId)
	if err2 != nil {
		fmt.Println(err2)
	}

	_, err3 := db.Exec("UPDATE public.statements set name=$1, lastname=$2, date=$3, status=$4, statement=$5, passportseries=$6, userid=$7, orgid=$8 where id=$9", st.Name, st.LastName, st.Date, 120, st.Statement, st.PassportSeries, st.UserId, st.OrgId, st.Id)
	if err3 != nil {
		panic(err3)
	}

	time := time.Now()
	_, err4 := db.Exec("INSERT INTO public.statementshistory (userid, statementid, date, action, orgid) VALUES ($1, $2, $3, $4, $5)", st.UserId, st.Id, time, "reject", st.OrgId)
	if err4 != nil {
		panic(err2)
	}
	http.Redirect(page, r, "/statements", http.StatusSeeOther)
}

func RejStatements(page http.ResponseWriter, r *http.Request) {
	connStr := "user=postgres password=123456 dbname=mygovdb sslmode=disable"
	db, err := sql.Open("postgres", connStr)

	if err != nil {
		panic(err)
	}

	defer db.Close()

	row, err2 := db.Query("SELECT * FROM public.statements WHERE status=$1", 120)

	if err2 != nil {
		panic(err2)
	}

	defer row.Close()

	statements := []StatementStruct{}
	for row.Next() {
		st := StatementStruct{}
		err3 := row.Scan(&st.Id, &st.Name, &st.LastName, &st.Date, &st.Status, &st.Statement, &st.PassportSeries, &st.Time, &st.UserId, &st.OrgId)
		if err3 != nil {
			fmt.Println(err3)
		}
		statements = append(statements, st)
	}

	tmpl, err := template.ParseFiles("html_files/rejstatements.html", "html_files/zagolovok.html")
	if err != nil {
		panic(err)
	}
	tmpl.ExecuteTemplate(page, "rejstatements", statements)
}

func CloseStatements(page http.ResponseWriter, r *http.Request) {
	connStr := "user=postgres password=123456 dbname=mygovdb sslmode=disable"
	db, err := sql.Open("postgres", connStr)

	if err != nil {
		panic(err)
	}

	defer db.Close()

	row, err2 := db.Query("SELECT * FROM public.statements WHERE status=$1", 200)

	if err2 != nil {
		panic(err2)
	}

	defer row.Close()

	statements := []StatementStruct{}
	for row.Next() {
		st := StatementStruct{}
		err3 := row.Scan(&st.Id, &st.Name, &st.LastName, &st.Date, &st.Status, &st.Statement, &st.PassportSeries, &st.Time, &st.UserId, &st.OrgId)
		if err3 != nil {
			fmt.Println(err3)
		}
		statements = append(statements, st)
	}

	tmpl, err := template.ParseFiles("html_files/closestatements.html", "html_files/zagolovok.html")
	if err != nil {
		panic(err)
	}
	tmpl.ExecuteTemplate(page, "closestatements", statements)
}

func main() {
	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("./static/"))))

	router := mux.NewRouter()
	router.HandleFunc("/", Login)
	router.HandleFunc("/login_check", LoginPost)
	router.HandleFunc("/statements", GetStatements)
	router.HandleFunc("/addorgpost", AddOrgPost)
	router.HandleFunc("/orgstatements", OrgStatements)
	router.HandleFunc("/rejstatements", RejStatements)
	router.HandleFunc("/closetatements", CloseStatements)
	router.HandleFunc("/getsttext/{id:[0-9]+}", GetStatementText)
	router.HandleFunc("/addorg/{id:[0-9]+}", AddOrg)
	router.HandleFunc("/rejectst/{id:[0-9]+}", RejectStatement)
	http.Handle("/", router)
	http.ListenAndServe(":8082", nil)
}
