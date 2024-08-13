package bot

import (
	"MailGatherBot/util"
	"github.com/PaulSonOfLars/gotgbot/v2"
	"github.com/PaulSonOfLars/gotgbot/v2/ext"
	"log"
)

// mailChangeRequest gets the user ID and the message that they sent and checks if it's a
// valid email address. If it's a valid email address, it will change the email address of user.
// Otherwise, it will send the current email address of user to them.
func (b *Bot) mailChangeRequest(bot *gotgbot.Bot, ctx *ext.Context) error {
	// Check if valid email
	email, isEmail := util.ValidateEmail(ctx.Message.Text)
	userID := ctx.EffectiveSender.Id()
	if isEmail {
		// Change email request
		err := b.Database.UpdateEmail(userID, email)
		if err != nil {
			log.Printf("cannot set user %d's email address to %s: %s\n", userID, email, err)
			_, err = ctx.EffectiveChat.SendMessage(bot, "Cannot set your email.", nil)
		} else {
			_, err = ctx.EffectiveChat.SendMessage(bot, "Changed your email!", nil)
		}
		return err
	} else {
		// Just send the email of user
		email, err := b.Database.GetEmail(userID)
		if err != nil {
			log.Printf("cannot get user %d's email address: %s\n", userID, err)
			_, err = ctx.EffectiveChat.SendMessage(bot, "Cannot get your email", nil)
		} else {
			if email == "" {
				_, err = ctx.EffectiveChat.SendMessage(bot, "You do not have an email address right now! Send your email address to me to set it.", nil)
			} else {
				_, err = ctx.EffectiveChat.SendMessage(bot, "Your email address is "+email+". You can send your email to me if you want to change it.", nil)
			}
		}
		return err
	}
}
