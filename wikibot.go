package main

import (
	"encoding/csv"
	"fmt"
	"log"
	"os"

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

	pages := []string{}
	pages = append(pages, "User: Edmond Dantès", "Deadbeats/pt-br")
	coolPages := bot.GetArticles(pages)
	fmt.Println(coolPages)
}
