package main

import (
	"os"
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"runtime/debug"
	"strings"

	"github.com/algebrain/w3/w3req"
	"github.com/algebrain/w3/w3sql"
	"github.com/algebrain/w3/w3ui"
	"gopkg.in/gorp.v1"

	_ "modernc.org/sqlite"
)

type testLogger struct{}

func (t testLogger) Print(s string) string {
	fmt.Fprintln(os.Stderr, "=====>>", s)
	return s
}

func (t testLogger) Printf(s string, arg any, args ...any) string {
	fmt.Fprintln(os.Stderr, append([]any{"=====>> "+s, arg}, args...)...)
	return s
}

var initStudentsTable = `
create table students (
	studentID integer primary key,
	firstName text,
	secondName text,
	age int,
	score int
);

insert into students (firstName, secondName, age, score)
values 
	('vanya', 'ivanov', 22, 99),
	('petya', 'petrov', 21, 88),
	('lena', 'lenina', 20, 77),
	('masha', 'marinina', 19, 66);
`

type Student struct {
	StudentID  int64  `db:"studentID"`
	FirstName  string `db:"firstName"`
	SecondName string `db:"secondName"`
	Age        int    `db:"age"`
	Score      int    `db:"score"`
}

var allSQL = w3sql.NewSQLString(`select * from students`)

var cond1 = `{
	"Search": {
		"Op": "OR",
		"Query": [
			{
				"Col":  "age",
				"Val":  20,
				"Op":   ">",
				"Type": "int"
			}
		]
	},
	"Sort": [
		{
			"Col": "grade",
			"Dir": "asc"
		},
		{
			"Col": "age",
			"Dir": "desc"
		}
	]
}`

var compileMap = map[string]string{
	"id":            "studentID",
	"firstName":     "", //значит будет использовано то же самое название колонки
	"secondName":    "",
	"age":           "",
	"grade":         "score",
	"secondNameLen": "length(secondName)",
}

var toLowerCols = []string{"firstName", "secondName"}

var requester1 = w3ui.NewDataRequester3[Student](
	allSQL,
	compileMap,
	toLowerCols,
	func() {
		if r := recover(); r != nil {
			fmt.Fprintln(os.Stderr, "=====PANIC:", r)
			debug.PrintStack()
		}
	},
)

func openStudents() (*gorp.DbMap, error) {
	db, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		return nil, err
	}
	_, err = db.Exec(initStudentsTable)
	if err != nil {
		return nil, err
	}

	dbmap := &gorp.DbMap{Db: db, Dialect: gorp.SqliteDialect{}}
	dbmap.AddTableWithName(Student{}, "students").SetKeys(true, "studentID")
	err = dbmap.CreateTablesIfNotExists()
	if err != nil {
		return nil, err
	}
	return dbmap, nil
}

func main() {
	db, err := openStudents()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(2)
	}

	requester1.InitOnce(func() w3ui.RequesterOptions[Student] {
		requester1.DumpRequests()
		return w3ui.RequesterOptions[Student]{
			GetDB:    func() w3req.DB { return db },
			FormatFields: func(r []Student) {
				for i := 0; i < len(r); i++ {
					r[i].FirstName = strings.ToUpper(r[i].FirstName)
				}
			},
		}
	})
	
	if len(os.Args) < 2 {
		fmt.Fprintln(os.Stderr, "no query")
		os.Exit(1)
	}
	
	arg := os.Args[1]
	fmt.Println("QUERY:", arg)

	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(arg))
	w := httptest.NewRecorder()
	handler := requester1.GetHttpRequestHandler(100, w3ui.MustReadJSON(cond1))
	handler(w, req)

	var answer struct {
		Status  string    `json:"status"`
		Total   int       `json:"total"`
		Records []Student `json:"records"`
	}
	b, err := io.ReadAll(w.Result().Body)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(3)
	}

	err = json.Unmarshal(b, &answer)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(4)
	}

	fmt.Println(w3ui.GetJSON(answer))
}
