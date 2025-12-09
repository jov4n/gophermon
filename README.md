# Gophermon Discord Bot

A Pokémon-style Discord bot game featuring procedurally generated gophers, turn-based battles, party management, evolution system, shiny gophers, and dynamic events!

## Features

- **Procedurally Generated Gophers**: Unique gophers created from gopherize.me artwork with 5 rarity tiers
- **Turn-Based Battles**: Fight wild gophers with abilities, status effects, and type advantages
- **Party Management**: Build a team of up to 6 gophers, store extras in PC
- **Evolution System**: Gophers evolve at levels 16 and 32 with visual and stat upgrades
- **Shiny Gophers**: Rare color-inverted gophers with golden glow effects and 25% stat boost (1/4096 base rate)
- **Event System**: 6 different event types that modify gameplay (auto-scheduled or manual)
- **Status Effects**: Burns, poison, paralysis, sleep, stat boosts/debuffs, and more
- **Type System**: 5 archetypes (Hacker, Tank, Speedy, Support, Mage) with type effectiveness
- **XP & Leveling**: Gain XP from battles, level up, and unlock new abilities

## Setup

1. Clone this repository
2. Copy `.env.example` to `.env` and fill in your Discord bot token
3. Get your Discord bot token from [Discord Developer Portal](https://discord.com/developers/applications)
4. Run `go mod download` to install dependencies
5. Download gopherize.me artwork (see [Artwork Setup](#artwork-setup) below)
6. Run database migrations (see [Database Setup](#database-setup) below)
7. Run `go run cmd/bot/main.go` to start the bot

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
- `010-Body/` - Base body features
- `020-Eyes/` - Eye features
- `021-Shirts/` - Shirt/clothing features
- `022-Hair/` - Hair styles
- `023-Facial_Hair/` - Beards and mustaches
- `024-Glasses/` - Glasses and sunglasses
- `025-Hats_and_Hair_Accessories/` - Hats and accessories
- `027-Extras/` - Extra decorative items

Each category folder contains PNG files that are overlaid in order to create the final gopher image.

## Database Setup

The bot uses SQLite for data storage. Database migrations are located in `migrations/`:

- `001_initial.sql` - Initial schema (trainers, gophers, parties, battles)
- `002_add_sprite_data.sql` - Base64 sprite storage
- `003_add_types.sql` - Type system (primary/secondary types)
- `004_add_status_effects.sql` - Status effects persistence
- `005_add_shiny.sql` - Shiny gopher support

The database is created automatically on first run. Migrations are applied automatically.

## Getting a Discord Bot Token

1. Go to https://discord.com/developers/applications
2. Click "New Application" and give it a name
3. Go to the "Bot" section
4. Click "Add Bot" and confirm
5. Under "Token", click "Reset Token" or "Copy" to get your bot token
6. Enable "Message Content Intent" and "Server Members Intent" in the Bot section
7. Go to "OAuth2" > "URL Generator"
8. Select "bot" and "applications.commands" scopes
9. Select necessary permissions (Send Messages, Embed Links, Attach Files, Manage Messages, etc.)
10. Copy the generated URL and open it in your browser to invite the bot to your server

## Environment Variables

Create a `.env` file in the project root with the following variables:

### Required
```env
DISCORD_TOKEN=your_discord_bot_token_here
```

### Optional
```env
# Database path (default: ./gophermon.db)
DB_PATH=./gophermon.db

# Bot prefix (not used for slash commands, default: !)
BOT_PREFIX=!

# Guild ID for faster command registration during development
GUILD_ID=your_guild_id_here

# Event system configuration
# Channel ID where event announcements will be sent (leave empty to disable)
EVENT_ANNOUNCE_CHANNEL=

# Enable/disable automatic event scheduling (true/false, default: true)
AUTO_EVENTS_ENABLED=true

# Hours between automatic events (default: 48)
AUTO_EVENT_INTERVAL=48

# Hours each automatic event lasts (default: 24)
AUTO_EVENT_DURATION=24
```

## Commands

### Player Commands

- `/ping` - Check if the bot is responding
- `/start` - Begin your journey and choose a starter gopher
- `/choose <number>` - Select your starter gopher (1, 2, or 3)
- `/party` - View your active party of gophers (up to 6)
- `/pc list [page]` - View gophers in PC storage (paginated)
- `/pc deposit <gopher_id>` - Move a gopher from party to PC
- `/pc withdraw <gopher_id>` - Move a gopher from PC to party
- `/wild` - Encounter a wild gopher (starts a battle)
- `/gopher info <gopher_id>` - View detailed information about a gopher

### Event Commands

- `/events list` - View all currently active events
- `/events start <type> [hours]` - Start a new event (admin only)
- `/events end <type>` - End an active event (admin only)

### Admin/Testing Commands

- `/generate_10` - Generate 15 gophers (3 of each rarity) with one shiny per tier

## Game Mechanics

### Gopher Rarities

- **COMMON** (60% of wild encounters) - Basic stats
- **UNCOMMON** (25% of wild encounters) - +15% stats
- **RARE** (10% of wild encounters) - +30% stats
- **EPIC** (4% of wild encounters) - +50% stats
- **LEGENDARY** (1% of wild encounters) - +80% stats, special abilities

### Gopher Archetypes

Each gopher has one of five archetypes that determine base stats and abilities:

- **Hacker** - High Attack and Speed, low Defense
- **Tank** - High HP and Defense, low Speed
- **Speedy** - Very high Speed, balanced other stats
- **Support** - Balanced stats, support abilities
- **Mage** - High Attack, balanced other stats

### Shiny Gophers

- **Base Rate**: 1 in 4,096 (1/4096)
- **Visual Effect**: Color-inverted sprite with golden glow effect
- **Stat Boost**: +25% to all stats (HP, Attack, Defense, Speed)
- **Shiny Hunt Event**: Increases rate to 1 in 100 (1/100)

### Evolution

- **Stage 1**: Level 16+ (reduced to 11+ during Evolution Festival)
- **Stage 2**: Level 32+ (reduced to 27+ during Evolution Festival)
- **Benefits**: Stat boosts, new sprite, rarity upgrade, new abilities

### Battles

- Turn-based combat with abilities
- Capture wild gophers with nets (chance based on HP and rarity)
- XP distribution to all participating gophers
- Auto-swap when active gopher faints
- Status effects (burns, poison, paralysis, sleep, stat modifiers)
- Type effectiveness system

### Events

The bot features 6 different event types that modify gameplay:

1. **Shiny Hunt** - Shiny spawn rate: 1/100 (normally 1/4096)
2. **Double XP** - 2x XP from battles
3. **Rare Encounter** - Higher chance of rare/legendary wild gophers
4. **Lucky Day** - 50% better capture rates
5. **Stat Boost** - All gophers get 10% stat boost in battles
6. **Evolution Festival** - Evolution requirements reduced by 5 levels

Events can be:
- **Auto-scheduled**: Random events start automatically (configurable interval)
- **Manual**: Admins can start/end events with `/events start` and `/events end`

## Project Structure

```
.
├── cmd/bot/              # Bot entry point
│   └── main.go          # Main application
├── internal/
│   ├── config/          # Configuration loading
│   ├── discord/         # Discord command handlers and routing
│   ├── game/            # Game logic
│   │   ├── abilities.go # Ability system
│   │   ├── battle.go    # Battle mechanics
│   │   ├── events.go    # Event system
│   │   ├── evolution.go # Evolution logic
│   │   ├── gopher.go    # Gopher data and stats
│   │   ├── rarity.go    # Rarity system
│   │   ├── service.go   # Game service layer
│   │   ├── trainer.go   # Trainer management
│   │   └── types.go     # Type effectiveness
│   ├── gopherkon/       # Sprite generation
│   │   ├── card.go      # Card image generation
│   │   └── generator.go # Sprite compositing and effects
│   └── storage/         # Database repositories
│       ├── battle_repo.go
│       ├── db.go
│       ├── gopher_repo.go
│       ├── party_repo.go
│       └── trainer_repo.go
├── migrations/          # Database schema migrations
├── assets/
│   ├── artwork/        # Gopherize.me artwork (numbered category folders)
│   └── generated/      # Generated gopher sprites
├── scripts/            # Utility scripts
│   ├── download_artwork.go
│   ├── generate_glitch_variants.go
│   ├── test_artwork_generation.go
│   ├── test_battle_card.go
│   └── test_gopher_generation.go
├── .env.example        # Example environment file
├── go.mod              # Go dependencies
└── README.md           # This file
```

## Development

### Running the Bot

```bash
go run cmd/bot/main.go
```

### Testing Commands

The bot includes several test scripts in `scripts/`:

- `test_gopher_generation.go` - Test gopher generation with various parameters
- `test_artwork_generation.go` - Test sprite generation
- `test_battle_card.go` - Test battle card rendering
- `generate_glitch_variants.go` - Generate variant gophers

### Database Migrations

Migrations are automatically applied on startup. To manually check the database:

```bash
sqlite3 gophermon.db
```

## Contributing

This is a personal project, but suggestions and improvements are welcome!

## License

This project uses artwork from [gopherize.me](https://gopherize.me). Please respect their terms of use.
