package bot

import (
	"MailGatherBot/database"
	"MailGatherBot/util"
	"github.com/PaulSonOfLars/gotgbot/v2"
	"github.com/PaulSonOfLars/gotgbot/v2/ext"
	"github.com/PaulSonOfLars/gotgbot/v2/ext/handlers"
	"github.com/PaulSonOfLars/gotgbot/v2/ext/handlers/filters/callbackquery"
	"github.com/PaulSonOfLars/gotgbot/v2/ext/handlers/filters/inlinequery"
	"github.com/PaulSonOfLars/gotgbot/v2/ext/handlers/filters/message"
	"github.com/go-faster/errors"
	"log"
	"time"
)

type Bot struct {
	ApiToken string
	Database database.Database
	// This channel must be closed when bot is done
	shutdownDone chan struct{}
	updater      *ext.Updater
}

func (b *Bot) Start() {
	// Make stuff ready for shutdown
	b.shutdownDone = make(chan struct{})
	defer close(b.shutdownDone)
	// Start the bot
	bot, err := gotgbot.NewBot(b.ApiToken, nil)
	if err != nil {
		log.Fatal("Cannot initialize the bot: ", err.Error())
	}
	log.Println("Bot authorized on account", bot.Username)
	dispatcher := ext.NewDispatcher(&ext.DispatcherOpts{
		Error: func(_ *gotgbot.Bot, _ *ext.Context, err error) ext.DispatcherAction {
			log.Println("An error occurred while handling update: ", err.Error())
			return ext.DispatcherActionNoop
		},
		MaxRoutines: ext.DefaultMaxRoutines,
	})
	b.updater = ext.NewUpdater(dispatcher, nil)
	// Add handlers
	dispatcher.AddHandler(handlers.NewCommand("about", func(b *gotgbot.Bot, ctx *ext.Context) error {
		_, err := ctx.EffectiveMessage.Reply(b, aboutMessage, nil)
		return err
	}))
	dispatcher.AddHandler(handlers.NewCommand("help", func(b *gotgbot.Bot, ctx *ext.Context) error {
		_, err := ctx.EffectiveMessage.Reply(b, helpMessage, nil)
		return err
	}))
	dispatcher.AddHandler(handlers.NewCallback(callbackquery.All, b.answerCallbackQuery))
	dispatcher.AddHandler(handlers.NewInlineQuery(inlinequery.All, b.handleCreateListQuery))
	dispatcher.AddHandler(handlers.NewMessage(message.Text, b.mailChangeRequest))
	// Wait for updates
	err = b.updater.StartPolling(bot, &ext.PollingOpts{
		DropPendingUpdates: true,
		GetUpdatesOpts: &gotgbot.GetUpdatesOpts{
			Timeout: 60,
			RequestOpts: &gotgbot.RequestOpts{
				Timeout: time.Second * 60,
			},
		},
	})
	if err != nil {
		panic("Failed to start polling: " + err.Error())
	}
	log.Printf("%s has been started...\n", bot.User.Username)

	// Idle, to keep updates coming in, and avoid bot stopping.
	b.updater.Idle()
}

// StopBot will gracefully stop the bot
func (b *Bot) StopBot() {
	_ = b.updater.Stop()
	<-b.shutdownDone        // wait for event loop to finish
	time.Sleep(time.Second) // wait for database operations to finish
	b.Database.Close()
}

// handleCreateListQuery will create a list for user
func (b *Bot) handleCreateListQuery(bot *gotgbot.Bot, ctx *ext.Context) error {
	queryID := ctx.InlineQuery.Id
	listName := ctx.InlineQuery.Query
	// If list name is empty just ignore it
	if listName == "" {
		_, err := bot.AnswerInlineQuery(queryID, []gotgbot.InlineQueryResult{}, nil)
		return err
	}
	// Create an ID for list
	listID := util.RandomID()
	// Create list in database
	err := b.Database.CreateList(listID, listName)
	if err != nil {
		return errors.Wrap(err, "cannot create list")
	}
	// Send it to user
	item := gotgbot.InlineQueryResultArticle{
		Id:    listID,
		Title: "List of " + listName,
		InputMessageContent: gotgbot.InputTextMessageContent{
			MessageText: "List of " + listName,
		},
		ReplyMarkup: &gotgbot.InlineKeyboardMarkup{InlineKeyboard: [][]gotgbot.InlineKeyboardButton{
			{
				{
					Text:         messageButtonText,
					CallbackData: listID,
				},
			},
		}},
	}
	_, err = bot.AnswerInlineQuery(queryID, []gotgbot.InlineQueryResult{item}, &gotgbot.AnswerInlineQueryOpts{
		CacheTime:  1,
		IsPersonal: false,
	})
	return err
}

func (b *Bot) answerCallbackQuery(bot *gotgbot.Bot, ctx *ext.Context) error {
	userID := ctx.EffectiveSender.Id()
	listID, queryID, messageID := ctx.CallbackQuery.Data, ctx.CallbackQuery.Id, ctx.CallbackQuery.InlineMessageId
	// Add or remove user from list
	err := b.Database.ParticipateOrOptOut(userID, listID)
	if err != nil {
		if errors.Is(err, database.NoEmailRegisteredErr) {
			_, err = bot.AnswerCallbackQuery(queryID, &gotgbot.AnswerCallbackQueryOpts{
				Url: "t.me/" + bot.Username + "?start=email",
			})
			return err
		} else {
			// Big fuckup
			return errors.Wrap(err, "cannot process user button click")
		}
	}
	// Refresh the list
	err = b.UpdateList(bot, listID, messageID)
	if err != nil {
		return errors.Wrap(err, "cannot update list")
	}
	// Answer it
	_, err = bot.AnswerCallbackQuery(queryID, nil)
	return err
}
