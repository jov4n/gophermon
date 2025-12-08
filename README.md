# Gophermon Discord Bot

A Pokémon-style Discord bot game featuring procedurally generated gophers, turn-based battles, party management, and evolution system.

## Setup

1. Clone this repository
2. Copy `.env.example` to `.env` and fill in your Discord bot token
3. Get your Discord bot token from [Discord Developer Portal](https://discord.com/developers/applications)
4. Run `go mod download` to install dependencies
5. Download gopherize.me artwork (see [Artwork Setup](#artwork-setup) below)
6. Run `go run cmd/bot/main.go` to start the bot

## Artwork Setup

This bot uses artwork from [gopherize.me](https://gopherize.me). You need to download the artwork before running the bot.

### Automatic Download (Recommended)

The artwork can be downloaded automatically from the gopherize.me API:

```bash
go run scripts/download_artwork.go
```

This script will:
- Fetch artwork metadata from https://gopherize.me/api/artwork
- Download all categories and images
- Organize them into the correct folder structure (`assets/artwork/010-Body/`, etc.)

### Manual Download

If the automatic download doesn't work, see `scripts/download_artwork_manual.md` for manual instructions.

The artwork should be organized in `assets/artwork/` with numbered category folders:
- `000-Body/` - Base body features
- `010-Eyes/` - Eye features
- `020-Mouth/` - Mouth features
- etc.

Each category folder contains PNG files that are overlaid in order to create the final gopher image.

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
- `assets/artwork/` - Gopherize.me artwork (numbered category folders)
- `assets/generated/` - Generated gopher sprites

