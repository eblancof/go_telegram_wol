package bot

import (
	"github.com/eblancof/telegram-bot/internal/device"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

func CreateDeviceKeyboard() tgbotapi.ReplyKeyboardMarkup {
	var rows [][]tgbotapi.KeyboardButton
	var row []tgbotapi.KeyboardButton

	if len(device.Devices) == 0 {
		return tgbotapi.NewReplyKeyboard(
			tgbotapi.NewKeyboardButtonRow(
				tgbotapi.NewKeyboardButton("/add"),
			),
		)
	}

	for i, dev := range device.Devices {
		row = append(row, tgbotapi.NewKeyboardButton(dev.Name))

		if (i+1)%2 == 0 || i == len(device.Devices)-1 {
			rows = append(rows, row)
			row = []tgbotapi.KeyboardButton{}
		}
	}

	return tgbotapi.NewReplyKeyboard(rows...)
}
