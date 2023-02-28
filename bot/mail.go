package bot

import (
	"MailGatherBot/util"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"log"
)

// mailChangeRequest gets the user ID and the message that they sent and checks if it's a
// valid email address. If it's a valid email address, it will change the email address of user.
// Otherwise, it will send the current email address of user to them.
func (b *Bot) mailChangeRequest(userID int64, message string) {
	// Check if valid email
	email, isEmail := util.ValidateEmail(message)
	if isEmail {
		// Change email request
		err := b.Database.UpdateEmail(userID, email)
		if err != nil {
			log.Printf("cannot set user %d's email address to %s: %s\n", userID, email, err)
			_, _ = b.bot.Send(tgbotapi.NewMessage(userID, "Cannot set your email."))
		} else {
			_, _ = b.bot.Send(tgbotapi.NewMessage(userID, "Changed your email!"))
		}
	} else {
		// Just send the email of user
		email, err := b.Database.GetEmail(userID)
		if err != nil {
			log.Printf("cannot get user %d's email address: %s\n", userID, err)
			_, _ = b.bot.Send(tgbotapi.NewMessage(userID, "Cannot get your email"))
		} else {
			if email == "" {
				_, _ = b.bot.Send(tgbotapi.NewMessage(userID, "You do not have an email address right now! Send your email address to me to set it."))
			} else {
				_, _ = b.bot.Send(tgbotapi.NewMessage(userID, "Your email address is "+email+". You can send your email to me if you want to change it."))
			}
		}
	}
}
