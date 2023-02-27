package bot

import (
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"log"
	"strings"
)

// UpdateList will refresh a list by its ID
func (b *Bot) UpdateList(listID string) {
	// Get the users
	title, emails, err := b.Database.GetListEmails(listID)
	if err != nil {
		log.Printf("cannot get email list of %s: %s\n", listID, err)
		return
	}
	// Edit the list
	edit := tgbotapi.EditMessageTextConfig{
		BaseEdit: tgbotapi.BaseEdit{
			InlineMessageID: listID,
		},
		Text: "List of " + title + "\n" + strings.Join(emails, "\n"),
	}
	_, _ = b.bot.Send(edit)
}
