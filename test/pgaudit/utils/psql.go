/*
Copyright 2023 Codenotary Inc. All rights reserved.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

	http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package main

import (
	"database/sql"
	"fmt"
	"log"
	"time"

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

	for i := 0; i < 100000; i++ {
		log.Printf("QUERY: %s\n", fmt.Sprintf("insert into audit_trail(id, ts, usr, action, sourceip, context) VALUES ('%s', NOW(), '%s', 1, '127.0.0.1', 'some context')",
			uuid.New().String(), "user"+fmt.Sprint(i)))
		_, err := db.Exec(fmt.Sprintf("insert into audit_trail(id, ts, usr, action, sourceip, context) VALUES ('%s', NOW(), '%s', 1, '127.0.0.1', 'some context')",
			uuid.New().String(), fmt.Sprintf("user%d", i)))
		if err != nil {
			log.Fatal(err)
		}
		time.Sleep(5 * time.Second)
	}

}

func main() {
	PopulatePSQL()
}
