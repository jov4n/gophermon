package discord

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
	"time"

	"gophermon-bot/internal/game"
	"gophermon-bot/internal/storage"

	"github.com/bwmarrin/discordgo"
	"github.com/google/uuid"
)

type Handlers struct {
	gameService *game.Service
	trainerRepo *storage.TrainerRepo
	gopherRepo  *storage.GopherRepo
	partyRepo   *storage.PartyRepo
	battleRepo  *storage.BattleRepo
	battles     map[string]*game.BattleState // In-memory battle cache
	starterSessions map[string][]string      // Session ID -> starter gopher IDs
}

func NewHandlers(
	gameService *game.Service,
	trainerRepo *storage.TrainerRepo,
	gopherRepo *storage.GopherRepo,
	partyRepo *storage.PartyRepo,
	battleRepo *storage.BattleRepo,
) *Handlers {
	return &Handlers{
		gameService: gameService,
		trainerRepo: trainerRepo,
		gopherRepo:  gopherRepo,
		partyRepo:   partyRepo,
		battleRepo:  battleRepo,
		battles:          make(map[string]*game.BattleState),
		starterSessions:  make(map[string][]string),
	}
}

func (h *Handlers) HandleInteraction(s *discordgo.Session, i *discordgo.InteractionCreate) {
	switch i.Type {
	case discordgo.InteractionApplicationCommand:
		h.handleCommand(s, i)
	case discordgo.InteractionMessageComponent:
		h.handleComponent(s, i)
	}
}

func (h *Handlers) handleCommand(s *discordgo.Session, i *discordgo.InteractionCreate) {
	data := i.ApplicationCommandData()

	switch data.Name {
	case "ping":
		h.handlePing(s, i)
	case "start":
		h.handleStart(s, i)
	case "choose":
		h.handleChoose(s, i)
	case "party":
		h.handleParty(s, i)
	case "pc":
		h.handlePC(s, i)
	case "wild":
		h.handleWild(s, i)
	case "gopher":
		h.handleGopher(s, i)
	case "generate_10":
		h.handleGenerate10(s, i)
	default:
		respondEphemeral(s, i, "Unknown command")
	}
}

func (h *Handlers) handleComponent(s *discordgo.Session, i *discordgo.InteractionCreate) {
	data := i.MessageComponentData()

	if strings.HasPrefix(data.CustomID, "battle_") {
		if strings.HasPrefix(data.CustomID, "battle_ability_") {
			h.handleBattleAbility(s, i)
		} else if strings.HasPrefix(data.CustomID, "battle_swap_") {
			h.handleBattleSwap(s, i)
		} else {
			h.handleBattleAction(s, i)
		}
	} else if strings.HasPrefix(data.CustomID, "choose_") {
		h.handleChooseStarter(s, i)
	} else {
		respondEphemeral(s, i, "Unknown action")
	}
}

func (h *Handlers) handlePing(s *discordgo.Session, i *discordgo.InteractionCreate) {
	respondEphemeral(s, i, "Pong! Bot is alive.")
}

func (h *Handlers) handleStart(s *discordgo.Session, i *discordgo.InteractionCreate) {
	discordID := i.Member.User.ID

	// Check if trainer already exists
	trainer, err := h.trainerRepo.GetByDiscordID(discordID)
	if err != nil {
		respondEphemeral(s, i, fmt.Sprintf("Error: %v", err))
		return
	}

	// If trainer exists, check if they have gophers in party
	if trainer != nil {
		party, err := h.gopherRepo.GetParty(trainer.ID)
		if err == nil && len(party) > 0 {
			respondEphemeral(s, i, "You already have a starter gopher! Use /party to view your gophers.")
			return
		}
		// Trainer exists but no gophers - allow them to get starters
	}

	// Get user's display name
	username := i.Member.User.Username
	if i.Member.Nick != "" {
		username = i.Member.Nick
	}

	// Create trainer if they don't exist
	if trainer == nil {
		trainer, err = h.trainerRepo.Create(discordID, username)
		if err != nil {
			respondEphemeral(s, i, fmt.Sprintf("Error creating trainer: %v", err))
			return
		}
	}

	// Generate 3 starter gophers
	starters, err := h.gameService.GenerateStarterGophers()
	if err != nil {
		respondEphemeral(s, i, fmt.Sprintf("Error generating starters: %v", err))
		return
	}

	// Store starters temporarily (we'll delete the unchosen ones)
	starterIDs := []string{}
	for _, starter := range starters {
		created, err := h.gopherRepo.Create(starter)
		if err != nil {
			respondEphemeral(s, i, fmt.Sprintf("Error saving starter: %v", err))
			return
		}
		starterIDs = append(starterIDs, created.ID)
	}

	// Generate starter card with all 3 gophers (returns base64)
	cardBase64, err := h.gameService.GenerateStarterCard(starters)
	var cardFile *discordgo.File
	var imageURL string
	
	if err == nil && cardBase64 != "" {
		// Decode base64 to bytes
		if fileData, err := base64.StdEncoding.DecodeString(cardBase64); err == nil {
			fileName := "starter_card.png"
			cardFile = &discordgo.File{
				Name:        fileName,
				ContentType: "image/png",
				Reader:      bytes.NewReader(fileData),
			}
			imageURL = fmt.Sprintf("attachment://%s", fileName)
		}
	}

	// Create embed with starter options
	embed := &discordgo.MessageEmbed{
		Title:       "Choose Your Starter Gopher!",
		Description: "Select one of the three starter gophers below:",
		Color:       0x00ff00,
		Fields:      []*discordgo.MessageEmbedField{},
	}

	// Add card image if we have it
	if imageURL != "" {
		embed.Image = &discordgo.MessageEmbedImage{
			URL: imageURL,
		}
	}

	// Add stats for each starter
	for idx, starter := range starters {
		hpBar := game.GetHPBar(starter.CurrentHP, starter.MaxHP, 10)
		embed.Fields = append(embed.Fields, &discordgo.MessageEmbedField{
			Name: fmt.Sprintf("Starter %d: %s", idx+1, starter.Name),
			Value: fmt.Sprintf("**Archetype:** %s\n**Level:** %d\n**HP:** %s\n**Stats:** ATK:%d DEF:%d SPD:%d\n**Rarity:** %s",
				starter.SpeciesArchetype, starter.Level, hpBar, starter.Attack, starter.Defense, starter.Speed, starter.Rarity),
			Inline: true,
		})
	}

	// Create a short session ID to store starter IDs and card path
	sessionID := fmt.Sprintf("%d", time.Now().UnixNano()%1000000) // 6-7 digit number
	h.starterSessions[sessionID] = starterIDs
	
	// Store card path in session for cleanup (we'll store it with a prefix in the session map)
	// Actually, let's store it separately or append to starterIDs with a marker
	// For now, we'll delete it in handleChooseStarter

	// Create buttons with short custom IDs
	components := []discordgo.MessageComponent{
		discordgo.ActionsRow{
			Components: []discordgo.MessageComponent{
				createButton("Choose Starter 1", discordgo.PrimaryButton, fmt.Sprintf("choose_%s_1", sessionID)),
				createButton("Choose Starter 2", discordgo.PrimaryButton, fmt.Sprintf("choose_%s_2", sessionID)),
				createButton("Choose Starter 3", discordgo.PrimaryButton, fmt.Sprintf("choose_%s_3", sessionID)),
			},
		},
	}

	// Send with card file if we have it
	if cardFile != nil {
		err := s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Embeds:     []*discordgo.MessageEmbed{embed},
				Components: components,
				Files:     []*discordgo.File{cardFile},
			},
		})
		if err != nil {
			log.Printf("Error responding with card: %v", err)
			// Fallback to without card
			respondWithComponents(s, i, "", embed, components, false)
		}
	} else {
		respondWithComponents(s, i, "", embed, components, false)
	}
}

func (h *Handlers) handleChooseStarter(s *discordgo.Session, i *discordgo.InteractionCreate) {
	data := i.MessageComponentData()
	// Format: choose_<sessionID>_<index> where index is 1, 2, or 3
	parts := strings.SplitN(data.CustomID, "_", 3)
	if len(parts) < 3 {
		respondEphemeral(s, i, "Invalid choice")
		return
	}

	sessionID := parts[1]
	starterIndex, err := strconv.Atoi(parts[2])
	if err != nil || starterIndex < 1 || starterIndex > 3 {
		respondEphemeral(s, i, "Invalid starter index")
		return
	}

	// Get starter IDs from session
	starterIDs, exists := h.starterSessions[sessionID]
	if !exists || len(starterIDs) < starterIndex {
		respondEphemeral(s, i, "Starter session expired or invalid. Please use /start again.")
		return
	}

	chosenID := starterIDs[starterIndex-1] // Convert 1-based to 0-based
	discordID := i.Member.User.ID

	trainer, err := h.trainerRepo.GetByDiscordID(discordID)
	if err != nil || trainer == nil {
		respondEphemeral(s, i, "Trainer not found. Use /start first.")
		return
	}

	// Get the chosen gopher
	chosenGopher, err := h.gopherRepo.GetByID(chosenID)
	if err != nil || chosenGopher == nil {
		respondEphemeral(s, i, "Gopher not found")
		return
	}

	// Delete the other starter gophers (sprites are stored as base64 in database, no file cleanup needed)
	for _, starterID := range starterIDs {
		if starterID == chosenID {
			continue // Skip the chosen one
		}
		
		// Delete from database
		h.gopherRepo.Delete(starterID)
	}

	// Clean up session
	delete(h.starterSessions, sessionID)

	// Assign chosen gopher to trainer and add to party
	chosenGopher.TrainerID = &trainer.ID
	chosenGopher.IsInParty = true
	chosenGopher.PCSlot = nil // Not in PC
	if err := h.gopherRepo.Update(chosenGopher); err != nil {
		respondEphemeral(s, i, fmt.Sprintf("Error assigning gopher to trainer: %v", err))
		return
	}

	// Update trainer's party slot count
	partySize, err := h.partyRepo.GetPartySize(trainer.ID)
	if err != nil {
		log.Printf("Error getting party size: %v", err)
		// Continue anyway - party size update is not critical
	} else {
		if err := h.trainerRepo.UpdatePartySlots(trainer.ID, partySize); err != nil {
			log.Printf("Error updating party slots: %v", err)
			// Continue anyway - party size update is not critical
		}
	}

	// Load chosen gopher's sprite for the embed
	var chosenGopherFile *discordgo.File
	var chosenImageURL string
	imageData, err := h.getGopherImageData(chosenGopher)
	if err == nil && imageData != nil {
		fileName := fmt.Sprintf("chosen_%s.png", chosenGopher.ID[:8])
		chosenGopherFile = &discordgo.File{
			Name:        fileName,
			ContentType: "image/png",
			Reader:      bytes.NewReader(imageData),
		}
		chosenImageURL = fmt.Sprintf("attachment://%s", fileName)
	}

	// Create embed without the card image - only show chosen gopher
	// Make sure Image is nil/not set to remove the old card image
	embed := &discordgo.MessageEmbed{
		Title:       "Starter Chosen!",
		Description: fmt.Sprintf("You chose **%s**! Your journey begins now!", chosenGopher.Name),
		Color:       0x00ff00,
		Fields: []*discordgo.MessageEmbedField{
			{
				Name:   "Your Starter",
				Value:  fmt.Sprintf("**%s** - Level %d %s", chosenGopher.Name, chosenGopher.Level, chosenGopher.SpeciesArchetype),
				Inline: false,
			},
		},
		Image: nil, // Explicitly set to nil to remove old image
	}

	// Add chosen gopher's sprite as thumbnail if available
	if chosenImageURL != "" {
		embed.Thumbnail = &discordgo.MessageEmbedThumbnail{
			URL: chosenImageURL,
		}
	} else {
		// If no sprite, at least show stats
		hpBar := game.GetHPBar(chosenGopher.CurrentHP, chosenGopher.MaxHP, 10)
		embed.Fields = append(embed.Fields, &discordgo.MessageEmbedField{
			Name:   "Stats",
			Value:  fmt.Sprintf("**HP:** %s\n**ATK:** %d | **DEF:** %d | **SPD:** %d\n**Rarity:** %s", hpBar, chosenGopher.Attack, chosenGopher.Defense, chosenGopher.Speed, chosenGopher.Rarity),
			Inline: false,
		})
	}

	// Respond to the interaction first (required within 3 seconds)
	// Use UpdateMessage to replace the embed and remove buttons
	err = s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseUpdateMessage,
		Data: &discordgo.InteractionResponseData{
			Embeds:     []*discordgo.MessageEmbed{embed},
			Components: []discordgo.MessageComponent{}, // Remove all buttons
		},
	})
	if err != nil {
		log.Printf("Error responding to interaction: %v", err)
		respondEphemeral(s, i, fmt.Sprintf("Error: %v", err))
		return
	}

	// Delete the original message to remove the card attachment, then send a new one
	// Discord doesn't allow removing attachments added via interaction responses
	err = s.ChannelMessageDelete(i.ChannelID, i.Message.ID)
	if err != nil {
		log.Printf("Error deleting original message: %v", err)
		// Continue anyway - we'll try to edit it
	}

	// Send a new message with just the chosen gopher
	webhookParams := &discordgo.WebhookParams{
		Embeds: []*discordgo.MessageEmbed{embed},
	}
	
	if chosenGopherFile != nil {
		webhookParams.Files = []*discordgo.File{chosenGopherFile}
	}
	
	_, err = s.FollowupMessageCreate(i.Interaction, false, webhookParams)
	if err != nil {
		log.Printf("Error sending followup message: %v", err)
	}

	// Send a followup message (we can't use respondEphemeral after InteractionResponseUpdateMessage)
	_, err = s.FollowupMessageCreate(i.Interaction, true, &discordgo.WebhookParams{
		Content: "Welcome to Gophermon! Use /wild to encounter gophers, /party to view your team.",
		Flags:   discordgo.MessageFlagsEphemeral,
	})
	if err != nil {
		log.Printf("Error sending followup message: %v", err)
	}
}

func (h *Handlers) handleChoose(s *discordgo.Session, i *discordgo.InteractionCreate) {
	// This command is handled via buttons in /start
	respondEphemeral(s, i, "Use /start to choose a starter gopher!")
}

func (h *Handlers) handleParty(s *discordgo.Session, i *discordgo.InteractionCreate) {
	discordID := i.Member.User.ID

	trainer, err := h.trainerRepo.GetByDiscordID(discordID)
	if err != nil || trainer == nil {
		respondEphemeral(s, i, "Trainer not found. Use /start first.")
		return
	}

	party, err := h.gopherRepo.GetParty(trainer.ID)
	if err != nil {
		respondEphemeral(s, i, fmt.Sprintf("Error: %v", err))
		return
	}

	if len(party) == 0 {
		respondEphemeral(s, i, "Your party is empty! Use /start to get a starter gopher.")
		return
	}

	embed := &discordgo.MessageEmbed{
		Title:       fmt.Sprintf("%s's Party", trainer.Name),
		Description: fmt.Sprintf("Active Party (%d/6)", len(party)),
		Color:       0x0099ff,
		Fields:      []*discordgo.MessageEmbedField{},
	}

	for _, gopher := range party {
		hpBar := game.GetHPBar(gopher.CurrentHP, gopher.MaxHP, 10)
		xpBar := game.GetXPBar(gopher.XP, gopher.Level, 10)
		embed.Fields = append(embed.Fields, &discordgo.MessageEmbedField{
			Name: fmt.Sprintf("%s (ID: %s)", gopher.Name, gopher.ID[:8]),
			Value: fmt.Sprintf("**Level:** %d\n**XP:** %s\n**HP:** %s\n**Stats:** ATK:%d DEF:%d SPD:%d\n**Type:** %s | **Rarity:** %s",
				gopher.Level, xpBar, hpBar, gopher.Attack, gopher.Defense, gopher.Speed, gopher.SpeciesArchetype, gopher.Rarity),
			Inline: false,
		})
	}

	respondEmbed(s, i, embed, true)
}

func (h *Handlers) handlePC(s *discordgo.Session, i *discordgo.InteractionCreate) {
	data := i.ApplicationCommandData()
	discordID := i.Member.User.ID

	trainer, err := h.trainerRepo.GetByDiscordID(discordID)
	if err != nil || trainer == nil {
		respondEphemeral(s, i, "Trainer not found. Use /start first.")
		return
	}

	if len(data.Options) == 0 {
		// List PC gophers
		page := 1
		limit := 10
		offset := (page - 1) * limit

		pcGophers, err := h.gopherRepo.GetPC(trainer.ID, limit, offset)
		if err != nil {
			respondEphemeral(s, i, fmt.Sprintf("Error: %v", err))
			return
		}

		total, err := h.gopherRepo.CountPC(trainer.ID)
		if err != nil {
			total = len(pcGophers)
		}

		if len(pcGophers) == 0 {
			respondEphemeral(s, i, "Your PC is empty!")
			return
		}

		embed := &discordgo.MessageEmbed{
			Title:       "PC Storage",
			Description: fmt.Sprintf("Stored Gophers (%d total)", total),
			Color:       0x9966ff,
			Fields:      []*discordgo.MessageEmbedField{},
		}

		for _, gopher := range pcGophers {
			hpBar := game.GetHPBar(gopher.CurrentHP, gopher.MaxHP, 8)
			embed.Fields = append(embed.Fields, &discordgo.MessageEmbedField{
				Name: fmt.Sprintf("%s (ID: %s)", gopher.Name, gopher.ID[:8]),
				Value: fmt.Sprintf("Lv.%d | %s | %s | %s", gopher.Level, hpBar, gopher.SpeciesArchetype, gopher.Rarity),
				Inline: true,
			})
		}

		respondEmbed(s, i, embed, true)
		return
	}

	subCommand := data.Options[0]
	switch subCommand.Name {
	case "deposit":
		gopherID := subCommand.Options[0].StringValue()
		if err := h.partyRepo.RemoveFromParty(trainer.ID, gopherID); err != nil {
			respondEphemeral(s, i, fmt.Sprintf("Error: %v", err))
			return
		}
		respondEphemeral(s, i, "Gopher deposited to PC!")

	case "withdraw":
		gopherID := subCommand.Options[0].StringValue()
		if err := h.partyRepo.AddToParty(trainer.ID, gopherID); err != nil {
			respondEphemeral(s, i, fmt.Sprintf("Error: %v", err))
			return
		}
		respondEphemeral(s, i, "Gopher withdrawn from PC!")

	default:
		respondEphemeral(s, i, "Unknown PC command")
	}
}

func (h *Handlers) handleWild(s *discordgo.Session, i *discordgo.InteractionCreate) {
	discordID := i.Member.User.ID

	trainer, err := h.trainerRepo.GetByDiscordID(discordID)
	if err != nil || trainer == nil {
		respondEphemeral(s, i, "Trainer not found. Use /start first.")
		return
	}

	// Get player's first party gopher
	party, err := h.gopherRepo.GetParty(trainer.ID)
	if err != nil || len(party) == 0 {
		respondEphemeral(s, i, "Your party is empty! Use /start to get a starter gopher.")
		return
	}

	// Check if all party members are dead - trigger blackout if so
	blackedOut, blackoutMsg := h.gameService.CheckAndHandleBlackout(trainer.ID)
	if blackedOut {
		respondEphemeral(s, i, blackoutMsg)
		return
	}

	// Find first alive gopher in party
	var playerGopherStorage *storage.Gopher
	for _, gopher := range party {
		if gopher.CurrentHP > 0 {
			playerGopherStorage = gopher
			break
		}
	}

	// If no alive gopher found (shouldn't happen after blackout check, but safety check)
	if playerGopherStorage == nil {
		respondEphemeral(s, i, "All your gophers are fainted! You need to rest.")
		return
	}

	// Generate wild gopher
	wildGopherStorage, err := h.gameService.GenerateWildGopher()
	if err != nil {
		respondEphemeral(s, i, fmt.Sprintf("Error generating wild gopher: %v", err))
		return
	}

	// Save wild gopher (without trainer_id)
	wildGopherStorage, err = h.gopherRepo.Create(wildGopherStorage)
	if err != nil {
		respondEphemeral(s, i, fmt.Sprintf("Error saving wild gopher: %v", err))
		return
	}

	// Convert to game gophers
	playerGopher, err := h.gameService.StorageGopherToGameGopher(playerGopherStorage)
	if err != nil {
		respondEphemeral(s, i, fmt.Sprintf("Error: %v", err))
		return
	}

	enemyGopher, err := h.gameService.StorageGopherToGameGopher(wildGopherStorage)
	if err != nil {
		respondEphemeral(s, i, fmt.Sprintf("Error: %v", err))
		return
	}

	// Get full party for battle (for swapping) - reuse party variable from earlier
	partyStorage, err := h.gopherRepo.GetParty(trainer.ID)
	if err != nil {
		respondEphemeral(s, i, fmt.Sprintf("Error getting party: %v", err))
		return
	}
	
	// Convert party to game gophers
	gameParty := make([]*game.Gopher, len(partyStorage))
	for idx, gopherStorage := range partyStorage {
		gameGopher, err := h.gameService.StorageGopherToGameGopher(gopherStorage)
		if err != nil {
			respondEphemeral(s, i, fmt.Sprintf("Error converting gopher: %v", err))
			return
		}
		gameParty[idx] = gameGopher
	}

	// Create battle state with party
	battleState := game.NewBattleState(trainer.ID, i.ChannelID, playerGopher, enemyGopher, gameParty)
	battleState.ID = uuid.New().String()

	// Create battle embed
	embed := h.createBattleEmbed(battleState)

	// Create battle buttons
	components := h.createBattleButtons(battleState, false)

	// Generate battle card with both gophers (enemy on top, player on bottom) - returns base64
	var battleCardFile *discordgo.File
	var battleImageURL string
	if (playerGopherStorage.SpriteData != "" || playerGopherStorage.SpritePath != "") && 
	   (wildGopherStorage.SpriteData != "" || wildGopherStorage.SpritePath != "") {
		cardBase64, err := h.gameService.GenerateBattleCard(wildGopherStorage, playerGopherStorage)
		if err != nil {
			log.Printf("Error generating battle card: %v", err)
		} else if cardBase64 != "" {
			// Decode base64 to bytes
			if fileData, err := base64.StdEncoding.DecodeString(cardBase64); err != nil {
				log.Printf("Error decoding battle card base64: %v", err)
			} else {
				fileName := fmt.Sprintf("battle_%s.png", battleState.ID[:8])
				battleCardFile = &discordgo.File{
					Name:        fileName,
					ContentType: "image/png",
					Reader:      bytes.NewReader(fileData),
				}
				battleImageURL = fmt.Sprintf("attachment://%s", fileName)
				// Add image to embed
				embed.Image = &discordgo.MessageEmbedImage{
					URL: battleImageURL,
				}
			}
		}
	}

	// Send battle message with battle card if available
	msgSend := &discordgo.MessageSend{
		Embeds:     []*discordgo.MessageEmbed{embed},
		Components: components,
	}
	if battleCardFile != nil {
		msgSend.Files = []*discordgo.File{battleCardFile}
	}

	msg, err := s.ChannelMessageSendComplex(i.ChannelID, msgSend)
	if err != nil {
		respondEphemeral(s, i, fmt.Sprintf("Error sending battle message: %v", err))
		return
	}

	battleState.MessageID = msg.ID

	// Save battle to DB
	battle := &storage.Battle{
		ID:            battleState.ID,
		ChannelID:     battleState.ChannelID,
		MessageID:     battleState.MessageID,
		TrainerID:     battleState.TrainerID,
		OpponentType:  "WILD",
		GopherIDPlayer: &playerGopher.ID,
		GopherIDEnemy:  &enemyGopher.ID,
		TurnOwner:     battleState.TurnOwner,
		State:         battleState.State,
	}
	_, err = h.battleRepo.Create(battle)
	if err != nil {
		log.Printf("Error saving battle: %v", err)
	}

	// Store in memory
	h.battles[battleState.ID] = battleState

	respondEphemeral(s, i, "Wild gopher encountered!")
}

func (h *Handlers) handleGopher(s *discordgo.Session, i *discordgo.InteractionCreate) {
	data := i.ApplicationCommandData()
	discordID := i.Member.User.ID

	trainer, err := h.trainerRepo.GetByDiscordID(discordID)
	if err != nil || trainer == nil {
		respondEphemeral(s, i, "Trainer not found. Use /start first.")
		return
	}

	if len(data.Options) == 0 || data.Options[0].Name != "info" {
		respondEphemeral(s, i, "Use /gopher info <gopher_id>")
		return
	}

	gopherID := data.Options[0].Options[0].StringValue()
	gopher, err := h.gopherRepo.GetByID(gopherID)
	if err != nil || gopher == nil {
		respondEphemeral(s, i, "Gopher not found")
		return
	}

	// Check ownership
	if gopher.TrainerID == nil || *gopher.TrainerID != trainer.ID {
		respondEphemeral(s, i, "This gopher doesn't belong to you")
		return
	}

	gameGopher, err := h.gameService.StorageGopherToGameGopher(gopher)
	if err != nil {
		respondEphemeral(s, i, fmt.Sprintf("Error: %v", err))
		return
	}

	hpBar := game.GetHPBar(gopher.CurrentHP, gopher.MaxHP, 15)
	xpBar := game.GetXPBar(gopher.XP, gopher.Level, 15)

	embed := &discordgo.MessageEmbed{
		Title:       gopher.Name,
		Description: fmt.Sprintf("**ID:** %s", gopher.ID),
		Color:       0xff9900,
		Fields: []*discordgo.MessageEmbedField{
			{
				Name:   "Level & XP",
				Value:  fmt.Sprintf("Level **%d**\n**XP:** %s", gopher.Level, xpBar),
				Inline: false,
			},
			{
				Name:   "HP",
				Value:  hpBar,
				Inline: false,
			},
			{
				Name:   "Stats",
				Value:  fmt.Sprintf("**Attack:** %d\n**Defense:** %d\n**Speed:** %d", gopher.Attack, gopher.Defense, gopher.Speed),
				Inline: true,
			},
			{
				Name:   "Info",
				Value:  fmt.Sprintf("**Type:** %s\n**Rarity:** %s\n**Evolution Stage:** %d", gopher.SpeciesArchetype, gopher.Rarity, gopher.EvolutionStage),
				Inline: true,
			},
		},
	}

	// Add abilities
	if len(gameGopher.Abilities) > 0 {
		abilityList := ""
		for idx, ability := range gameGopher.Abilities {
			abilityList += fmt.Sprintf("%d. **%s** - %s (Power: %d)\n", idx+1, ability.Name, ability.Description, ability.Power)
		}
		embed.Fields = append(embed.Fields, &discordgo.MessageEmbedField{
			Name:   "Abilities",
			Value:  abilityList,
			Inline: false,
		})
	}

	respondEmbed(s, i, embed, true)
}

func (h *Handlers) handleGenerate10(s *discordgo.Session, i *discordgo.InteractionCreate) {
	// Respond immediately to avoid timeout (Discord requires response within 3 seconds)
	err := s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseDeferredChannelMessageWithSource,
	})
	if err != nil {
		log.Printf("Error deferring interaction: %v", err)
		return
	}

	// Do the heavy work in a goroutine so we don't block
	go func() {
		// Generate 3 gophers of each rarity (15 total)
		rarities := []game.Rarity{
			game.RarityCommon,
			game.RarityUncommon,
			game.RarityRare,
			game.RarityEpic,
			game.RarityLegendary,
		}

		var gophers []*storage.Gopher
		seedOffset := int64(0)

		// Generate 3 of each rarity
		for _, rarity := range rarities {
			for j := 0; j < 3; j++ {
				gopher, err := h.gameService.GenerateGopherWithRarity(rarity, seedOffset)
				if err != nil {
					s.FollowupMessageCreate(i.Interaction, true, &discordgo.WebhookParams{
						Content: fmt.Sprintf("Error generating %s gopher %d: %v", rarity, j+1, err),
						Flags:   discordgo.MessageFlagsEphemeral,
					})
					return
				}
				gophers = append(gophers, gopher)
				seedOffset++
			}
		}

		// Generate card with 15 gophers in a 5x3 grid (5 columns, 3 rows) - returns base64
		cardBase64, err := h.gameService.GenerateGopherCard(gophers, 5)
		if err != nil {
			s.FollowupMessageCreate(i.Interaction, true, &discordgo.WebhookParams{
				Content: fmt.Sprintf("Error generating card: %v", err),
				Flags:   discordgo.MessageFlagsEphemeral,
			})
			return
		}

		// Decode base64 to bytes
		var cardFile *discordgo.File
		var imageURL string
		if fileData, err := base64.StdEncoding.DecodeString(cardBase64); err == nil {
			fileName := fmt.Sprintf("gopher_card_15_%d.png", time.Now().Unix())
			cardFile = &discordgo.File{
				Name:        fileName,
				ContentType: "image/png",
				Reader:      bytes.NewReader(fileData),
			}
			imageURL = fmt.Sprintf("attachment://%s", fileName)
		} else {
			s.FollowupMessageCreate(i.Interaction, true, &discordgo.WebhookParams{
				Content: fmt.Sprintf("Error decoding card data: %v", err),
				Flags:   discordgo.MessageFlagsEphemeral,
			})
			return
		}

		// Create embed
		embed := &discordgo.MessageEmbed{
			Title:       "Generated 15 Gophers (3 of Each Rarity)",
			Description: "3 COMMON | 3 UNCOMMON | 3 RARE | 3 EPIC | 3 LEGENDARY\n\nArranged in a 5x3 grid showing all rarity tiers!",
			Color:       0x00ff00,
			Image: &discordgo.MessageEmbedImage{
				URL: imageURL,
			},
		}

		// Send followup message with card
		_, err = s.FollowupMessageCreate(i.Interaction, false, &discordgo.WebhookParams{
			Embeds: []*discordgo.MessageEmbed{embed},
			Files:  []*discordgo.File{cardFile},
		})
		if err != nil {
			log.Printf("Error sending followup message with card: %v", err)
			return
		}

		// No cleanup needed - sprites are stored as base64 in database
	}()
}

func (h *Handlers) handleBattleAction(s *discordgo.Session, i *discordgo.InteractionCreate) {
	data := i.MessageComponentData()
	discordID := i.Member.User.ID

	// Find battle by message ID
	battleState := h.findBattleByMessage(i.ChannelID, i.Message.ID)
	if battleState == nil {
		respondEphemeral(s, i, "Battle not found or already ended")
		return
	}

	// Verify ownership - get trainer from Discord ID and compare
	trainer, err := h.trainerRepo.GetByDiscordID(discordID)
	if err != nil || trainer == nil {
		respondEphemeral(s, i, "Trainer not found")
		return
	}
	if battleState.TrainerID != trainer.ID {
		respondEphemeral(s, i, "This isn't your battle!")
		return
	}

	action := strings.TrimPrefix(data.CustomID, "battle_")
	var abilityIndex int = -1

	// Map button actions to battle actions
	if action == "net" {
		action = "throw_net"
	}

	// Acknowledge the interaction first (needed for all actions)
	err = s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseDeferredMessageUpdate,
	})
	if err != nil {
		log.Printf("Error acknowledging interaction: %v", err)
		return
	}

	if action == "swap" {
		// Swap action shows party selection
		h.showPartySwapMenu(s, i, battleState)
		return
	}

	if action == "fight" {
		// Show ability buttons instead
		embed := h.createBattleEmbed(battleState)
		components := h.createBattleButtons(battleState, true)
		h.editBattleMessage(s, i, "", embed, components)
		return
	}

	// Execute action
	messages, err := battleState.PlayerAction(action, abilityIndex)
	if err != nil {
		// Can't use respondEphemeral after deferred update, so edit the message with error
		h.editBattleMessage(s, i, fmt.Sprintf("Error: %v", err), h.createBattleEmbed(battleState), h.createBattleButtons(battleState, false))
		return
	}

	// Update battle state in DB
	battle, _ := h.battleRepo.GetByID(battleState.ID)
	if battle != nil {
		battle.State = battleState.State
		battle.TurnOwner = battleState.TurnOwner
		h.battleRepo.Update(battle)
	}

	// Check for evolution after level up - check all participating gophers
	evolutionMessages := []string{}
	if strings.Contains(strings.Join(messages, " "), "leveled up") {
		// Check evolution for all participating gophers that may have leveled up
		for _, gopher := range battleState.ParticipatingGophers {
			// Check if this gopher is at an evolution threshold
			shouldCheckEvolution := (gopher.Level >= 16 && gopher.EvolutionStage == 0) ||
				(gopher.Level >= 32 && gopher.EvolutionStage == 1)
			
			if shouldCheckEvolution {
				evolved, evolutionMsg := h.gameService.CheckEvolution(gopher)
				if evolved {
					evolutionMessages = append(evolutionMessages, evolutionMsg)
					// Update gopher after evolution
					h.gopherRepo.Update(h.gameGopherToStorage(gopher))
				}
			}
		}
	}

	// Update all participating gophers in DB (they may have gained XP and leveled up)
	// IMPORTANT: Save HP changes BEFORE checking for blackout
	for _, gopher := range battleState.ParticipatingGophers {
		h.gopherRepo.Update(h.gameGopherToStorage(gopher))
	}
	
	// Update active player gopher (in case it wasn't in participating list)
	h.gopherRepo.Update(h.gameGopherToStorage(battleState.PlayerGopher))
	
	// Update all party members in DB to ensure HP is saved
	partyStorage, _ := h.gopherRepo.GetParty(battleState.TrainerID)
	for _, partyGopher := range partyStorage {
		// Find matching gopher in battle state
		for _, battleGopher := range battleState.PlayerParty {
			if battleGopher.ID == partyGopher.ID {
				partyGopher.CurrentHP = battleGopher.CurrentHP
				partyGopher.MaxHP = battleGopher.MaxHP
				h.gopherRepo.Update(partyGopher)
				break
			}
		}
	}
	
	if battleState.State != "WON" && battleState.State != "ESCAPED" {
		h.gopherRepo.Update(h.gameGopherToStorage(battleState.EnemyGopher))
	}

	// Handle battle end
	if battleState.State != "ACTIVE" {
		if battleState.State == "WON" {
			// Check if captured
			if strings.Contains(strings.Join(messages, " "), "captured") {
				// Add to party or PC
				partySize, _ := h.partyRepo.GetPartySize(battleState.TrainerID)
				enemyStorage := h.gameGopherToStorage(battleState.EnemyGopher)
				enemyStorage.TrainerID = &battleState.TrainerID
				enemyStorage.IsInParty = partySize < 6
				h.gopherRepo.Update(enemyStorage)
				if enemyStorage.IsInParty {
					h.trainerRepo.UpdatePartySlots(battleState.TrainerID, partySize+1)
				}
			}
		} else if battleState.State == "LOST" {
			// Check for blackout after battle loss (HP is already saved above)
			blackedOut, blackoutMsg := h.gameService.CheckAndHandleBlackout(battleState.TrainerID)
			if blackedOut {
				messages = append(messages, "")
				messages = append(messages, blackoutMsg)
			}
		}
		delete(h.battles, battleState.ID)
	}

	// Update embed with battle card image (don't regenerate for regular actions)
	embed := h.createBattleEmbed(battleState)
	h.addBattleCardToEmbed(embed, battleState, false) // false = don't regenerate, just reference existing
	
	var components []discordgo.MessageComponent
	if battleState.State == "ACTIVE" {
		components = h.createBattleButtons(battleState, false)
	}
	h.editBattleMessage(s, i, strings.Join(messages, "\n"), embed, components)
}

func (h *Handlers) handleBattleAbility(s *discordgo.Session, i *discordgo.InteractionCreate) {
	data := i.MessageComponentData()
	discordID := i.Member.User.ID

	// Find battle
	battleState := h.findBattleByMessage(i.ChannelID, i.Message.ID)
	if battleState == nil {
		respondEphemeral(s, i, "Battle not found")
		return
	}

	// Verify ownership - get trainer from Discord ID and compare
	trainer, err := h.trainerRepo.GetByDiscordID(discordID)
	if err != nil || trainer == nil {
		respondEphemeral(s, i, "Trainer not found")
		return
	}
	if battleState.TrainerID != trainer.ID {
		respondEphemeral(s, i, "This isn't your battle!")
		return
	}

	// Acknowledge the interaction first
	err = s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseDeferredMessageUpdate,
	})
	if err != nil {
		log.Printf("Error acknowledging interaction: %v", err)
		return
	}

	// Parse ability index
	parts := strings.Split(data.CustomID, "_")
	if len(parts) < 3 {
		h.editBattleMessage(s, i, "Invalid ability", h.createBattleEmbed(battleState), h.createBattleButtons(battleState, false))
		return
	}

	abilityIndex, err := strconv.Atoi(parts[2])
	if err != nil || abilityIndex < 1 || abilityIndex > len(battleState.PlayerGopher.Abilities) {
		h.editBattleMessage(s, i, "Invalid ability", h.createBattleEmbed(battleState), h.createBattleButtons(battleState, false))
		return
	}

	// Execute ability
	messages, err := battleState.PlayerAction("fight", abilityIndex-1)
	if err != nil {
		h.editBattleMessage(s, i, fmt.Sprintf("Error: %v", err), h.createBattleEmbed(battleState), h.createBattleButtons(battleState, false))
		return
	}

	// Update battle state
	battle, _ := h.battleRepo.GetByID(battleState.ID)
	if battle != nil {
		battle.State = battleState.State
		battle.TurnOwner = battleState.TurnOwner
		h.battleRepo.Update(battle)
	}

	// Check for evolution after level up - check all participating gophers
	evolutionMessages := []string{}
	if strings.Contains(strings.Join(messages, " "), "leveled up") {
		// Check evolution for all participating gophers that may have leveled up
		for _, gopher := range battleState.ParticipatingGophers {
			// Check if this gopher is at an evolution threshold
			shouldCheckEvolution := (gopher.Level >= 16 && gopher.EvolutionStage == 0) ||
				(gopher.Level >= 32 && gopher.EvolutionStage == 1)
			
			if shouldCheckEvolution {
				evolved, evolutionMsg := h.gameService.CheckEvolution(gopher)
				if evolved {
					evolutionMessages = append(evolutionMessages, evolutionMsg)
					// Update gopher after evolution
					h.gopherRepo.Update(h.gameGopherToStorage(gopher))
				}
			}
		}
	}

	// Update all participating gophers in DB (they may have gained XP and leveled up)
	// IMPORTANT: Save HP changes BEFORE checking for blackout
	for _, gopher := range battleState.ParticipatingGophers {
		h.gopherRepo.Update(h.gameGopherToStorage(gopher))
	}
	
	// Update active player gopher (in case it wasn't in participating list)
	h.gopherRepo.Update(h.gameGopherToStorage(battleState.PlayerGopher))
	
	// Update all party members in DB to ensure HP is saved
	partyStorage, _ := h.gopherRepo.GetParty(battleState.TrainerID)
	for _, partyGopher := range partyStorage {
		// Find matching gopher in battle state
		for _, battleGopher := range battleState.PlayerParty {
			if battleGopher.ID == partyGopher.ID {
				partyGopher.CurrentHP = battleGopher.CurrentHP
				partyGopher.MaxHP = battleGopher.MaxHP
				h.gopherRepo.Update(partyGopher)
				break
			}
		}
	}
	
	if battleState.State != "WON" && battleState.State != "ESCAPED" {
		h.gopherRepo.Update(h.gameGopherToStorage(battleState.EnemyGopher))
	}

	// Handle battle end
	if battleState.State != "ACTIVE" {
		if battleState.State == "WON" {
			partySize, _ := h.partyRepo.GetPartySize(battleState.TrainerID)
			enemyStorage := h.gameGopherToStorage(battleState.EnemyGopher)
			enemyStorage.TrainerID = &battleState.TrainerID
			enemyStorage.IsInParty = partySize < 6
			h.gopherRepo.Update(enemyStorage)
			if enemyStorage.IsInParty {
				h.trainerRepo.UpdatePartySlots(battleState.TrainerID, partySize+1)
			}
		} else if battleState.State == "LOST" {
			// Check for blackout after battle loss (HP is already saved above)
			blackedOut, blackoutMsg := h.gameService.CheckAndHandleBlackout(battleState.TrainerID)
			if blackedOut {
				messages = append(messages, "")
				messages = append(messages, blackoutMsg)
			}
		}
		delete(h.battles, battleState.ID)
	}

	// Combine messages
	allMessages := append(messages, evolutionMessages...)
	messageText := strings.Join(allMessages, "\n")
	if messageText == "" {
		messageText = strings.Join(messages, "\n")
	}

	// Update embed with battle card image (don't regenerate for regular actions)
	embed := h.createBattleEmbed(battleState)
	h.addBattleCardToEmbed(embed, battleState, false) // false = don't regenerate, just reference existing
	
	var components []discordgo.MessageComponent
	if battleState.State == "ACTIVE" {
		components = h.createBattleButtons(battleState, false)
	}
	h.editBattleMessage(s, i, messageText, embed, components)
}

func (h *Handlers) findBattleByMessage(channelID, messageID string) *game.BattleState {
	// Check in-memory cache first
	for _, battle := range h.battles {
		if battle.ChannelID == channelID && battle.MessageID == messageID {
			return battle
		}
	}

	// Load from DB
	battle, err := h.battleRepo.GetByMessageID(channelID, messageID)
	if err != nil || battle == nil {
		return nil
	}

	// Reconstruct battle state
	playerGopherStorage, _ := h.gopherRepo.GetByID(*battle.GopherIDPlayer)
	enemyGopherStorage, _ := h.gopherRepo.GetByID(*battle.GopherIDEnemy)

	if playerGopherStorage == nil || enemyGopherStorage == nil {
		return nil
	}

	playerGopher, _ := h.gameService.StorageGopherToGameGopher(playerGopherStorage)
	enemyGopher, _ := h.gameService.StorageGopherToGameGopher(enemyGopherStorage)

	// Get party for battle state
	partyStorage, _ := h.gopherRepo.GetParty(battle.TrainerID)
	gameParty := make([]*game.Gopher, len(partyStorage))
	for i, gopherStorage := range partyStorage {
		gameGopher, _ := h.gameService.StorageGopherToGameGopher(gopherStorage)
		gameParty[i] = gameGopher
	}

	battleState := &game.BattleState{
		ID:                battle.ID,
		ChannelID:         battle.ChannelID,
		MessageID:         battle.MessageID,
		TrainerID:          battle.TrainerID,
		OpponentType:      battle.OpponentType,
		PlayerGopher:      playerGopher,
		EnemyGopher:       enemyGopher,
		PlayerParty:       gameParty,
		ParticipatingGophers: []*game.Gopher{playerGopher}, // Initialize with current
		TurnOwner:         battle.TurnOwner,
		State:             battle.State,
	}

	h.battles[battleState.ID] = battleState
	return battleState
}

func (h *Handlers) createBattleEmbed(battleState *game.BattleState) *discordgo.MessageEmbed {
	playerHPBar := game.GetHPBar(battleState.PlayerGopher.CurrentHP, battleState.PlayerGopher.MaxHP, 12)
	enemyHPBar := game.GetHPBar(battleState.EnemyGopher.CurrentHP, battleState.EnemyGopher.MaxHP, 12)

	description := ""
	if len(battleState.Log) > 0 {
		// Show last 3 log entries
		start := len(battleState.Log) - 3
		if start < 0 {
			start = 0
		}
		description = strings.Join(battleState.Log[start:], "\n")
	}

	color := 0x00ff00
	if battleState.State == "LOST" {
		color = 0xff0000
	} else if battleState.State == "WON" {
		color = 0x00ff00
	} else if battleState.State == "ESCAPED" {
		color = 0xffff00
	}

	embed := &discordgo.MessageEmbed{
		Title:       "Battle",
		Description: description,
		Color:       color,
		Fields: []*discordgo.MessageEmbedField{
			{
				Name:   fmt.Sprintf("%s (Lv.%d)", battleState.PlayerGopher.Name, battleState.PlayerGopher.Level),
				Value:  fmt.Sprintf("HP: %s", playerHPBar),
				Inline: false,
			},
			{
				Name:   fmt.Sprintf("%s (Lv.%d) - %s", battleState.EnemyGopher.Name, battleState.EnemyGopher.Level, battleState.EnemyGopher.Rarity),
				Value:  fmt.Sprintf("HP: %s", enemyHPBar),
				Inline: false,
			},
		},
	}

	if battleState.State != "ACTIVE" {
		embed.Fields = append(embed.Fields, &discordgo.MessageEmbedField{
			Name:   "Battle Result",
			Value:  battleState.State,
			Inline: false,
		})
	}

	return embed
}

// addBattleCardToEmbed adds the battle card image reference to the embed
// Set forceRegen=true to force regeneration (e.g., after swap), false to just reference existing
func (h *Handlers) addBattleCardToEmbed(embed *discordgo.MessageEmbed, battleState *game.BattleState, forceRegen bool) {
	fileName := fmt.Sprintf("battle_%s.png", battleState.ID[:8])
	imageURL := fmt.Sprintf("attachment://%s", fileName)
	
	if forceRegen {
		// Only regenerate if explicitly requested (e.g., after gopher swap)
		// This is expensive, so we avoid it for regular updates
		playerStorage := h.gameGopherToStorage(battleState.PlayerGopher)
		enemyStorage := h.gameGopherToStorage(battleState.EnemyGopher)
		
		cardBase64, err := h.gameService.GenerateBattleCard(enemyStorage, playerStorage)
		if err != nil {
			log.Printf("Error regenerating battle card: %v", err)
			// Fall through to set URL anyway
		} else if cardBase64 != "" {
			// Mark that we need to regenerate the image
			// The editBattleMessage function will handle the actual regeneration
			embed.Image = &discordgo.MessageEmbedImage{
				URL: imageURL,
			}
			return
		}
	}
	
	// For regular updates, just reference the existing attachment
	embed.Image = &discordgo.MessageEmbedImage{
		URL: imageURL,
	}
}

func (h *Handlers) createBattleButtons(battleState *game.BattleState, showAbilities bool) []discordgo.MessageComponent {
	if battleState.State != "ACTIVE" {
		return nil
	}

	if showAbilities && battleState.TurnOwner == "PLAYER" {
		// Show ability buttons
		buttons := []discordgo.MessageComponent{}
		for idx, ability := range battleState.PlayerGopher.Abilities {
			if idx >= 4 {
				break
			}
			buttons = append(buttons, createButton(ability.Name, discordgo.PrimaryButton, fmt.Sprintf("battle_ability_%d", idx+1)))
		}
		return []discordgo.MessageComponent{
			discordgo.ActionsRow{Components: buttons},
		}
	}

	// Show main action buttons
	return []discordgo.MessageComponent{
		discordgo.ActionsRow{
			Components: []discordgo.MessageComponent{
				createButton("Fight", discordgo.PrimaryButton, "battle_fight"),
				createButton("Swap", discordgo.SecondaryButton, "battle_swap"),
				createButton("Run", discordgo.DangerButton, "battle_run"),
				createButton("Throw Net", discordgo.SuccessButton, "battle_net"),
			},
		},
	}
}

func (h *Handlers) gameGopherToStorage(gameGopher *game.Gopher) *storage.Gopher {
	return &storage.Gopher{
		ID:               gameGopher.ID,
		TrainerID:       gameGopher.TrainerID,
		Name:             gameGopher.Name,
		Level:            gameGopher.Level,
		XP:               gameGopher.XP,
		CurrentHP:        gameGopher.CurrentHP,
		MaxHP:            gameGopher.MaxHP,
		Attack:           gameGopher.Attack,
		Defense:          gameGopher.Defense,
		Speed:            gameGopher.Speed,
		Rarity:           gameGopher.Rarity,
		ComplexityScore:  gameGopher.ComplexityScore,
		SpeciesArchetype: gameGopher.SpeciesArchetype,
		EvolutionStage:   gameGopher.EvolutionStage,
		PrimaryType:      string(gameGopher.PrimaryType),
		SecondaryType:    string(gameGopher.SecondaryType),
		SpritePath:       gameGopher.SpritePath,
		SpriteData:       gameGopher.SpriteData,
		GopherkonLayers:  gameGopher.GopherkonLayers,
		IsInParty:        gameGopher.IsInParty,
		PCSlot:           gameGopher.PCSlot,
	}
}

// Helper functions
// ButtonWithoutEmoji is a workaround for discordgo.Button's Emoji field
// that doesn't have omitempty, causing Discord to reject empty emoji objects
type ButtonWithoutEmoji struct {
	Label    string                `json:"label"`
	Style    discordgo.ButtonStyle `json:"style"`
	Disabled bool                  `json:"disabled,omitempty"`
	URL      string                `json:"url,omitempty"`
	CustomID string                `json:"custom_id,omitempty"`
	// Emoji is intentionally omitted to avoid Discord API errors
}

func (b *ButtonWithoutEmoji) Type() discordgo.ComponentType {
	return discordgo.ButtonComponent
}

func (b *ButtonWithoutEmoji) MarshalJSON() ([]byte, error) {
	// Custom marshaler that includes the type field
	return json.Marshal(struct {
		Type     discordgo.ComponentType `json:"type"`
		Label    string                   `json:"label"`
		Style    discordgo.ButtonStyle    `json:"style"`
		Disabled bool                     `json:"disabled,omitempty"`
		URL      string                   `json:"url,omitempty"`
		CustomID string                   `json:"custom_id,omitempty"`
	}{
		Type:     discordgo.ButtonComponent,
		Label:    b.Label,
		Style:    b.Style,
		Disabled: b.Disabled,
		URL:      b.URL,
		CustomID: b.CustomID,
	})
}

// createButton creates a button without emoji to avoid Discord API errors
func createButton(label string, style discordgo.ButtonStyle, customID string) discordgo.MessageComponent {
	return &ButtonWithoutEmoji{
		Label:    label,
		Style:    style,
		CustomID: customID,
	}
}

func respondEphemeral(s *discordgo.Session, i *discordgo.InteractionCreate, content string) {
	err := s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: content,
			Flags:   discordgo.MessageFlagsEphemeral,
		},
	})
	if err != nil {
		log.Printf("Error responding to interaction: %v", err)
	}
}

func respondEmbed(s *discordgo.Session, i *discordgo.InteractionCreate, embed *discordgo.MessageEmbed, ephemeral bool) {
	var flags discordgo.MessageFlags
	if ephemeral {
		flags = discordgo.MessageFlagsEphemeral
	}

	err := s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Embeds: []*discordgo.MessageEmbed{embed},
			Flags:  flags,
		},
	})
	if err != nil {
		log.Printf("Error responding with embed: %v", err)
	}
}

func respondWithComponents(s *discordgo.Session, i *discordgo.InteractionCreate, content string, embed *discordgo.MessageEmbed, components []discordgo.MessageComponent, ephemeral bool) {
	var flags discordgo.MessageFlags
	if ephemeral {
		flags = discordgo.MessageFlagsEphemeral
	}

	data := &discordgo.InteractionResponseData{
		Components: components,
		Flags:      flags,
	}

	if content != "" {
		data.Content = content
	}
	if embed != nil {
		data.Embeds = []*discordgo.MessageEmbed{embed}
	}

	err := s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: data,
	})
	if err != nil {
		log.Printf("Error responding with components: %v", err)
	}
}

// editBattleMessage edits the original battle message, only regenerating image when needed
func (h *Handlers) editBattleMessage(s *discordgo.Session, i *discordgo.InteractionCreate, content string, embed *discordgo.MessageEmbed, components []discordgo.MessageComponent) {
	// Find battle state
	battleState := h.findBattleByMessage(i.ChannelID, i.Message.ID)
	
	// Only regenerate image if embed explicitly requests it (e.g., after swap)
	// For regular updates, just preserve the existing attachment
	needsImageRegen := embed != nil && embed.Image != nil && embed.Image.URL != ""
	
	if battleState != nil && needsImageRegen {
		// Only regenerate if we actually need a new image (e.g., gopher swapped)
		// This is expensive, so we avoid it for regular HP/status updates
		playerStorage := h.gameGopherToStorage(battleState.PlayerGopher)
		enemyStorage := h.gameGopherToStorage(battleState.EnemyGopher)
		
		cardBase64, err := h.gameService.GenerateBattleCard(enemyStorage, playerStorage)
		if err == nil && cardBase64 != "" {
			// Decode base64 to bytes
			fileData, err := base64.StdEncoding.DecodeString(cardBase64)
			if err == nil {
				fileName := fmt.Sprintf("battle_%s.png", battleState.ID[:8])
				
				// Delete old message and send new one with updated image
				err = s.ChannelMessageDelete(i.ChannelID, i.Message.ID)
				if err != nil {
					log.Printf("Error deleting old battle message: %v", err)
					h.editBattleMessageSimple(s, i, content, embed, components)
					return
				}
				
				// Create new message with updated battle card
				battleCardFile := &discordgo.File{
					Name:        fileName,
					ContentType: "image/png",
					Reader:      bytes.NewReader(fileData),
				}
				
				embed.Image.URL = fmt.Sprintf("attachment://%s", fileName)
				
				msgSend := &discordgo.MessageSend{
					Content:    content,
					Embeds:     []*discordgo.MessageEmbed{embed},
					Components: components,
					Files:      []*discordgo.File{battleCardFile},
				}
				
				msg, err := s.ChannelMessageSendComplex(i.ChannelID, msgSend)
				if err != nil {
					log.Printf("Error sending updated battle message: %v", err)
					h.editBattleMessageSimple(s, i, content, embed, components)
				} else {
					// Update battle state with new message ID
					battleState.MessageID = msg.ID
					battle, _ := h.battleRepo.GetByID(battleState.ID)
					if battle != nil {
						battle.MessageID = msg.ID
						h.battleRepo.Update(battle)
					}
				}
				return
			}
		}
	}
	
	// For regular updates, just do a simple edit that preserves existing attachment
	// Set image URL to reference the original attachment if not already set
	if battleState != nil && embed != nil && embed.Image == nil {
		// Preserve the original image attachment by referencing it
		fileName := fmt.Sprintf("battle_%s.png", battleState.ID[:8])
		embed.Image = &discordgo.MessageEmbedImage{
			URL: fmt.Sprintf("attachment://%s", fileName),
		}
	}
	
	h.editBattleMessageSimple(s, i, content, embed, components)
}

// editBattleMessageSimple does a simple message edit without regenerating images
func (h *Handlers) editBattleMessageSimple(s *discordgo.Session, i *discordgo.InteractionCreate, content string, embed *discordgo.MessageEmbed, components []discordgo.MessageComponent) {
	edit := &discordgo.MessageEdit{
		Channel: i.ChannelID,
		ID:      i.Message.ID,
	}

	if content != "" {
		edit.Content = &content
	}

	if embed != nil {
		edit.Embeds = []*discordgo.MessageEmbed{embed}
	}

	if components != nil {
		edit.Components = components
	} else {
		edit.Components = []discordgo.MessageComponent{}
	}

	_, err := s.ChannelMessageEditComplex(edit)
	if err != nil {
		log.Printf("Error editing battle message: %v", err)
	}
}

func editMessage(s *discordgo.Session, i *discordgo.InteractionCreate, content string, embed *discordgo.MessageEmbed, components []discordgo.MessageComponent) {
	data := &discordgo.WebhookEdit{}

	if content != "" {
		data.Content = &content
	}

	if embed != nil {
		data.Embeds = &[]*discordgo.MessageEmbed{embed}
	}

	if components != nil {
		data.Components = &components
	}

	_, err := s.InteractionResponseEdit(i.Interaction, data)
	if err != nil {
		log.Printf("Error editing message: %v", err)
	}
}

// getGopherImageData returns image bytes from base64 or file path
func (h *Handlers) getGopherImageData(gopher *storage.Gopher) ([]byte, error) {
	if gopher.SpriteData != "" {
		// Decode base64 to bytes
		return base64.StdEncoding.DecodeString(gopher.SpriteData)
	} else if gopher.SpritePath != "" {
		// Fallback: read from file (for backward compatibility)
		return os.ReadFile(gopher.SpritePath)
	}
	return nil, fmt.Errorf("gopher has no sprite data")
}

// showPartySwapMenu displays available party members to swap to
func (h *Handlers) showPartySwapMenu(s *discordgo.Session, i *discordgo.InteractionCreate, battleState *game.BattleState) {
	if battleState.State != "ACTIVE" || battleState.TurnOwner != "PLAYER" {
		// Note: interaction already acknowledged in handleBattleAction, so we need to edit the message
		h.editBattleMessage(s, i, "You can't swap right now!", h.createBattleEmbed(battleState), h.createBattleButtons(battleState, false))
		return
	}

	// Note: interaction is already acknowledged in handleBattleAction before this is called

	// Create swap buttons for each party member (excluding current)
	buttons := []discordgo.MessageComponent{}
	for idx, gopher := range battleState.PlayerParty {
		if gopher.ID == battleState.PlayerGopher.ID {
			continue // Skip current gopher
		}
		
		status := ""
		if gopher.CurrentHP <= 0 {
			status = " (FAINTED)"
		}
		
		label := fmt.Sprintf("%s%s", gopher.Name, status)
		if len(label) > 80 {
			label = label[:77] + "..."
		}
		
		buttons = append(buttons, createButton(label, discordgo.SecondaryButton, fmt.Sprintf("battle_swap_%d", idx)))
		
		if len(buttons) >= 5 {
			break // Discord limit
		}
	}

	if len(buttons) == 0 {
		h.editBattleMessage(s, i, "No other party members available!", h.createBattleEmbed(battleState), h.createBattleButtons(battleState, false))
		return
	}

	// Update message with swap options (don't regenerate image for menu)
	embed := h.createBattleEmbed(battleState)
	embed.Description = "Choose a party member to swap to:"
	h.addBattleCardToEmbed(embed, battleState, false) // false = don't regenerate for menu
	
	h.editBattleMessage(s, i, "", embed, []discordgo.MessageComponent{
		discordgo.ActionsRow{Components: buttons},
	})
}

// handleBattleSwap handles swapping to a different party member
func (h *Handlers) handleBattleSwap(s *discordgo.Session, i *discordgo.InteractionCreate) {
	data := i.MessageComponentData()
	discordID := i.Member.User.ID

	// Find battle
	battleState := h.findBattleByMessage(i.ChannelID, i.Message.ID)
	if battleState == nil {
		respondEphemeral(s, i, "Battle not found")
		return
	}

	// Verify ownership
	trainer, err := h.trainerRepo.GetByDiscordID(discordID)
	if err != nil || trainer == nil {
		respondEphemeral(s, i, "Trainer not found")
		return
	}
	if battleState.TrainerID != trainer.ID {
		respondEphemeral(s, i, "This isn't your battle!")
		return
	}

	// Acknowledge interaction
	err = s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseDeferredMessageUpdate,
	})
	if err != nil {
		log.Printf("Error acknowledging swap: %v", err)
		return
	}

	// Parse party member index
	parts := strings.Split(data.CustomID, "_")
	if len(parts) < 3 {
		h.editBattleMessage(s, i, "Invalid swap selection", h.createBattleEmbed(battleState), h.createBattleButtons(battleState, false))
		return
	}

	partyIndex, err := strconv.Atoi(parts[2])
	if err != nil || partyIndex < 0 || partyIndex >= len(battleState.PlayerParty) {
		h.editBattleMessage(s, i, "Invalid party member", h.createBattleEmbed(battleState), h.createBattleButtons(battleState, false))
		return
	}

	// Execute swap
	messages, err := battleState.PlayerAction("swap", partyIndex)
	if err != nil {
		h.editBattleMessage(s, i, fmt.Sprintf("Error: %v", err), h.createBattleEmbed(battleState), h.createBattleButtons(battleState, false))
		return
	}

	// Update gophers in DB
	h.gopherRepo.Update(h.gameGopherToStorage(battleState.PlayerGopher))
	for _, gopher := range battleState.PlayerParty {
		h.gopherRepo.Update(h.gameGopherToStorage(gopher))
	}

	// Update embed with battle card image (don't regenerate for regular actions)
	embed := h.createBattleEmbed(battleState)
	h.addBattleCardToEmbed(embed, battleState, false) // false = don't regenerate, just reference existing
	
	var components []discordgo.MessageComponent
	if battleState.State == "ACTIVE" {
		components = h.createBattleButtons(battleState, false)
	}
	h.editBattleMessage(s, i, strings.Join(messages, "\n"), embed, components)
}
