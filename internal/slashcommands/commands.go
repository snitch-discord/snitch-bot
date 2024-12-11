package slashcommands

import "github.com/bwmarrin/discordgo"

var Commands = []*discordgo.ApplicationCommand{
	{
		Name:        "register",
		Description: "Registers server",
	},
	{
		Name:        "report",
		Description: "Reports a user",
		Options: []*discordgo.ApplicationCommandOption{
			{
				Name:        "new",
				Description: "Creates a new report for the targeted user",
				Type:        discordgo.ApplicationCommandOptionSubCommand,
				Options: []*discordgo.ApplicationCommandOption{
					{
						Name:        "reported-user",
						Type:        discordgo.ApplicationCommandOptionUser,
						Description: "The user to report",
						Required:    true,
					},
					{
						Name:        "report-reason",
						Type:        discordgo.ApplicationCommandOptionString,
						Description: "Report reason",
						Required:    false,
					},
				},
			},
			{
				Name:        "list",
				Description: "Retrieves reports",
				Type:        discordgo.ApplicationCommandOptionSubCommand,
				Options: []*discordgo.ApplicationCommandOption{
					{
						Name:        "reported-user",
						Type:        discordgo.ApplicationCommandOptionUser,
						Description: "The user to retrieve reports for",
						Required:    false,
					},
					{
						Name:        "reporter-user",
						Type:        discordgo.ApplicationCommandOptionUser,
						Description: "The user who created the reports",
						Required:    false,
					},
				},
			},
			{
				Name:        "delete",
				Description: "Deletes a report",
				Type:        discordgo.ApplicationCommandOptionSubCommand,
				Options: []*discordgo.ApplicationCommandOption{
					{
						Name:        "report-id",
						Type:        discordgo.ApplicationCommandOptionString,
						Description: "Report ID",
						Required:    true,
					},
				},
			},
		},
	},
}
