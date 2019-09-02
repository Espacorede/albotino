package main

import (
	"io/ioutil"
	"log"
	"strings"
	"time"

	"./client"
)

func main() {
	configCsv, err := client.ReadCsv("config.csv")
	if err != nil {
		log.Panicln(err)
	}

	err = client.SetupDatabase(configCsv[2], configCsv[3], configCsv[4], configCsv[5], configCsv[6])
	if err != nil {
		log.Panicln(err)
	}
	defer client.Database.Close()

	botUsername := configCsv[0]
	botPassword := configCsv[1]

	bot := client.Wiki(botUsername, botPassword)

	bot.ProcessArticle("Deadbeats/pt-br", true)

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
			bot.ProcessArticle(trim, true)

			pages = pages[0:i]
			ioutil.WriteFile("queue.txt", []byte(strings.Join(pages, "\n")), 0644)
		}
		time.Sleep(time.Second * 30)
	}
}
