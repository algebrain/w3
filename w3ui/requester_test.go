package w3ui

import (
	"database/sql"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/algebrain/w3/w3req"
	"github.com/algebrain/w3/w3sql"
	"gopkg.in/gorp.v1"

	_ "modernc.org/sqlite"
)

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

var allSQL = w3sql.NewSQLString(`
	select * from students
`)

var cond1 = `{
	"Search": {
		"Col":  "secondNameLen",
		"Val":  7,
		"Op":   "<=",
		"Type": "int"
	}
}`

var cond2 = `{
	"Search": {
		"Op": "OR",
		"Query": [
			{
				"Col":  "age",
				"Val":  20,
				"Op":   ">",
				"Type": "int"
			},
			{
				"Col":  "grade",
				"Val":  77,
				"Op":   "==",
				"Type": "int"
			}
		]
	}
}`

var compileMap = map[string]string{
	"id":            "studentID",
	"firstName":     "", //значит будет использовано то же самое название колонки
	"secondName":    "",
	"age":           "",
	"grade":         "score",
	"secondNameLen": "length(secondName)",
}

var errorCodes = map[string]int{
	"System error. Try again later.": 1,
	"Invalid parameters":             2,
}

var toLowerCols = []string{"firstName", "secondName"}

var requester1 = NewDataRequester[Student](allSQL, compileMap, toLowerCols, errorCodes)

func openStudents(t *testing.T) *gorp.DbMap {
	db, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatal(err)
	}
	_, err = db.Exec(initStudentsTable)
	if err != nil {
		t.Fatal(err)
	}

	dbmap := &gorp.DbMap{Db: db, Dialect: gorp.SqliteDialect{}}
	dbmap.AddTableWithName(Student{}, "students").SetKeys(true, "studentID")
	err = dbmap.CreateTablesIfNotExists()
	if err != nil {
		t.Fatal(err)
	}
	return dbmap
}

func TestRequesterSelect(t *testing.T) {
	db := openStudents(t)

	requester1.InitOnce(func() RequesterOptions[Student] {
		return RequesterOptions[Student]{
			Or:                  MustReadJSON(cond1),
			GetDatabaseProvider: func() w3req.Conn { return db },
			ErrorLog:            nil,
			FormatFields:        func(r []Student) {
				t.Log(">>>>>>>>>>>>>>>>>HERE")
				for i := 0; i < len(r); i++ {
					r[i].FirstName = strings.ToUpper(r[i].FirstName)
				}
			},
		}
	})

	var allStudents []Student
	_, err := db.Select(&allStudents, "select * from students")

	if err != nil {
		t.Fatal(err)
	}

	t.Log("ALL STUDENTS:", getJSON(allStudents))

	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(cond2))
	w := httptest.NewRecorder()
	handler := requester1.GetHttpRequestHandler(100)
	handler(w, req)

	var answer struct {
		Status  string    `json:"status"`
		Total   int       `json:"total"`
		Records []Student `json:"records"`
	}
	b, err := io.ReadAll(w.Result().Body)
	if err != nil {
		t.Fatal(err)
	}

	err = json.Unmarshal(b, &answer)
	if err != nil {
		t.Fatal(err)
	}

	t.Log("ANSWER:", getJSON(answer))

	if answer.Total != 3 {
		t.Fatal("total=3 expected, got", answer.Total)
	}
}
