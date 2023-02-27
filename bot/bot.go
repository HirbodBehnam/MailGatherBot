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
			// TODO: for now we only have add to list
			_ = b.Database.AddToList(update.CallbackQuery.From.ID, update.CallbackQuery.Data)
			// Refresh the list
			b.UpdateList(update.CallbackQuery.Data)
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
			_, _ = b.bot.Send(tgbotapi.InlineConfig{
				InlineQueryID: update.InlineQuery.ID,
				Results: []interface{}{
					tgbotapi.NewInlineQueryResultArticle(inlineBotID, "List of "+update.InlineQuery.Query, "List of "+update.InlineQuery.Query),
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
