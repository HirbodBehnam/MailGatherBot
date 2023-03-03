package bot

import (
	"MailGatherBot/database"
	"MailGatherBot/util"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"log"
	"sync/atomic"
	"time"
)

type Bot struct {
	ApiToken string
	Database database.Database
	bot      *tgbotapi.BotAPI
	// This channel must be closed when bot is done
	shutdownDone chan struct{}
	botStopped   atomic.Bool
}

func (b *Bot) Start() {
	// Make stuff ready for shutdown
	b.shutdownDone = make(chan struct{})
	defer close(b.shutdownDone)
	b.botStopped.Store(false)
	// Start the bot
	var err error
	b.bot, err = tgbotapi.NewBotAPI(b.ApiToken)
	if err != nil {
		log.Fatal("Cannot initialize the bot: ", err.Error())
	}
	log.Println("Bot authorized on account", b.bot.Self.UserName)
	u := tgbotapi.NewUpdate(0)
	u.Timeout = 30
	updates := b.bot.GetUpdatesChan(u)
	for update := range updates {
		// We need this because of this:
		// https://github.com/go-telegram-bot-api/telegram-bot-api/issues/451
		if b.botStopped.Load() {
			break
		}
		if update.CallbackQuery != nil {
			go b.answerCallbackQuery(update.CallbackQuery.From.ID, update.CallbackQuery.Data, update.CallbackQuery.ID, update.CallbackQuery.InlineMessageID)
			continue
		}
		if update.InlineQuery != nil {
			go b.handleCreateListQuery(update.InlineQuery.Query, update.InlineQuery.ID)
			continue
		}
		// No clue what the hell is this
		if update.Message == nil {
			continue
		}
		// Check for commands
		if update.Message.IsCommand() {
			switch update.Message.Command() {
			case "about":
				_, _ = b.bot.Send(tgbotapi.NewMessage(update.Message.Chat.ID, aboutMessage))
				continue
			case "help":
				_, _ = b.bot.Send(tgbotapi.NewMessage(update.Message.Chat.ID, helpMessage))
				continue
			}
		}
		// Mail change?
		go b.mailChangeRequest(update.Message.From.ID, update.Message.Text)
	}
}

// StopBot will gracefully stop the bot
func (b *Bot) StopBot() {
	b.botStopped.Store(true)
	b.bot.StopReceivingUpdates()
	<-b.shutdownDone        // wait for event loop to finish
	time.Sleep(time.Second) // wait for database operations to finish
}

// handleCreateListQuery will create a list for user
func (b *Bot) handleCreateListQuery(listName, queryID string) {
	// If list name is empty just ignore it
	if listName == "" {
		_, _ = b.bot.Send(tgbotapi.InlineConfig{
			InlineQueryID: queryID,
		})
	}
	// Create an ID for list
	listID := util.RandomID()
	// Create list in database
	err := b.Database.CreateList(listID, listName)
	if err != nil {
		log.Println("cannot create list:", err)
		return
	}
	// Send it to user
	item := tgbotapi.NewInlineQueryResultArticle(listID, "List of "+listName, "List of "+listName)
	item.ReplyMarkup = &tgbotapi.InlineKeyboardMarkup{InlineKeyboard: [][]tgbotapi.InlineKeyboardButton{
		{
			{
				Text:         messageButtonText,
				CallbackData: &listID,
			},
		},
	}}
	_, _ = b.bot.Send(tgbotapi.InlineConfig{
		InlineQueryID: queryID,
		Results: []interface{}{
			item,
		},
		CacheTime:  0,
		IsPersonal: false,
	})
}

func (b *Bot) answerCallbackQuery(userID int64, listID, queryID, messageID string) {
	// Add or remove user from list
	err := b.Database.ParticipateOrOptOut(userID, listID)
	if err != nil {
		if err == database.NoEmailRegisteredErr {
			_, _ = b.bot.Send(tgbotapi.CallbackConfig{
				CallbackQueryID: queryID,
				URL:             "t.me/" + b.bot.Self.UserName + "?start=email",
			})
		} else {
			// Big fuckup
			log.Printf("cannot process user button click: %s\n", err)
		}
		return
	}
	// Refresh the list
	b.UpdateList(listID, messageID)
	// Answer it
	_, _ = b.bot.Send(tgbotapi.CallbackConfig{
		CallbackQueryID: queryID,
	})
}
