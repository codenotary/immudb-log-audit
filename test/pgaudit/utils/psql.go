package main

import (
	"database/sql"
	"fmt"
	"log"

	"github.com/google/uuid"
	_ "github.com/lib/pq"
)

func PopulatePSQL() {
	db, err := sql.Open("postgres", "host=localhost port=5432 user=postgres sslmode=disable")
	if err != nil {
		log.Fatal("could not connect", err)
	}

	defer db.Close()

	_, err = db.Exec("create table if not exists audit_trail (id VARCHAR, ts TIMESTAMP, usr VARCHAR, action INTEGER, sourceip VARCHAR, context VARCHAR, PRIMARY KEY(id));")
	if err != nil {
		log.Fatal("could not create table", err)
	}

	for i := 0; i < 3000; i++ {
		log.Printf("QUERY: %s\n", fmt.Sprintf("insert into audit_trail(id, ts, usr, action, sourceip, context) VALUES ('%s', NOW(), '%s', 1, '127.0.0.1', 'some context')",
			uuid.New().String(), "user"+fmt.Sprint(i)))
		_, err := db.Exec(fmt.Sprintf("insert into audit_trail(id, ts, usr, action, sourceip, context) VALUES ('%s', NOW(), '%s', 1, '127.0.0.1', 'some context')",
			uuid.New().String(), fmt.Sprintf("user%d", i)))
		if err != nil {
			log.Fatal(err)
		}
	}

}

func main() {
	PopulatePSQL()
}
