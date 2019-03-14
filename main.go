package main

import (
	"database/sql"
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"time"

	//postgresql driver
	_ "github.com/lib/pq"
)

//Person -
type Person struct {
	Name struct {
		First string `json:"first"`
		Last  string `json:"last"`
		Title string `json:"title"`
	} `json:"name"`
	Gender string `json:"gender"`
	DOB    struct {
		Age  int       `json:"age"`
		Date time.Time `json:"date"`
	} `json:"dob"`
}

type RandomPerson struct {
	Results []Person `json:"results"`
}

type PersonEntry struct {
	ID     int    `db:"id"`
	Person Person `db:"person"`
}

type GenderClause struct {
	Gender string `json:"gender"`
}

func testErr(err error) {
	if err != nil {
		log.Fatal(err)
	}
}

func main() {
	db, err := sql.Open("postgres", "postgres://localhost/ssimpson?sslmode=disable")
	testErr(err)
	defer db.Close()

	clauseJSON, err := json.Marshal(GenderClause{Gender: "female"})
	testErr(err)

	selectStmt := "select id, person from people where person @> $1"
	rows, err := db.Query(selectStmt, string(clauseJSON))
	testErr(err)

	for rows.Next() {
		var person PersonEntry
		rows.Scan(&person.ID, &person.Person)
		fmt.Println(person.Person)
	}
}

func SeedData() {
	db, err := sql.Open("postgres",
		"postgres://localhost/ssimpson?sslmode=disable")
	testErr(err)
	defer db.Close()

	resp, err := http.Get("https://randomuser.me/api?inc=gender,name,dob&results=100")
	testErr(err)
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	testErr(err)

	var results RandomPerson
	err = json.Unmarshal(body, &results)
	testErr(err)

	stmt, err := db.Prepare("insert into people (person) values ($1)")
	testErr(err)

	for _, result := range results.Results {
		_, err = stmt.Exec(&result)
		testErr(err)
	}
}

//Value of the person to the sql driver.
func (p Person) Value() (driver.Value, error) {
	js, err := json.Marshal(&p)
	if err != nil {
		return nil, err
	}
	return js, nil
}

//Scan should unmarshall the json field into the correct struct
func (p *Person) Scan(src interface{}) error {
	err := json.Unmarshal(src.([]byte), p)
	return err
}

func (p Person) String() string {
	return fmt.Sprintf("%s %s %s", p.Name.Title, p.Name.First, p.Name.Last)
}
