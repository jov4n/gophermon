# Gophermon Discord Bot

A Pokémon-style Discord bot game featuring procedurally generated gophers, turn-based battles, party management, and evolution system.

## Setup

1. Clone this repository
2. Copy `.env.example` to `.env` and fill in your Discord bot token
3. Get your Discord bot token from [Discord Developer Portal](https://discord.com/developers/applications)
4. Run `go mod download` to install dependencies
5. Run `go run cmd/bot/main.go` to start the bot

## Getting a Discord Bot Token

1. Go to https://discord.com/developers/applications
2. Click "New Application" and give it a name
3. Go to the "Bot" section
4. Click "Add Bot" and confirm
5. Under "Token", click "Reset Token" or "Copy" to get your bot token
6. Enable "Message Content Intent" and "Server Members Intent" in the Bot section
7. Go to "OAuth2" > "URL Generator"
8. Select "bot" and "applications.commands" scopes
9. Select necessary permissions (Send Messages, Embed Links, Attach Files, etc.)
10. Copy the generated URL and open it in your browser to invite the bot to your server

## Commands

- `/start` - Begin your journey and choose a starter gopher
- `/choose <1|2|3>` - Select your starter gopher
- `/party` - View your active party
- `/pc` - View gophers in storage
- `/pc deposit <gopher_id>` - Move a gopher from party to PC
- `/pc withdraw <gopher_id>` - Move a gopher from PC to party
- `/wild` - Encounter a wild gopher
- `/gopher info <gopher_id>` - View detailed gopher information

## Project Structure

- `cmd/bot/` - Bot entry point
- `internal/discord/` - Discord command handlers
- `internal/game/` - Game logic (gophers, battles, abilities)
- `internal/gopherkon/` - Sprite generation
- `internal/storage/` - Database repositories
- `migrations/` - Database schema migrations
- `assets/gopherkon/` - Gopherkon sprite assets

