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
