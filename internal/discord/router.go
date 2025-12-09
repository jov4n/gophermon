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
		{
			Name:        "shop",
			Description: "View and buy items from the shop",
			Options: []*discordgo.ApplicationCommandOption{
				{
					Type:        discordgo.ApplicationCommandOptionSubCommand,
					Name:        "view",
					Description: "View available items",
				},
				{
					Type:        discordgo.ApplicationCommandOptionSubCommand,
					Name:        "buy",
					Description: "Buy an item",
					Options: []*discordgo.ApplicationCommandOption{
						{
							Type:        discordgo.ApplicationCommandOptionString,
							Name:        "item",
							Description: "Item to buy",
							Required:    true,
							Choices: []*discordgo.ApplicationCommandOptionChoice{
								{Name: "Potion", Value: "POTION"},
								{Name: "Revive", Value: "REVIVE"},
								{Name: "XP Booster", Value: "XP_BOOSTER"},
								{Name: "Evolution Stone", Value: "EVOLUTION_STONE"},
								{Name: "Shiny Charm", Value: "SHINY_CHARM"},
							},
						},
						{
							Type:        discordgo.ApplicationCommandOptionInteger,
							Name:        "quantity",
							Description: "Quantity to buy (default: 1)",
							Required:    false,
						},
					},
				},
			},
		},
		{
			Name:        "achievements",
			Description: "View your achievements",
		},
		{
			Name:        "quests",
			Description: "View your active quests",
		},
		{
			Name:        "challenge",
			Description: "Challenge another trainer to a PvP battle",
			Options: []*discordgo.ApplicationCommandOption{
				{
					Type:        discordgo.ApplicationCommandOptionUser,
					Name:        "user",
					Description: "The user to challenge",
					Required:    true,
				},
			},
		},
		{
			Name:        "stats",
			Description: "View your statistics",
			Options: []*discordgo.ApplicationCommandOption{
				{
					Type:        discordgo.ApplicationCommandOptionUser,
					Name:        "user",
					Description: "View another user's stats (optional)",
					Required:    false,
				},
			},
		},
		{
			Name:        "leaderboard",
			Description: "View leaderboards",
			Options: []*discordgo.ApplicationCommandOption{
				{
					Type:        discordgo.ApplicationCommandOptionString,
					Name:        "type",
					Description: "Leaderboard type",
					Required:    true,
					Choices: []*discordgo.ApplicationCommandOptionChoice{
						{Name: "PvP Rating", Value: "pvp"},
						{Name: "Battles Won", Value: "wins"},
						{Name: "Shinies", Value: "shinies"},
						{Name: "Gophers Caught", Value: "caught"},
					},
				},
			},
		},
		{
			Name:        "gopherdex",
			Description: "View your Gopherdex (collection)",
		},
		{
			Name:        "trade",
			Description: "Trade with another trainer",
			Options: []*discordgo.ApplicationCommandOption{
				{
					Type:        discordgo.ApplicationCommandOptionSubCommand,
					Name:        "offer",
					Description: "Offer a trade",
					Options: []*discordgo.ApplicationCommandOption{
						{
							Type:        discordgo.ApplicationCommandOptionUser,
							Name:        "user",
							Description: "User to trade with",
							Required:    true,
						},
						{
							Type:        discordgo.ApplicationCommandOptionString,
							Name:        "gopher_id",
							Description: "Gopher to trade (optional)",
							Required:    false,
						},
						{
							Type:        discordgo.ApplicationCommandOptionInteger,
							Name:        "currency",
							Description: "Currency to offer (optional)",
							Required:    false,
						},
					},
				},
				{
					Type:        discordgo.ApplicationCommandOptionSubCommand,
					Name:        "accept",
					Description: "Accept a pending trade",
					Options: []*discordgo.ApplicationCommandOption{
						{
							Type:        discordgo.ApplicationCommandOptionString,
							Name:        "trade_id",
							Description: "Trade ID to accept",
							Required:    true,
						},
					},
				},
				{
					Type:        discordgo.ApplicationCommandOptionSubCommand,
					Name:        "list",
					Description: "List pending trades",
				},
			},
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
				{
					Type:        discordgo.ApplicationCommandOptionSubCommand,
					Name:        "rename",
					Description: "Rename a gopher",
					Options: []*discordgo.ApplicationCommandOption{
						{
							Type:        discordgo.ApplicationCommandOptionString,
							Name:        "gopher_id",
							Description: "The ID of the gopher",
							Required:    true,
						},
						{
							Type:        discordgo.ApplicationCommandOptionString,
							Name:        "new_name",
							Description: "New name for the gopher",
							Required:    true,
						},
					},
				},
				{
					Type:        discordgo.ApplicationCommandOptionSubCommand,
					Name:        "favorite",
					Description: "Mark/unmark a gopher as favorite",
					Options: []*discordgo.ApplicationCommandOption{
						{
							Type:        discordgo.ApplicationCommandOptionString,
							Name:        "gopher_id",
							Description: "The ID of the gopher",
							Required:    true,
						},
					},
				},
				{
					Type:        discordgo.ApplicationCommandOptionSubCommand,
					Name:        "release",
					Description: "Release a gopher (get currency)",
					Options: []*discordgo.ApplicationCommandOption{
						{
							Type:        discordgo.ApplicationCommandOptionString,
							Name:        "gopher_id",
							Description: "The ID of the gopher to release",
							Required:    true,
						},
					},
				},
			},
		},
		{
			Name:        "party",
			Description: "View your active party of gophers",
			Options: []*discordgo.ApplicationCommandOption{
				{
					Type:        discordgo.ApplicationCommandOptionSubCommand,
					Name:        "view",
					Description: "View your party",
				},
				{
					Type:        discordgo.ApplicationCommandOptionSubCommand,
					Name:        "heal",
					Description: "Heal all party members (costs currency)",
				},
			},
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
				{
					Type:        discordgo.ApplicationCommandOptionSubCommand,
					Name:        "search",
					Description: "Search PC gophers",
					Options: []*discordgo.ApplicationCommandOption{
						{
							Type:        discordgo.ApplicationCommandOptionString,
							Name:        "rarity",
							Description: "Filter by rarity (optional)",
							Required:    false,
							Choices: []*discordgo.ApplicationCommandOptionChoice{
								{Name: "Common", Value: "COMMON"},
								{Name: "Uncommon", Value: "UNCOMMON"},
								{Name: "Rare", Value: "RARE"},
								{Name: "Epic", Value: "EPIC"},
								{Name: "Legendary", Value: "LEGENDARY"},
							},
						},
						{
							Type:        discordgo.ApplicationCommandOptionString,
							Name:        "archetype",
							Description: "Filter by archetype (optional)",
							Required:    false,
							Choices: []*discordgo.ApplicationCommandOptionChoice{
								{Name: "Hacker", Value: "Hacker"},
								{Name: "Tank", Value: "Tank"},
								{Name: "Speedy", Value: "Speedy"},
								{Name: "Support", Value: "Support"},
								{Name: "Mage", Value: "Mage"},
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

