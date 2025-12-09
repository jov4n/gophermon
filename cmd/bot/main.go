package main

import (
	"fmt"
	"log"
	"math/rand"
	"os"
	"os/signal"
	"syscall"
	"time"

	"gophermon-bot/internal/config"
	"gophermon-bot/internal/discord"
	"gophermon-bot/internal/game"
	"gophermon-bot/internal/gopherkon"
	"gophermon-bot/internal/storage"

	"github.com/bwmarrin/discordgo"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Error loading config: %v", err)
	}

	if cfg.DiscordToken == "" {
		log.Fatal("DISCORD_TOKEN is required. Please set it in your .env file.")
	}

	// Create Discord session
	dg, err := discordgo.New("Bot " + cfg.DiscordToken)
	if err != nil {
		log.Fatalf("Error creating Discord session: %v", err)
	}

	// Set intents
	dg.Identify.Intents = discordgo.IntentsGuildMessages | discordgo.IntentsGuilds | discordgo.IntentsMessageContent

	// Initialize database
	log.Println("Connecting to database...")
	db, err := storage.NewDB(cfg.DBPath)
	if err != nil {
		log.Fatalf("Error connecting to database: %v", err)
	}
	defer db.Close()
	log.Println("Database connected successfully")

	// Initialize repositories
	trainerRepo := storage.NewTrainerRepo(db)
	gopherRepo := storage.NewGopherRepo(db)
	partyRepo := storage.NewPartyRepo(db, gopherRepo, trainerRepo)
	battleRepo := storage.NewBattleRepo(db)
	itemRepo := storage.NewItemRepo(db)

	// Initialize gopherkon generator (now uses gopherize.me artwork structure)
	log.Println("Initializing sprite generator...")
	assetsPath := "assets/artwork"
	generator, err := gopherkon.NewGenerator(assetsPath)
	if err != nil {
		log.Printf("Warning: Could not load gopherkon assets: %v. Sprite generation may be limited.", err)
		// Create a generator anyway - it will work with empty assets
		generator, _ = gopherkon.NewGenerator(assetsPath)
	}

	// Initialize evolution service
	evolutionService := game.NewEvolutionService(generator, assetsPath)

	// Initialize game service
	gameService := game.NewService(
		trainerRepo,
		gopherRepo,
		partyRepo,
		battleRepo,
		generator,
		evolutionService,
		assetsPath,
	)

	// Set event manager for evolution service
	eventManager := gameService.GetEventManager()
	evolutionService.SetEventManager(eventManager)

	// Set event announcement channel if configured
	if cfg.EventAnnounceChannel != "" {
		eventManager.SetAnnouncementChannel(cfg.EventAnnounceChannel)
		log.Printf("Event announcements will be sent to channel: %s", cfg.EventAnnounceChannel)
	}

	// Initialize handlers
	handlers := discord.NewHandlers(
		gameService,
		trainerRepo,
		gopherRepo,
		partyRepo,
		battleRepo,
		itemRepo,
	)

	// Register event handlers
	dg.AddHandler(func(s *discordgo.Session, r *discordgo.Ready) {
		log.Printf("Bot is ready! Logged in as: %v#%v", s.State.User.Username, s.State.User.Discriminator)
	})

	dg.AddHandler(func(s *discordgo.Session, i *discordgo.InteractionCreate) {
		handlers.HandleInteraction(s, i)
	})

	// Open connection
	err = dg.Open()
	if err != nil {
		log.Fatalf("Error opening connection: %v", err)
	}
	defer dg.Close()

	// Register commands (global commands - can take up to 1 hour to propagate)
	// For faster testing, use guild commands by passing a guild ID
	log.Println("Registering commands...")
	
	// For development, you can use a specific guild ID for instant command updates
	// Replace with your test server's guild ID, or use "" for global commands
	guildID := os.Getenv("GUILD_ID") // Optional: set in .env for faster command updates
	
	if guildID != "" {
		err = discord.RegisterCommands(dg, guildID)
		if err != nil {
			log.Printf("Error registering guild commands: %v", err)
		} else {
			log.Printf("Commands registered for guild: %s", guildID)
		}
	} else {
		// Global commands (slower to update)
		err = discord.RegisterCommands(dg, "")
		if err != nil {
			log.Printf("Error registering global commands: %v", err)
		} else {
			log.Println("Global commands registered (may take up to 1 hour to propagate)")
		}
	}

	// Start automatic event scheduler if enabled
	if cfg.AutoEventsEnabled {
		log.Printf("Automatic event scheduling enabled (interval: %d hours, duration: %d hours)", 
			cfg.AutoEventInterval, cfg.AutoEventDuration)
		go startEventScheduler(dg, eventManager, cfg.AutoEventInterval, cfg.AutoEventDuration)
	} else {
		log.Println("Automatic event scheduling is disabled")
	}

	log.Println("Bot is running. Press CTRL-C to exit.")

	// Wait for interrupt signal
	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt)
	<-sc

	log.Println("Shutting down...")
}

// startEventScheduler runs in the background and automatically starts random events
func startEventScheduler(s *discordgo.Session, eventManager *game.EventManager, intervalHours, durationHours int) {
	// Wait a bit before starting first event (let bot fully initialize)
	initialDelay := time.Duration(rand.Intn(30)+10) * time.Minute // 10-40 minutes
	log.Printf("Event scheduler will start first event in %v", initialDelay)
	time.Sleep(initialDelay)

	interval := time.Duration(intervalHours) * time.Hour
	duration := time.Duration(durationHours) * time.Hour

	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	// Start first event immediately
	startAutoEvent(s, eventManager, duration)

	// Then schedule events at regular intervals
	for range ticker.C {
		startAutoEvent(s, eventManager, duration)
	}
}

// startAutoEvent starts a random event and announces it in Discord
func startAutoEvent(s *discordgo.Session, eventManager *game.EventManager, duration time.Duration) {
	// Check if there are already too many active events (max 2 at once)
	activeEvents := eventManager.GetActiveEvents()
	if len(activeEvents) >= 2 {
		log.Printf("Skipping auto event - already %d active events", len(activeEvents))
		return
	}

	// Start random event
	event := eventManager.StartRandomEvent(duration)
	hours := int(duration.Hours())

	log.Printf("Auto-started event: %s (duration: %d hours)", event.Name, hours)

	// Announce in Discord
	embed := &discordgo.MessageEmbed{
		Title:       fmt.Sprintf("🎉 %s Started! 🎉", event.Name),
		Description: fmt.Sprintf("%s\n\n⏰ Duration: %d hours\n\nThis event was automatically started! Get out there and enjoy!", 
			event.Description, hours),
		Color:       0xffd700, // Gold color
		Timestamp:   time.Now().Format(time.RFC3339),
		Footer: &discordgo.MessageEmbedFooter{
			Text: "🤖 Auto-scheduled event",
		},
	}

	// Try to send to announcement channel first
	channelID := eventManager.GetAnnouncementChannel()
	if channelID != "" {
		_, err := s.ChannelMessageSendEmbed(channelID, embed)
		if err != nil {
			log.Printf("Error sending event announcement to channel %s: %v", channelID, err)
		}
	} else {
		// If no announcement channel, log it
		log.Printf("Event started but no announcement channel configured")
	}
}

