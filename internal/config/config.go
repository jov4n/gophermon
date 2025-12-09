package config

import (
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

type Config struct {
	DiscordToken        string
	DBPath              string
	BotPrefix           string
	EventAnnounceChannel string // Discord channel ID for event announcements
	AutoEventsEnabled   bool    // Enable automatic event scheduling
	AutoEventInterval   int     // Hours between auto events (default: 48)
	AutoEventDuration   int     // Hours each auto event lasts (default: 24)
}

func Load() (*Config, error) {
	// Try to load .env file, but don't fail if it doesn't exist
	_ = godotenv.Load()

	// Parse boolean for auto events (default: true)
	autoEventsEnabled := true
	if getEnv("AUTO_EVENTS_ENABLED", "true") == "false" {
		autoEventsEnabled = false
	}

	// Parse integers with defaults
	autoEventInterval := parseInt(getEnv("AUTO_EVENT_INTERVAL", "48")) // 48 hours between events
	if autoEventInterval < 1 {
		autoEventInterval = 48
	}
	
	autoEventDuration := parseInt(getEnv("AUTO_EVENT_DURATION", "24")) // 24 hours per event
	if autoEventDuration < 1 {
		autoEventDuration = 24
	}

	return &Config{
		DiscordToken:         getEnv("DISCORD_TOKEN", ""),
		DBPath:              getEnv("DB_PATH", "./gophermon.db"),
		BotPrefix:           getEnv("BOT_PREFIX", "!"),
		EventAnnounceChannel: getEnv("EVENT_ANNOUNCE_CHANNEL", ""),
		AutoEventsEnabled:   autoEventsEnabled,
		AutoEventInterval:   autoEventInterval,
		AutoEventDuration:   autoEventDuration,
	}, nil
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func parseInt(s string) int {
	val, err := strconv.Atoi(s)
	if err != nil {
		return 0
	}
	return val
}

