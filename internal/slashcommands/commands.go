package slashcommands

import "github.com/bwmarrin/discordgo"

var Commands = []*discordgo.ApplicationCommand{
	{
		Name: "register",
		Description: "Registers server",
	},
	// {
	// 	Name: "report",
	// 	Description: "Reports a user",
	// 	Options: []*discordgo.ApplicationCommandOption{
	// 		{
	// 			Type: discordgo.ApplicationCommandOptionUser,
	// 			Name: "reported-user-option",
	// 			Description: "The user to report",
	// 			Required: true,
	// 		},
	// 	},
	// },
}
