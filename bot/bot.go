package bot

import (
	"MailGatherBot/database"
	"MailGatherBot/util"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"log"
)

type Bot struct {
	ApiToken string
	Database database.Database
	bot      *tgbotapi.BotAPI
}

func (b *Bot) Start() {
	var err error
	b.bot, err = tgbotapi.NewBotAPI(b.ApiToken)
	if err != nil {
		log.Fatal("Cannot initialize the bot: ", err.Error())
	}
	log.Println("Bot authorized on account", b.bot.Self.UserName)
	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60
	updates := b.bot.GetUpdatesChan(u)
	for update := range updates {
		if update.CallbackQuery != nil {
			// Add or remove user from list
			err = b.Database.ParticipateOrOptOut(update.CallbackQuery.From.ID, update.CallbackQuery.Data)
			if err != nil {
				if err == database.NoEmailRegisteredErr {
					_, _ = b.bot.Send(tgbotapi.CallbackConfig{
						CallbackQueryID: update.CallbackQuery.ID,
						URL:             "t.me/" + b.bot.Self.UserName + "?start=email",
					})
				} else {
					// Big fuckup
					log.Printf("cannot process user button click: %s\n", err)
				}
				continue
			}
			// Refresh the list
			b.UpdateList(update.CallbackQuery.Data, update.CallbackQuery.InlineMessageID)
			// Answer it
			_, _ = b.bot.Send(tgbotapi.CallbackConfig{
				CallbackQueryID: update.CallbackQuery.ID,
			})
			continue
		}
		if update.InlineQuery != nil {
			inlineBotID := util.RandomID()
			// Create list in database
			err = b.Database.CreateList(inlineBotID, update.InlineQuery.Query)
			if err != nil {
				log.Println("cannot create list:", err)
				continue
			}
			// Send it to user
			item := tgbotapi.NewInlineQueryResultArticle(inlineBotID, "List of "+update.InlineQuery.Query, "List of "+update.InlineQuery.Query)
			item.ReplyMarkup = &tgbotapi.InlineKeyboardMarkup{InlineKeyboard: [][]tgbotapi.InlineKeyboardButton{
				{
					{
						Text:         messageButtonText,
						CallbackData: &inlineBotID,
					},
				},
			}}
			_, _ = b.bot.Send(tgbotapi.InlineConfig{
				InlineQueryID: update.InlineQuery.ID,
				Results: []interface{}{
					item,
				},
				CacheTime:  0,
				IsPersonal: false,
			})
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
		b.mailChangeRequest(update.Message.From.ID, update.Message.Text)
	}
}
