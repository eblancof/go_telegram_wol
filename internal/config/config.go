package config

import (
	"os"
	"path/filepath"
	"strconv"
	"sync"

	"github.com/joho/godotenv"
)

type Config struct {
	BotToken    string
	ChatID      int64
	BroadcastIP string
	Port        int
	DataFile    string
}

var (
	instance *Config
	once     sync.Once
)

func Load() *Config {
	once.Do(func() {
		// Try to load .env from multiple possible locations
		locations := []string{
			".env",
			"../../.env",
			filepath.Join(os.Getenv("HOME"), "go_wolserver", ".env"),
		}

		for _, loc := range locations {
			if err := godotenv.Load(loc); err == nil {
				break
			}
		}

		chatID, _ := strconv.ParseInt(os.Getenv("CHAT_ID"), 10, 64)
		instance = &Config{
			BotToken:    os.Getenv("BOT_TOKEN"),
			ChatID:      chatID,
			BroadcastIP: os.Getenv("BROADCAST_IP"),
			Port:        9,
			DataFile:    "devices.json",
		}
	})
	return instance
}

// Getters
func GetChatID() int64 {
	return Load().ChatID
}

func GetBroadcastIP() string {
	return Load().BroadcastIP
}

func GetPort() int {
	return Load().Port
}

func GetDataFile() string {
	return Load().DataFile
}

func GetBotToken() string {
	return Load().BotToken
}
