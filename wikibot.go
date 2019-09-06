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

	argv := os.Args[1:]

	db, err := client.SetupDatabase(configCsv[2], configCsv[3], configCsv[4], configCsv[5], configCsv[6])
	if err != nil {
		log.Panicln(err)
	}
	defer db.Close()

	botUsername := configCsv[0]
	botPassword := configCsv[1]

	bot := client.Wiki(botUsername, botPassword)

	for _, page := range argv {
		bot.ProcessArticle(page)
	}

	if err != nil {
		log.Printf("[Main] Error writing wikilist.txt->\n\t%s\n", err)
	}

	firstLoop := true

	client.RenderPages()

	for {
		pagesFile, err := ioutil.ReadFile("queue.txt")
		if err != nil {
			log.Fatal(err)
		}
		pages := strings.Split(string(pagesFile), "\n")

		for i := len(pages) - 1; i >= 0; i-- {
			page := pages[i]
			trim := strings.Trim(page, " \r")
			if trim == "" {
				continue
			}
			log.Println("Processing " + trim)
			bot.ProcessArticle(trim)

			pages = pages[0:i]
			ioutil.WriteFile("queue.txt", []byte(strings.Join(pages, "\n")), 0644)
		}

		if firstLoop {
			db, err := client.GetDBEntries(true)
			if err != nil {
				log.Printf("Error getting outdated entries from the DB.")
			}

			for _, pages := range db {
				for _, page := range pages {
					bot.ProcessArticle(page.Title)
				}
			}

			firstLoop = false
		}
		time.Sleep(time.Second * 30)
	}
}
