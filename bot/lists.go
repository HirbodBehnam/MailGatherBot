package bot

import (
	"github.com/PaulSonOfLars/gotgbot/v2"
	"github.com/go-faster/errors"
	"strconv"
	"strings"
)

// UpdateList will refresh a list by its ID
func (b *Bot) UpdateList(bot *gotgbot.Bot, listID, inlineMessageID string) error {
	// Get the users
	title, emails, err := b.Database.GetListEmails(listID)
	if err != nil {
		return errors.Wrap(err, "cannot get email list of "+listID)
	}
	// Edit the list
	_, _, err = bot.EditMessageText("List of "+title+"\nParticipants: "+strconv.Itoa(len(emails))+"\n`"+strings.Join(emails, "\n")+"`", &gotgbot.EditMessageTextOpts{
		InlineMessageId: inlineMessageID,
		ParseMode:       gotgbot.ParseModeMarkdownV2,
		ReplyMarkup: gotgbot.InlineKeyboardMarkup{
			InlineKeyboard: [][]gotgbot.InlineKeyboardButton{
				{
					{
						Text:         messageButtonText,
						CallbackData: listID,
					},
				},
			},
		},
	})
	return err
}
