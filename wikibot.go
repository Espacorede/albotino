package main

import (
	"encoding/csv"
	"log"
	"net"
	"os"

	"./client"
)

const host = "localhost:3000"

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

	log.Println(bot.GetRedirects([]string{"\u706b\u7130\u5175", "\u30d1\u30a4\u30ed", "\u041f\u043e\u0434\u0440\u044b\u0432\u043d\u0438\u043a"}))

	listen, err := net.Listen("tcp", host)

	if err != nil {
		log.Panicf("Error starting up TCP server: %s", err)
	}
	defer listen.Close()
	log.Printf("Server listening at %s", host)

	for {
		connection, err := listen.Accept()
		if err != nil {
			log.Printf("Server error right here blud: %s", err)
		}

		go serverHandle(connection)
	}
}

func serverHandle(connection net.Conn) {
	buffer := make([]byte, 2048)
	requestLength, err := connection.Read(buffer)
	if err != nil {
		log.Panicf("Error handling request: %s", requestLength)
	}
	connection.Write([]byte("i gotchu fam thx bye"))
	connection.Close()
}
