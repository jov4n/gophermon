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
		{
			Name:        "events",
			Description: "View and manage active events",
			Options: []*discordgo.ApplicationCommandOption{
				{
					Type:        discordgo.ApplicationCommandOptionSubCommand,
					Name:        "list",
					Description: "List all currently active events",
				},
				{
					Type:        discordgo.ApplicationCommandOptionSubCommand,
					Name:        "start",
					Description: "Start a new event (admin only)",
					Options: []*discordgo.ApplicationCommandOption{
						{
							Type:        discordgo.ApplicationCommandOptionString,
							Name:        "type",
							Description: "Event type",
							Required:    true,
							Choices: []*discordgo.ApplicationCommandOptionChoice{
								{Name: "Shiny Hunt", Value: "SHINY_HUNT"},
								{Name: "Double XP", Value: "DOUBLE_XP"},
								{Name: "Rare Encounter", Value: "RARE_ENCOUNTER"},
								{Name: "Lucky Day", Value: "LUCKY_DAY"},
								{Name: "Stat Boost", Value: "STAT_BOOST"},
								{Name: "Evolution Festival", Value: "EVOLUTION_FEST"},
							},
						},
						{
							Type:        discordgo.ApplicationCommandOptionInteger,
							Name:        "hours",
							Description: "Duration in hours (default: 24)",
							Required:    false,
						},
					},
				},
				{
					Type:        discordgo.ApplicationCommandOptionSubCommand,
					Name:        "end",
					Description: "End an active event (admin only)",
					Options: []*discordgo.ApplicationCommandOption{
						{
							Type:        discordgo.ApplicationCommandOptionString,
							Name:        "type",
							Description: "Event type to end",
							Required:    true,
							Choices: []*discordgo.ApplicationCommandOptionChoice{
								{Name: "Shiny Hunt", Value: "SHINY_HUNT"},
								{Name: "Double XP", Value: "DOUBLE_XP"},
								{Name: "Rare Encounter", Value: "RARE_ENCOUNTER"},
								{Name: "Lucky Day", Value: "LUCKY_DAY"},
								{Name: "Stat Boost", Value: "STAT_BOOST"},
								{Name: "Evolution Festival", Value: "EVOLUTION_FEST"},
							},
						},
					},
				},
			},
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

