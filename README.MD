# Mail Gather Bot

A Telegram bot to gather mails for an event in a group.

## Running

To run the bot, at first build it. Then set `DATABASE_NAME` and `API_TOKEN` environment variables. For example:

```bash
go build -o mail-gather-bot
export API_TOKEN=123456:ABC-DEF1234ghIkl-zyx57W2v1u123ew11
export DATABASE_NAME=database.db
./mail-gather-bot
```