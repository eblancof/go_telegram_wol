# Wake on LAN (WoL) Telegram Bot

A Telegram bot written in Go that allows you to remotely wake up computers using Wake-on-LAN technology.

## Features

- 🚀 Wake computers remotely via Telegram commands
- ⚡ Fast and lightweight
- 💾 Store devices in a json file
- 🖥️ Support for multiple machines

## Prerequisites

- Go 1.20 or higher
- A Telegram Bot Token (get it from [@BotFather](https://t.me/botfather))
- Computers with Wake-on-LAN enabled in BIOS/UEFI

## Installation

1. Clone the repository:
    
```bash
git clone https://github.com/eblancof/go_telegram_wol.git
```
2. Create a .env file in the root directory with the following content:

```bash
# Create a Telegram bot with BotFather and paste the token here
BOT_TOKEN=xxxxxxxxx:yyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyy
# Chat ID to restrict access to the bot (Your user ID)
# To get your chat ID use https://t.me/getmyid_bot
CHAT_ID=xxxxxxxxx
# Broadcast IP address
BROADCAST_IP=192.168.1.255
```
3. Install the dependencies:

```bash
go mod tidy
```
4. Build the project:

```bash
go build -o go_telegram_wol
```
5. Run the binary:

```bash
./go_telegram_wol
```
## Usage
Available commands:

* /wol - 🖥️ Wake up a computer
* /add - ➕ Add a new computer
* /delete - ❌ Remove a computer
* /modify - ✏️ Modify a computer
* /list - 📋 List all computers
* /help - ℹ️ Show help message


