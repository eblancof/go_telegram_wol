package bot

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net"
	"os"
	"strings"

	"github.com/eblancof/telegram-bot/internal/config"
	"github.com/eblancof/telegram-bot/internal/device"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

const (
	port        = 9
	dataFile    = "devices.json"
	cmdWOL      = "wol"
	cmdAdd      = "add"
	cmdModify   = "modify"
	cmdDelete   = "delete"
	cmdCancel   = "cancel"
	cmdList     = "list"
	botCommands = `
[
    {"command":"wol","description":"Wake up a device"},
    {"command":"add","description":"Add a new device"},
    {"command":"modify","description":"Modify existing device"},
    {"command":"delete","description":"Delete a device"},
    {"command":"list","description":"List all devices"},
    {"command":"help","description":"Show available options"}
]`
	cmdAddName      = "add_name"
	cmdAddMAC       = "add_mac"
	cmdModifyName   = "modify_name"
	cmdModifyMAC    = "modify_mac"
	cmdModifyDevice = "modify_device"
)

var devices []device.Computer

func loadDevices() error {
	file, err := os.ReadFile(dataFile)
	if err != nil {
		return err
	}
	return json.Unmarshal(file, &devices)
}

func saveDevices() error {
	data, err := json.MarshalIndent(devices, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(dataFile, data, 0644)
}

func sendWakeOnLAN(macAddress string) error {
	mac, err := hex.DecodeString(strings.Replace(macAddress, ":", "", -1))
	if err != nil {
		return err
	}

	var packet []byte
	packet = append(packet, []byte{255, 255, 255, 255, 255, 255}...)
	for i := 0; i < 16; i++ {
		packet = append(packet, mac...)
	}

	conn, err := net.DialUDP("udp", nil, &net.UDPAddr{
		IP:   net.ParseIP(config.GetBroadcastIP()),
		Port: port,
	})
	if err != nil {
		return err
	}
	defer conn.Close()

	_, err = conn.Write(packet)
	return err
}

func HandleMessages(bot *tgbotapi.BotAPI) {
	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates := bot.GetUpdatesChan(u)
	for update := range updates {
		if update.Message != nil && update.Message.Chat.ID != config.GetChatID() {
			sendUnauthorizedMessage(bot, update.Message.Chat.ID)
			continue
		}

		if update.CallbackQuery != nil {
			handleCallbackQuery(bot, update.CallbackQuery)
			continue
		}

		if update.Message == nil {
			continue
		}

		if update.Message.Chat.ID != config.GetChatID() {
			sendUnauthorizedMessage(bot, update.Message.Chat.ID)
			continue
		}

		handleCommand(bot, update.Message)
	}
}

func handleCallbackQuery(bot *tgbotapi.BotAPI, query *tgbotapi.CallbackQuery) {
	deleteMsg := tgbotapi.NewDeleteMessage(query.Message.Chat.ID, query.Message.MessageID)
	bot.Send(deleteMsg)

	cleanupButtonMessages(bot, query.Message.Chat.ID)

	data := strings.Split(query.Data, ":")
	switch data[0] {
	case cmdWOL:
		if len(data) > 1 {
			for _, device := range devices {
				if device.Name == data[1] {
					err := sendWakeOnLAN(device.MAC)
					replyText := "WoL packet sent to " + device.Name
					if err != nil {
						replyText = "Failed to send WoL packet"
					}
					bot.Send(tgbotapi.NewMessage(query.Message.Chat.ID, replyText))
				}
			}
		} else {
			sendWolMessage(bot, query.Message.Chat.ID)
		}
	case cmdAdd:
		startAddDevice(bot, query.Message.Chat.ID)
	case cmdModify:
		if len(data) == 1 {
			sendModifyMessage(bot, query.Message.Chat.ID)
		} else {
			sendModifyOptionsMessage(bot, query.Message.Chat.ID, data[1])
		}
	case cmdModifyName:
		if len(data) > 1 {
			startModifyName(bot, query.Message.Chat.ID, data[1])
		}
	case cmdModifyMAC:
		if len(data) > 1 {
			startModifyMAC(bot, query.Message.Chat.ID, data[1])
		}
	case cmdDelete:
		if len(data) > 1 {
			handleDeleteDevice(bot, data[1], query.Message.Chat.ID)
		} else {
			sendDeleteMessage(bot, query.Message.Chat.ID)
		}
	case cmdCancel:
		delete(addDeviceStates, query.Message.Chat.ID)
		delete(modifyDeviceStates, query.Message.Chat.ID)
		bot.Send(tgbotapi.NewMessage(query.Message.Chat.ID, "Operation cancelled"))
	}
}

func sendUnauthorizedMessage(bot *tgbotapi.BotAPI, chatID int64) {
	msg := tgbotapi.NewMessage(chatID, "Unauthorized user")
	bot.Send(msg)
}

func sendDeviceList(bot *tgbotapi.BotAPI, chatID int64) {
	if len(devices) == 0 {
		bot.Send(tgbotapi.NewMessage(chatID, "No devices found."))
		return
	}

	var deviceList string
	for _, device := range devices {
		deviceList += fmt.Sprintf("📱 %s\nMAC: %s\n\n", device.Name, device.MAC)
	}

	msg := tgbotapi.NewMessage(chatID, "Saved Devices:\n\n"+deviceList)
	bot.Send(msg)
}

func handleCommand(bot *tgbotapi.BotAPI, message *tgbotapi.Message) {
	switch message.Command() {
	case "help":
		sendHelpMessage(bot, message.Chat.ID)
	case cmdWOL:
		sendWolMessage(bot, message.Chat.ID)
	case cmdAdd:
		startAddDevice(bot, message.Chat.ID)
	case cmdModify:
		sendModifyMessage(bot, message.Chat.ID)
	case cmdDelete:
		sendDeleteMessage(bot, message.Chat.ID)
	case cmdList:
		sendDeviceList(bot, message.Chat.ID)
	default:
		handleDefaultMessage(bot, message)
	}
}

func sendHelpMessage(bot *tgbotapi.BotAPI, chatID int64) {
	helpText := `📱 *WOL Device Manager* 📱

Available Commands:
/help - Show this help message
/wol - Wake up a device
/add - Add a new device
/modify - Modify existing device
/delete - Delete a device
/list - List all saved devices

How to use:
1. Quick Wake Up:
   • Use the keyboard buttons below to instantly wake up devices
   • Just tap a device name to wake it up

2. Device Management:
   • Add: Use /add and follow the prompts
   • Modify: Use /modify to change name or MAC address
   • Delete: Use /delete to remove devices
   • List: Use /list to see all devices and their MACs

3. Manual Wake Up:
   • Type a device name to wake it up
   • Use /wol command for button interface

MAC Address Format: XX:XX:XX:XX:XX:XX

Note: The keyboard below updates automatically when you add/modify/delete devices.`

	msg := tgbotapi.NewMessage(chatID, helpText)
	msg.ParseMode = "Markdown"

	buttons := [][]tgbotapi.InlineKeyboardButton{
		{
			tgbotapi.NewInlineKeyboardButtonData("💻 Wake Device", cmdWOL),
			tgbotapi.NewInlineKeyboardButtonData("➕ Add Device", cmdAdd),
		},
		{
			tgbotapi.NewInlineKeyboardButtonData("✏️ Modify Device", cmdModify),
			tgbotapi.NewInlineKeyboardButtonData("❌ Delete Device", cmdDelete),
		},
	}

	msg.ReplyMarkup = tgbotapi.InlineKeyboardMarkup{InlineKeyboard: buttons}
	sent, _ := bot.Send(msg)
	if sent.MessageID != 0 {
		addButtonMessage(chatID, sent.MessageID)
	}
}

func sendWolMessage(bot *tgbotapi.BotAPI, chatID int64) {
	msg := tgbotapi.NewMessage(chatID, "Select a device to wake up:")
	var buttons [][]tgbotapi.InlineKeyboardButton

	for _, device := range devices {
		button := tgbotapi.NewInlineKeyboardButtonData(device.Name, fmt.Sprintf("%s:%s", cmdWOL, device.Name))
		buttons = append(buttons, []tgbotapi.InlineKeyboardButton{button})
	}

	buttons = append(buttons, []tgbotapi.InlineKeyboardButton{
		tgbotapi.NewInlineKeyboardButtonData("❌ Cancel", fmt.Sprintf("%s", cmdCancel)),
	})

	msg.ReplyMarkup = tgbotapi.InlineKeyboardMarkup{InlineKeyboard: buttons}
	sent, _ := bot.Send(msg)
	if sent.MessageID != 0 {
		addButtonMessage(chatID, sent.MessageID)
	}
}

func sendAddMessage(bot *tgbotapi.BotAPI, chatID int64) {
	msg := tgbotapi.NewMessage(chatID, "Send the device info in the format: name,mac")
	bot.Send(msg)
}

func sendModifyMessage(bot *tgbotapi.BotAPI, chatID int64) {
	msg := tgbotapi.NewMessage(chatID, "Select a device to modify:")
	var buttons [][]tgbotapi.InlineKeyboardButton

	for _, device := range devices {
		button := tgbotapi.NewInlineKeyboardButtonData(device.Name, fmt.Sprintf("%s:%s", cmdModify, device.Name))
		buttons = append(buttons, []tgbotapi.InlineKeyboardButton{button})
	}

	buttons = append(buttons, []tgbotapi.InlineKeyboardButton{
		tgbotapi.NewInlineKeyboardButtonData("❌ Cancel", cmdCancel),
	})

	msg.ReplyMarkup = tgbotapi.InlineKeyboardMarkup{InlineKeyboard: buttons}
	sent, _ := bot.Send(msg)
	if sent.MessageID != 0 {
		addButtonMessage(chatID, sent.MessageID)
	}
}

func sendModifyOptionsMessage(bot *tgbotapi.BotAPI, chatID int64, deviceName string) {
	msg := tgbotapi.NewMessage(chatID, fmt.Sprintf("What would you like to modify for %s?", deviceName))
	buttons := [][]tgbotapi.InlineKeyboardButton{
		{
			tgbotapi.NewInlineKeyboardButtonData("Modify Name", fmt.Sprintf("%s:%s", cmdModifyName, deviceName)),
			tgbotapi.NewInlineKeyboardButtonData("Modify MAC", fmt.Sprintf("%s:%s", cmdModifyMAC, deviceName)),
		},
		{
			tgbotapi.NewInlineKeyboardButtonData("❌ Cancel", cmdCancel),
		},
	}
	msg.ReplyMarkup = tgbotapi.InlineKeyboardMarkup{InlineKeyboard: buttons}
	sent, _ := bot.Send(msg)
	if sent.MessageID != 0 {
		addButtonMessage(chatID, sent.MessageID)
		confirm := tgbotapi.NewMessage(chatID, "Select what you'd like to modify:")
		bot.Send(confirm)
	}
}

func sendDeleteMessage(bot *tgbotapi.BotAPI, chatID int64) {
	msg := tgbotapi.NewMessage(chatID, "Select a device to delete:")
	var buttons [][]tgbotapi.InlineKeyboardButton

	for _, device := range devices {
		button := tgbotapi.NewInlineKeyboardButtonData(device.Name, fmt.Sprintf("%s:%s", cmdDelete, device.Name))
		buttons = append(buttons, []tgbotapi.InlineKeyboardButton{button})
	}

	buttons = append(buttons, []tgbotapi.InlineKeyboardButton{
		tgbotapi.NewInlineKeyboardButtonData("❌ Cancel", fmt.Sprintf("%s", cmdCancel)),
	})

	msg.ReplyMarkup = tgbotapi.InlineKeyboardMarkup{InlineKeyboard: buttons}
	sent, _ := bot.Send(msg)
	if sent.MessageID != 0 {
		addButtonMessage(chatID, sent.MessageID)
	}
}

func handleDefaultMessage(bot *tgbotapi.BotAPI, message *tgbotapi.Message) {
	if state, exists := addDeviceStates[message.Chat.ID]; exists {
		handleAddDeviceState(bot, message, state)
		return
	}
	if state, exists := modifyDeviceStates[message.Chat.ID]; exists {
		handleModifyDeviceState(bot, message, state)
		return
	}
	if strings.Contains(message.Text, ",") {
		processDeviceCommand(bot, message.Text, message.Chat.ID)
	} else {
		checkAndSendWolPacket(bot, message)
	}
}

func handleModifyDeviceState(bot *tgbotapi.BotAPI, message *tgbotapi.Message, state *ModifyDeviceState) {
	for i, device := range devices {
		if device.Name == state.DeviceName {
			switch state.Field {
			case "name":
				oldName := device.Name
				devices[i].Name = message.Text
				msg := tgbotapi.NewMessage(message.Chat.ID,
					fmt.Sprintf("Device name updated from %s to %s\nWould you like to modify the MAC address as well?", oldName, message.Text))
				buttons := [][]tgbotapi.InlineKeyboardButton{
					{
						tgbotapi.NewInlineKeyboardButtonData("Modify MAC", fmt.Sprintf("%s:%s", cmdModifyMAC, message.Text)),
						tgbotapi.NewInlineKeyboardButtonData("❌ Done", cmdCancel),
					},
				}
				msg.ReplyMarkup = tgbotapi.InlineKeyboardMarkup{InlineKeyboard: buttons}
				sent, _ := bot.Send(msg)
				if sent.MessageID != 0 {
					addButtonMessage(message.Chat.ID, sent.MessageID)
				}
			case "mac":
				if validateMAC(message.Text) {
					devices[i].MAC = message.Text
					msg := tgbotapi.NewMessage(message.Chat.ID,
						fmt.Sprintf("MAC address updated for %s\nWould you like to modify the name as well?", state.DeviceName))
					buttons := [][]tgbotapi.InlineKeyboardButton{
						{
							tgbotapi.NewInlineKeyboardButtonData("Modify Name", fmt.Sprintf("%s:%s", cmdModifyName, state.DeviceName)),
							tgbotapi.NewInlineKeyboardButtonData("❌ Done", cmdCancel),
						},
					}
					msg.ReplyMarkup = tgbotapi.InlineKeyboardMarkup{InlineKeyboard: buttons}
					sent, _ := bot.Send(msg)
					if sent.MessageID != 0 {
						addButtonMessage(message.Chat.ID, sent.MessageID)
					}
				} else {
					bot.Send(tgbotapi.NewMessage(message.Chat.ID, "Invalid MAC address format. Operation cancelled."))
				}
			}
			saveDevices()
			updateKeyboard(bot, message.Chat.ID)
			delete(modifyDeviceStates, message.Chat.ID)
			return
		}
	}
	bot.Send(tgbotapi.NewMessage(message.Chat.ID, "Device not found. Operation cancelled."))
	delete(modifyDeviceStates, message.Chat.ID)
}

func checkAndSendWolPacket(bot *tgbotapi.BotAPI, message *tgbotapi.Message) {
	for _, device := range devices {
		if message.Text == device.Name {
			err := sendWakeOnLAN(device.MAC)
			if err != nil {
				bot.Send(tgbotapi.NewMessage(message.Chat.ID, "Failed to send WoL packet."))
			} else {
				bot.Send(tgbotapi.NewMessage(message.Chat.ID, "WoL packet sent to "+device.Name))
			}
			break
		}
	}
}

func processDeviceCommand(bot *tgbotapi.BotAPI, text string, chatID int64) {
	parts := strings.Split(text, ",")
	switch len(parts) {
	case 2:
		handleAddDevice(bot, parts, chatID)
	case 3:
		handleModifyDevice(bot, parts, chatID)
	case 1:
		handleDeleteDevice(bot, parts[0], chatID)
	default:
		bot.Send(tgbotapi.NewMessage(chatID, "Invalid command format."))
	}
}

func handleAddDevice(bot *tgbotapi.BotAPI, parts []string, chatID int64) {
	newDevice := device.Computer{Name: parts[0], MAC: parts[1]}
	if validateMAC(newDevice.MAC) {
		devices = append(devices, newDevice)
		saveDevices()
		bot.Send(tgbotapi.NewMessage(chatID, "Device added: "+newDevice.Name))
		updateKeyboard(bot, chatID)
	} else {
		bot.Send(tgbotapi.NewMessage(chatID, "Invalid MAC address format."))
	}
}

func handleModifyDevice(bot *tgbotapi.BotAPI, parts []string, chatID int64) {
	oldName := parts[0]
	newName := parts[1]
	newMAC := parts[2]
	for i, device := range devices {
		if device.Name == oldName {
			if validateMAC(newMAC) {
				devices[i].Name = newName
				devices[i].MAC = newMAC
				saveDevices()
				bot.Send(tgbotapi.NewMessage(chatID, "Device modified: "+newName))
				updateKeyboard(bot, chatID)
			} else {
				bot.Send(tgbotapi.NewMessage(chatID, "Invalid MAC address format."))
			}
			return
		}
	}
	bot.Send(tgbotapi.NewMessage(chatID, "Device not found."))
}

func handleDeleteDevice(bot *tgbotapi.BotAPI, deviceName string, chatID int64) {
	for i, device := range devices {
		if device.Name == deviceName {
			devices = append(devices[:i], devices[i+1:]...)
			saveDevices()
			bot.Send(tgbotapi.NewMessage(chatID, "Device deleted: "+deviceName))
			updateKeyboard(bot, chatID)
			return
		}
	}
	bot.Send(tgbotapi.NewMessage(chatID, "Device not found."))
}

func validateMAC(mac string) bool {
	_, err := hex.DecodeString(strings.Replace(mac, ":", "", -1))
	return err == nil
}

func setBotCommands(bot *tgbotapi.BotAPI) error {
	var commands []tgbotapi.BotCommand
	if err := json.Unmarshal([]byte(botCommands), &commands); err != nil {
		return err
	}

	_, err := bot.Request(tgbotapi.NewSetMyCommands(commands...))
	return err
}

func startAddDevice(bot *tgbotapi.BotAPI, chatID int64) {
	addDeviceStates[chatID] = &AddDeviceState{Stage: cmdAddName}
	msg := tgbotapi.NewMessage(chatID, "Please enter the name for the new device:")
	cancelButton := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("❌ Cancel", cmdCancel),
		),
	)
	msg.ReplyMarkup = cancelButton
	sent, _ := bot.Send(msg)
	if sent.MessageID != 0 {
		addButtonMessage(chatID, sent.MessageID)
	}
}

func handleAddDeviceState(bot *tgbotapi.BotAPI, message *tgbotapi.Message, state *AddDeviceState) {
	switch state.Stage {
	case cmdAddName:
		state.Name = message.Text
		state.Stage = cmdAddMAC
		msg := tgbotapi.NewMessage(message.Chat.ID,
			fmt.Sprintf("Device name set to: %s\nPlease enter the MAC address (format: XX:XX:XX:XX:XX:XX):", state.Name))
		cancelButton := tgbotapi.NewInlineKeyboardMarkup(
			tgbotapi.NewInlineKeyboardRow(
				tgbotapi.NewInlineKeyboardButtonData("❌ Cancel", cmdCancel),
			),
		)
		msg.ReplyMarkup = cancelButton
		sent, _ := bot.Send(msg)
		if sent.MessageID != 0 {
			addButtonMessage(message.Chat.ID, sent.MessageID)
		}

	case cmdAddMAC:
		if validateMAC(message.Text) {
			newDevice := device.Computer{Name: state.Name, MAC: message.Text}
			devices = append(devices, newDevice)
			saveDevices()
			bot.Send(tgbotapi.NewMessage(message.Chat.ID,
				fmt.Sprintf("Device added successfully!\nName: %s\nMAC: %s", newDevice.Name, newDevice.MAC)))
			updateKeyboard(bot, message.Chat.ID)
			delete(addDeviceStates, message.Chat.ID)
		} else {
			msg := tgbotapi.NewMessage(message.Chat.ID,
				"Invalid MAC address format. Please try again (format: XX:XX:XX:XX:XX:XX):")
			cancelButton := tgbotapi.NewInlineKeyboardMarkup(
				tgbotapi.NewInlineKeyboardRow(
					tgbotapi.NewInlineKeyboardButtonData("❌ Cancel", cmdCancel),
				),
			)
			msg.ReplyMarkup = cancelButton
			sent, _ := bot.Send(msg)
			if sent.MessageID != 0 {
				addButtonMessage(message.Chat.ID, sent.MessageID)
			}
		}
	}
}

func startModifyName(bot *tgbotapi.BotAPI, chatID int64, deviceName string) {
	modifyDeviceStates[chatID] = &ModifyDeviceState{
		DeviceName: deviceName,
		Field:      "name",
	}
	msg := tgbotapi.NewMessage(chatID, fmt.Sprintf("Enter the new name for %s:", deviceName))
	cancelButton := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("❌ Cancel", cmdCancel),
		),
	)
	msg.ReplyMarkup = cancelButton
	sent, _ := bot.Send(msg)
	if sent.MessageID != 0 {
		addButtonMessage(chatID, sent.MessageID)
	}
}

func startModifyMAC(bot *tgbotapi.BotAPI, chatID int64, deviceName string) {
	modifyDeviceStates[chatID] = &ModifyDeviceState{
		DeviceName: deviceName,
		Field:      "mac",
	}
	msg := tgbotapi.NewMessage(chatID, fmt.Sprintf("Enter the new MAC address for %s (format: XX:XX:XX:XX:XX:XX):", deviceName))
	cancelButton := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("❌ Cancel", cmdCancel),
		),
	)
	msg.ReplyMarkup = cancelButton
	sent, _ := bot.Send(msg)
	if sent.MessageID != 0 {
		addButtonMessage(chatID, sent.MessageID)
	}
}

func addButtonMessage(chatID int64, messageID int) {
	buttonMessages[chatID] = append(buttonMessages[chatID], messageID)
}

func cleanupButtonMessages(bot *tgbotapi.BotAPI, chatID int64) {
	if messages, exists := buttonMessages[chatID]; exists {
		for _, msgID := range messages {
			deleteMsg := tgbotapi.NewDeleteMessage(chatID, msgID)
			bot.Send(deleteMsg)
		}
		delete(buttonMessages, chatID)
	}
}

func createDeviceKeyboard() tgbotapi.ReplyKeyboardMarkup {
	var rows [][]tgbotapi.KeyboardButton
	var row []tgbotapi.KeyboardButton
	if len(devices) == 0 {
		return tgbotapi.NewReplyKeyboard(
			tgbotapi.NewKeyboardButtonRow(
				tgbotapi.NewKeyboardButton("/add"),
			),
		)
	}

	for i, device := range devices {
		row = append(row, tgbotapi.NewKeyboardButton(device.Name))

		if (i+1)%2 == 0 || i == len(devices)-1 {
			rows = append(rows, row)
			row = []tgbotapi.KeyboardButton{}
		}
	}

	return tgbotapi.NewReplyKeyboard(rows...)
}

func updateKeyboard(bot *tgbotapi.BotAPI, chatID int64) {
	msg := tgbotapi.NewMessage(chatID, "Keyboard updated with current devices.")
	msg.ReplyMarkup = createDeviceKeyboard()
	bot.Send(msg)
}
