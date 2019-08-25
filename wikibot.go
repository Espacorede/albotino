package main

import (
	"encoding/csv"
	"log"
	"net/http"
	"os"

	"./client"
)

const host = ":3000"

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

	bot.CompareTranslations("Deadbeats/pt-br")

	fs := http.FileServer(http.Dir("static/"))

	http.Handle("/", http.StripPrefix("/static/", fs))

	log.Printf("Server listening at %s", host)

	log.Fatalln(http.ListenAndServe(host, nil))
}
