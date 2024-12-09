package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/signal"
	"snitch/snitchbot/internal/botconfig"
	"time"

	"github.com/bwmarrin/discordgo"
)

type registrationRequest struct {
	ServerID string `json:"serverId"` // we need to tell go that our number is encoded as a string, hence ',string'
	UserID   string `json:"userId"`   // we need to tell go that our number is encoded as a string, hence ',string'
}

type registrationResponse struct {
	ServerID string    `json:"serverId"` // we need to tell go that our number is encoded as a string, hence ',string'
	GroupID  string `json:"groupId"`
}

var commands = []*discordgo.ApplicationCommand{
	{
		Name: "register",
		Description: "Registers server",
	},
	// {
	// 	Name: "report-command",
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

var httpClient = &http.Client {
	Timeout: 5 * time.Second,
}

func createRegisterHandler(httpClient *http.Client, botconfig botconfig.BotConfig) func(*discordgo.Session, *discordgo.InteractionCreate) {
	backendURL, err := botconfig.BackendURL()
	if err != nil {
		log.Fatal(backendURL)
	}
	
	return func(session *discordgo.Session, interaction *discordgo.InteractionCreate) {
		context, cancel := context.WithTimeout(context.Background(), time.Second * 5)
		defer cancel()

		serverId := interaction.GuildID
		userId := interaction.Member.User.ID
		
		requestStruct := &registrationRequest{ ServerID: serverId, UserID: userId }
		
		requestBody, err := json.Marshal(requestStruct)
		if err != nil {
			log.Print(err)
			return
		}
		
		requestURL := backendURL.JoinPath("databases")
		request, err := http.NewRequestWithContext(context, "POST", requestURL.String(), bytes.NewBuffer(requestBody))
		if err != nil {
			log.Print(err)
			return
		}
		
		response, err := httpClient.Do(request)
		if (err != nil) {
			log.Print(err)
			return
		}

		if response.StatusCode >= 300 || response.StatusCode < 200 {
			log.Printf("Unexpected Response; Status: %d", response.StatusCode)
			return
		}
		
		body, err := io.ReadAll(response.Body)
		defer response.Body.Close()
		if err != nil {
			log.Print(err)
			return
		}
		
		var registrationResponse registrationResponse
		if err := json.Unmarshal(body, &registrationResponse); err != nil {
			log.Print(err)
			return
		}
		
		if err = session.InteractionRespond(interaction.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: fmt.Sprintf("Created group %s for this server.", registrationResponse.GroupID),
			},
		}); err != nil {
			log.Print(err)
			return
		}
	}
}

func main() {
	config, err := botconfig.BotConfigFromEnv()
	if err != nil {
		log.Fatal(err)
	}

	commandHandlers := map[string] func(session *discordgo.Session, interaction *discordgo.InteractionCreate) {
		"register": createRegisterHandler(httpClient, config),
	}

	mainSession, err := discordgo.New("Bot " + config.DiscordToken)
	if err != nil {
		log.Fatal(err)
	}
	defer mainSession.Close()

	mainSession.AddHandler(func(session *discordgo.Session, interaction *discordgo.Ready) {
		log.Printf("Logged in as: %s#%s", session.State.User.Username, session.State.User.Discriminator)
	})
	mainSession.AddHandler(func(session *discordgo.Session, interaction *discordgo.InteractionCreate) {
		if handler, ok := commandHandlers[interaction.ApplicationCommandData().Name]; ok {
			handler(session, interaction)
		}
	})
	
	if err = mainSession.Open(); err != nil {
		log.Fatal(err)
	}

	registeredCommands := make([]*discordgo.ApplicationCommand, len(commands))
	for index, applicationCommand := range commands {
		createdCommand, err := mainSession.ApplicationCommandCreate(mainSession.State.User.ID, "", applicationCommand)
		if err != nil {
			log.Panicf("Cannot create '%v' command: %v", applicationCommand.Name, err)
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
	
	for _, registeredCommand := range registeredCommands {
		if err = mainSession.ApplicationCommandDelete(mainSession.State.User.ID, "", registeredCommand.ID); err != nil {
			log.Panicf("Cannot delete '%v' command: '%v'", registeredCommand.Name, err)
		}
	}
}
