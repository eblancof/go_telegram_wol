package main

import (
	"log"

	"github.com/eblancof/telegram-bot/internal/bot"
	"github.com/eblancof/telegram-bot/internal/config"
	"github.com/eblancof/telegram-bot/internal/device"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

func main() {
	cfg := config.Load()

	if err := device.LoadDevices(); err != nil {
		log.Println("No existing devices found. Starting fresh.")
	}

	botAPI, err := tgbotapi.NewBotAPI(cfg.BotToken)
	if err != nil {
		log.Panic(err)
	}

	botAPI.Debug = true
	log.Printf("Authorized on account %s", botAPI.Self.UserName)

	if err := bot.SetCommands(botAPI); err != nil {
		log.Printf("Failed to set commands: %v", err)
	}

	msg := tgbotapi.NewMessage(cfg.ChatID, "Started bot")
	msg.ReplyMarkup = bot.CreateDeviceKeyboard()
	botAPI.Send(msg)

	bot.HandleMessages(botAPI)
}
