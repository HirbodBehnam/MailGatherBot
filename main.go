package main

import (
	"MailGatherBot/bot"
	"MailGatherBot/database"
	"log"
	"os"
	"os/signal"
	"syscall"
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
	apiToken := os.Getenv("API_TOKEN")
	if apiToken == "" {
		log.Fatalln("please provide API_TOKEN as environment variable")
	}
	b := bot.Bot{
		ApiToken: apiToken,
		Database: db,
	}
	go b.Start()
	// Wait for signal
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	<-sigs
	log.Println("Shutting down...")
	b.StopBot()
}
