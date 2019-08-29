package main

import (
	"database/sql"
	"encoding/csv"
	"io/ioutil"
	"log"
	"os"
	"strings"
	"time"

	_ "github.com/lib/pq"

	"./client"
)

func main() {
	cfg, err := os.Open("config.csv")
	if err != nil {
		log.Panicln(err)
	}
	csvreader := csv.NewReader(cfg)

	values, err := csvreader.Read()
	if err != nil {
		log.Panicln(err)
	}

	botUsername := values[0]
	botPassword := values[1]

	bot := client.Wiki(botUsername, botPassword)

	databaseData, err := ioutil.ReadFile("db.txt")
	if err != nil {
		log.Fatal("Error opening db.txt\n" + err)
	}

	database, databaseErr := sql.Open("postgres", databaseData)
	if databaseErr != nil {
		log.Fatal("Error opening database. " + databaseErr)
	}
	defer database.Close()

	err := database.Ping()
	if err != nil {
		log.Fatal("Error connecting to database. This is most likely a problem with db.txt. " + err)
	}

	statement, tableErr := database.Prepare("CREATE TABLE IF NOT EXISTS wikipages (title TEXT PRIMARY KEY, points FLOAT, lastseen DATE, brokenlinks TEXT, wronglinks TEXT)")
	statement.Exec()
	if tableErr != nil {
		log.Fatal("Error creating table. " + tableErr)
	}

	// bot.CompareTranslations("Deadbeats/pt-br")

	for {
		pagesFile, err := ioutil.ReadFile("pages.txt")
		if err != nil {
			log.Fatal(err)
		}
		pages := strings.Split(string(pagesFile), "\n")

		for i := len(pages) - 1; i >= 0; i-- {
			page := pages[i]
			trim := strings.Trim(page, " ")
			if trim == "" {
				continue
			}
			log.Println("Processing " + trim)
			bot.ProcessArticle(trim, true)

			pages = pages[0:i]
			ioutil.WriteFile("queue.txt", []byte(strings.Join(pages, "\n")), 0644)
		}
		time.Sleep(time.Second * 30)
	}
}
