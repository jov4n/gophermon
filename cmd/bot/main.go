package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"

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

	// Initialize gopherkon generator
	log.Println("Initializing sprite generator...")
	assetsPath := "assets/gopherkon"
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

	// Initialize handlers
	handlers := discord.NewHandlers(
		gameService,
		trainerRepo,
		gopherRepo,
		partyRepo,
		battleRepo,
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

	log.Println("Bot is running. Press CTRL-C to exit.")

	// Wait for interrupt signal
	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt)
	<-sc

	log.Println("Shutting down...")
}

