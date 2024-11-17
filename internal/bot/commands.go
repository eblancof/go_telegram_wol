package bot

import (
	"encoding/json"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

const botCommandsList = `
[
	{"command":"wol","description":"Wake up a device"},
	{"command":"add","description":"Add a new device"},
	{"command":"modify","description":"Modify existing device"},
	{"command":"delete","description":"Delete a device"},
	{"command":"list","description":"List all devices"},
	{"command":"help","description":"Show available options"}
]`

func SetCommands(bot *tgbotapi.BotAPI) error {
	var commands []tgbotapi.BotCommand
	if err := json.Unmarshal([]byte(botCommandsList), &commands); err != nil {
		return err
	}

	_, err := bot.Request(tgbotapi.NewSetMyCommands(commands...))
	return err
}
