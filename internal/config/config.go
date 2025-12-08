package config

import (
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	DiscordToken string
	DBPath       string
	BotPrefix    string
}

func Load() (*Config, error) {
	// Try to load .env file, but don't fail if it doesn't exist
	_ = godotenv.Load()

	return &Config{
		DiscordToken: getEnv("DISCORD_TOKEN", ""),
		DBPath:       getEnv("DB_PATH", "./gophermon.db"),
		BotPrefix:    getEnv("BOT_PREFIX", "!"),
	}, nil
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

