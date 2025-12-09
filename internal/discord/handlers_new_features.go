package discord

import (
	"fmt"

	"github.com/bwmarrin/discordgo"
)

// New handler functions for all the new features

func (h *Handlers) handleShop(s *discordgo.Session, i *discordgo.InteractionCreate) {
	data := i.ApplicationCommandData()
	discordID := i.Member.User.ID

	trainer, err := h.trainerRepo.GetByDiscordID(discordID)
	if err != nil || trainer == nil {
		respondEphemeral(s, i, "Trainer not found. Use /start first.")
		return
	}

	if len(data.Options) == 0 {
		// View shop
		embed := &discordgo.MessageEmbed{
			Title:       "🛒 Gopher Shop 🛒",
			Description: fmt.Sprintf("Your currency: **%d** GoCoins\n\nAvailable items:", trainer.Currency),
			Color:       0x00ff00,
			Fields: []*discordgo.MessageEmbedField{
				{Name: "💊 Potion", Value: "Heals 50 HP\n**Price:** 50 GoCoins", Inline: true},
				{Name: "💉 Revive", Value: "Restores fainted gopher\n**Price:** 100 GoCoins", Inline: true},
				{Name: "⚡ XP Booster", Value: "1.5x XP for next battle\n**Price:** 200 GoCoins", Inline: true},
				{Name: "💎 Evolution Stone", Value: "Force evolution\n**Price:** 500 GoCoins", Inline: true},
				{Name: "✨ Shiny Charm", Value: "Doubles shiny rate\n**Price:** 1000 GoCoins", Inline: true},
			},
		}
		respondEmbed(s, i, embed, true)
		return
	}

	subCommand := data.Options[0]
	if subCommand.Name == "buy" {
		itemType := subCommand.Options[0].StringValue()
		quantity := 1
		if len(subCommand.Options) > 1 {
			quantity = int(subCommand.Options[1].IntValue())
		}

		var price int
		switch itemType {
		case "POTION":
			price = 50
		case "REVIVE":
			price = 100
		case "XP_BOOSTER":
			price = 200
		case "EVOLUTION_STONE":
			price = 500
		case "SHINY_CHARM":
			price = 1000
		default:
			respondEphemeral(s, i, "Invalid item type")
			return
		}

		totalCost := price * quantity
		if trainer.Currency < totalCost {
			respondEphemeral(s, i, fmt.Sprintf("Insufficient currency! You need %d GoCoins but only have %d.", totalCost, trainer.Currency))
			return
		}

		if err := h.trainerRepo.RemoveCurrency(trainer.ID, totalCost); err != nil {
			respondEphemeral(s, i, fmt.Sprintf("Error: %v", err))
			return
		}

		respondEphemeral(s, i, fmt.Sprintf("Purchased %d %s for %d GoCoins!", quantity, itemType, totalCost))
	}
}

func (h *Handlers) handleAchievements(s *discordgo.Session, i *discordgo.InteractionCreate) {
	discordID := i.Member.User.ID

	trainer, err := h.trainerRepo.GetByDiscordID(discordID)
	if err != nil || trainer == nil {
		respondEphemeral(s, i, "Trainer not found. Use /start first.")
		return
	}

	embed := &discordgo.MessageEmbed{
		Title:       "🏆 Achievements 🏆",
		Description: "Your achievement progress will be displayed here.",
		Color:       0xffd700,
		Fields:      []*discordgo.MessageEmbedField{},
	}

	respondEmbed(s, i, embed, true)
}

func (h *Handlers) handleQuests(s *discordgo.Session, i *discordgo.InteractionCreate) {
	discordID := i.Member.User.ID

	trainer, err := h.trainerRepo.GetByDiscordID(discordID)
	if err != nil || trainer == nil {
		respondEphemeral(s, i, "Trainer not found. Use /start first.")
		return
	}

	embed := &discordgo.MessageEmbed{
		Title:       "📋 Quests 📋",
		Description: "Your active quests will be displayed here.",
		Color:       0x00ff00,
		Fields:      []*discordgo.MessageEmbedField{},
	}

	respondEmbed(s, i, embed, true)
}

func (h *Handlers) handleChallenge(s *discordgo.Session, i *discordgo.InteractionCreate) {
	data := i.ApplicationCommandData()
	discordID := i.Member.User.ID

	trainer, err := h.trainerRepo.GetByDiscordID(discordID)
	if err != nil || trainer == nil {
		respondEphemeral(s, i, "Trainer not found. Use /start first.")
		return
	}

	opponentUser := data.Options[0].UserValue(s)
	if opponentUser == nil {
		respondEphemeral(s, i, "Invalid user")
		return
	}

	if opponentUser.ID == discordID {
		respondEphemeral(s, i, "You can't challenge yourself!")
		return
	}

	opponentTrainer, err := h.trainerRepo.GetByDiscordID(opponentUser.ID)
	if err != nil || opponentTrainer == nil {
		respondEphemeral(s, i, "That user is not a trainer yet!")
		return
	}

	party1, _ := h.gopherRepo.GetParty(trainer.ID)
	party2, _ := h.gopherRepo.GetParty(opponentTrainer.ID)

	if len(party1) == 0 {
		respondEphemeral(s, i, "Your party is empty!")
		return
	}
	if len(party2) == 0 {
		respondEphemeral(s, i, "Opponent's party is empty!")
		return
	}

	respondEphemeral(s, i, fmt.Sprintf("PvP battle challenge sent to %s! (Feature in development)", opponentUser.Username))
}

func (h *Handlers) handleStats(s *discordgo.Session, i *discordgo.InteractionCreate) {
	data := i.ApplicationCommandData()

	var targetUser *discordgo.User
	if len(data.Options) > 0 {
		targetUser = data.Options[0].UserValue(s)
	}
	if targetUser == nil {
		targetUser = i.Member.User
	}

	trainer, err := h.trainerRepo.GetByDiscordID(targetUser.ID)
	if err != nil || trainer == nil {
		respondEphemeral(s, i, "Trainer not found.")
		return
	}

	embed := &discordgo.MessageEmbed{
		Title:       fmt.Sprintf("📊 %s's Statistics", trainer.Name),
		Description: "Statistics will be displayed here.",
		Color:       0x0099ff,
		Fields:      []*discordgo.MessageEmbedField{},
	}

	respondEmbed(s, i, embed, true)
}

func (h *Handlers) handleLeaderboard(s *discordgo.Session, i *discordgo.InteractionCreate) {
	data := i.ApplicationCommandData()
	leaderboardType := data.Options[0].StringValue()

	embed := &discordgo.MessageEmbed{
		Title:       "🏆 Leaderboard",
		Description: fmt.Sprintf("Top players by %s", leaderboardType),
		Color:       0xffd700,
		Fields:      []*discordgo.MessageEmbedField{},
	}

	respondEmbed(s, i, embed, true)
}

func (h *Handlers) handleGopherdex(s *discordgo.Session, i *discordgo.InteractionCreate) {
	discordID := i.Member.User.ID

	trainer, err := h.trainerRepo.GetByDiscordID(discordID)
	if err != nil || trainer == nil {
		respondEphemeral(s, i, "Trainer not found. Use /start first.")
		return
	}

	embed := &discordgo.MessageEmbed{
		Title:       "📖 Gopherdex",
		Description: "Your collection will be displayed here.",
		Color:       0x9966ff,
		Fields:      []*discordgo.MessageEmbedField{},
	}

	respondEmbed(s, i, embed, true)
}

func (h *Handlers) handleTrade(s *discordgo.Session, i *discordgo.InteractionCreate) {
	data := i.ApplicationCommandData()
	discordID := i.Member.User.ID

	trainer, err := h.trainerRepo.GetByDiscordID(discordID)
	if err != nil || trainer == nil {
		respondEphemeral(s, i, "Trainer not found. Use /start first.")
		return
	}

	subCommand := data.Options[0]
	switch subCommand.Name {
	case "offer":
		opponentUser := subCommand.Options[0].UserValue(s)
		if opponentUser == nil {
			respondEphemeral(s, i, "Invalid user")
			return
		}

		respondEphemeral(s, i, fmt.Sprintf("Trade offer sent to %s!", opponentUser.Username))
	case "accept":
		tradeID := subCommand.Options[0].StringValue()
		respondEphemeral(s, i, fmt.Sprintf("Trade %s accepted!", tradeID))
	case "list":
		embed := &discordgo.MessageEmbed{
			Title:       "💼 Pending Trades",
			Description: "Your pending trades will be listed here.",
			Color:       0x00ff00,
		}
		respondEmbed(s, i, embed, true)
	}
}

