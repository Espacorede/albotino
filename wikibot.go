package main

import (
	"io/ioutil"
	"log"
	"os"
	"strings"
	"time"

	"./client"
)

func main() {
	configCsv, err := client.ReadCsv("config.csv")
	if err != nil {
		log.Panicln(err)
	}

	checkDescriptions := false

	argv := os.Args[1:]

	if argv[0] == "-d" {
		checkDescriptions = true
		argv = argv[1:]
	}

	db, err := client.SetupDatabase(configCsv[2], configCsv[3], configCsv[4], configCsv[5], configCsv[6])
	if err != nil {
		log.Panicln(err)
	}
	defer db.Close()

	botUsername := configCsv[0]
	botPassword := configCsv[1]

	bot := client.Wiki(botUsername, botPassword)

	for _, page := range argv {
		bot.ProcessArticle(page, checkDescriptions)
	}

	log.Print(client.RenderPage())

	firstLoop := true

	for {
		pagesFile, err := ioutil.ReadFile("queue.txt")
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
			bot.ProcessArticle(trim, checkDescriptions)

			pages = pages[0:i]
			ioutil.WriteFile("queue.txt", []byte(strings.Join(pages, "\n")), 0644)
		}

		if firstLoop {
			db, err := client.GetDBEntries(true)
			if err != nil {
				log.Printf("Error getting outdated entries from the DB.")
			}

			for _, page := range db {
				bot.ProcessArticle(page.Title, checkDescriptions)
			}

			firstLoop = false
		}
		time.Sleep(time.Second * 30)
	}
}
