package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"snitch/snitchbot/internal/botconfig"
	"snitch/snitchbot/internal/slashcommands"
	"time"

	"github.com/bwmarrin/discordgo"
)

func main() {
	testingGuildID := "1315524176936964117"

	config, err := botconfig.BotConfigFromEnv()
	if err != nil {
		log.Fatal(err)
	}

	// initialize map of command name to command handler
	commandHandlers := map[string] slashcommands.SlashCommandHandlerFunc {
		"register": slashcommands.CreateRegisterCommandHandler(config),
	}

	for _, command := range slashcommands.Commands {
		_, handlerPresent := commandHandlers[command.Name]

		if (!handlerPresent) {
			log.Fatalf("Missing Handler for %s", command.Name)
		}
	}

	mainSession, err := discordgo.New("Bot " + config.DiscordToken)
	if err != nil {
		log.Fatal(err)
	}
	defer mainSession.Close()

	mainSession.AddHandler(func(session *discordgo.Session, interaction *discordgo.Ready) {
		log.Printf("Logged in as: %s#%s", session.State.User.Username, session.State.User.Discriminator)
	})
	// setup our listeners for interaction events (a user using a slash command)

  handler := func(session *discordgo.Session, interaction *discordgo.InteractionCreate, context context.Context) {
		if handler, ok := commandHandlers[interaction.ApplicationCommandData().Name]; ok {
			handler(session, interaction, context)
		}
	}
	handler = slashcommands.ResponseTime(handler)
	handler = slashcommands.Recovery(handler)
	handler = slashcommands.Log(handler)
	handler = slashcommands.WithTimeout(handler, time.Second * 10)
	mainSession.AddHandler(slashcommands.SlashCommandHandlerFunc(handler).Adapt())
	
	if err = mainSession.Open(); err != nil {
		log.Fatal(err)
	}

	// tells discord about the commands we support
	registeredCommands := make([]*discordgo.ApplicationCommand, len(slashcommands.Commands))
	for index, applicationCommand := range slashcommands.Commands {
		createdCommand, err := mainSession.ApplicationCommandCreate(mainSession.State.User.ID, testingGuildID, applicationCommand)
		if err != nil {
			log.Fatalf("Cannot register '%v' command: %v", applicationCommand.Name, err)
		}
		registeredCommands[index] = createdCommand
	}

	if err != nil {
		log.Fatal(err)
	}
	
	stopChannel := make(chan os.Signal, 1)
	signal.Notify(stopChannel, os.Interrupt)
	<-stopChannel
	
	log.Println("Shutting down gracefully...")

	// cleanup commands
	for _, registeredCommand := range registeredCommands {
		if err = mainSession.ApplicationCommandDelete(mainSession.State.User.ID, testingGuildID, registeredCommand.ID); err != nil {
			log.Panicf("Cannot delete '%v' command: '%v'", registeredCommand.Name, err)
		}
	}
}
