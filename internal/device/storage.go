package device

import (
	"encoding/json"
	"os"

	"github.com/eblancof/telegram-bot/internal/config"
)

func LoadDevices() error {
	file, err := os.ReadFile(config.GetDataFile())
	if err != nil {
		return err
	}
	return json.Unmarshal(file, &Devices)
}

func SaveDevices() error {
	data, err := json.MarshalIndent(Devices, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(config.GetDataFile(), data, 0644)
}
