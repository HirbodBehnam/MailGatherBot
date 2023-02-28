package bot

import (
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"log"
	"strconv"
	"strings"
)

// UpdateList will refresh a list by its ID
func (b *Bot) UpdateList(listID, inlineMessageID string) {
	// Get the users
	title, emails, err := b.Database.GetListEmails(listID)
	if err != nil {
		log.Printf("cannot get email list of %s: %s\n", listID, err)
		return
	}
	// Edit the list
	edit := tgbotapi.EditMessageTextConfig{
		BaseEdit: tgbotapi.BaseEdit{
			InlineMessageID: inlineMessageID,
			ReplyMarkup: &tgbotapi.InlineKeyboardMarkup{
				InlineKeyboard: [][]tgbotapi.InlineKeyboardButton{
					{
						{
							Text:         messageButtonText,
							CallbackData: &listID,
						},
					},
				},
			},
		},
		Text:      "List of " + title + "\nParticipants: " + strconv.Itoa(len(emails)) + "\n`" + strings.Join(emails, "\n") + "`",
		ParseMode: "MarkdownV2",
	}
	_, _ = b.bot.Send(edit)
}
