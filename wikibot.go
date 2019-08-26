package main

import (
	"encoding/csv"
	"io/ioutil"
	"log"
	"os"
	"strings"
	"time"

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

	// bot.CompareTranslations("Deadbeats/pt-br")

	os.Create("categories.txt")
	os.Create("pages.txt")

	for {
		categoriesFile, err := ioutil.ReadFile("categories.txt")
		if err != nil {
			log.Fatal(err)
		}
		categories := strings.Split(string(categoriesFile), "\n")

		for i := len(categories) - 1; i >= 0; i-- {
			category := categories[i]
			trim := strings.Trim(category, " ")
			if trim == "" {
				continue
			}
			bot.CompareLinks(trim)

			categories = categories[0:i]
			ioutil.WriteFile("categories.txt", []byte(strings.Join(categories, "\n")), 0644)
		}

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
			bot.CompareTranslations(trim)

			pages = pages[0:i]
			ioutil.WriteFile("pages.txt", []byte(strings.Join(pages, "\n")), 0644)
		}

		time.Sleep(time.Second * 30)
	}
}
