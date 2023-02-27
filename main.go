package main

import (
	"MailGatherBot/bot"
	"MailGatherBot/database"
	"log"
	"os"
)

func main() {
	// Create database
	databaseName := os.Getenv("DATABASE_NAME")
	if databaseName == "" {
		log.Fatalln("please provide DATABASE_NAME as environment variable")
	}
	db, err := database.NewDatabase(databaseName)
	if err != nil {
		log.Fatalf("cannot open database: %s\n", err)
	}
	// Create the bot
	b := bot.Bot{
		ApiToken: os.Getenv("API_TOKEN"),
		Database: db,
	}
	b.Start()
}
