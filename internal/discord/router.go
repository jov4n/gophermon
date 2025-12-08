package discord

import (
	"github.com/bwmarrin/discordgo"
)

func RegisterCommands(s *discordgo.Session, guildID string) error {
	commands := []*discordgo.ApplicationCommand{
		{
			Name:        "ping",
			Description: "Check if the bot is responding",
		},
		{
			Name:        "start",
			Description: "Begin your gopher journey and choose a starter",
		},
		{
			Name:        "choose",
			Description: "Choose your starter gopher",
			Options: []*discordgo.ApplicationCommandOption{
				{
					Type:        discordgo.ApplicationCommandOptionInteger,
					Name:        "number",
					Description: "Starter number (1, 2, or 3)",
					Required:    true,
					Choices: []*discordgo.ApplicationCommandOptionChoice{
						{Name: "Starter 1", Value: 1},
						{Name: "Starter 2", Value: 2},
						{Name: "Starter 3", Value: 3},
					},
				},
			},
		},
		{
			Name:        "party",
			Description: "View your active party of gophers",
		},
		{
			Name:        "pc",
			Description: "View gophers in your PC storage",
			Options: []*discordgo.ApplicationCommandOption{
				{
					Type:        discordgo.ApplicationCommandOptionSubCommand,
					Name:        "list",
					Description: "List gophers in PC",
					Options: []*discordgo.ApplicationCommandOption{
						{
							Type:        discordgo.ApplicationCommandOptionInteger,
							Name:        "page",
							Description: "Page number (default: 1)",
							Required:    false,
						},
					},
				},
				{
					Type:        discordgo.ApplicationCommandOptionSubCommand,
					Name:        "deposit",
					Description: "Move a gopher from party to PC",
					Options: []*discordgo.ApplicationCommandOption{
						{
							Type:        discordgo.ApplicationCommandOptionString,
							Name:        "gopher_id",
							Description: "The ID of the gopher to deposit",
							Required:    true,
						},
					},
				},
				{
					Type:        discordgo.ApplicationCommandOptionSubCommand,
					Name:        "withdraw",
					Description: "Move a gopher from PC to party",
					Options: []*discordgo.ApplicationCommandOption{
						{
							Type:        discordgo.ApplicationCommandOptionString,
							Name:        "gopher_id",
							Description: "The ID of the gopher to withdraw",
							Required:    true,
						},
					},
				},
			},
		},
		{
			Name:        "wild",
			Description: "Encounter a wild gopher",
		},
		{
			Name:        "gopher",
			Description: "View detailed information about a gopher",
			Options: []*discordgo.ApplicationCommandOption{
				{
					Type:        discordgo.ApplicationCommandOptionSubCommand,
					Name:        "info",
					Description: "Get detailed info about a gopher",
					Options: []*discordgo.ApplicationCommandOption{
						{
							Type:        discordgo.ApplicationCommandOptionString,
							Name:        "gopher_id",
							Description: "The ID of the gopher",
							Required:    true,
						},
					},
				},
			},
		},
		{
			Name:        "generate_10",
			Description: "Generate 10 random gophers on a card to test generation",
		},
	}

	for _, cmd := range commands {
		_, err := s.ApplicationCommandCreate(s.State.User.ID, guildID, cmd)
		if err != nil {
			return err
		}
	}

	return nil
}

