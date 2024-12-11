package slashcommands

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"log/slog"
	"net/http"
	"snitch/snitchbot/internal/botconfig"
	"snitch/snitchbot/pkg/ctxutil"

	"github.com/bwmarrin/discordgo"
)

type RegistrationRequest struct {
	ServerID string `json:"serverId"` // we need to tell go that our number is encoded as a string, hence ',string'
	UserID   string `json:"userId"`   // we need to tell go that our number is encoded as a string, hence ',string'
}

type RegistrationResponse struct {
	ServerID string    `json:"serverId"` // we need to tell go that our number is encoded as a string, hence ',string'
	GroupID  string `json:"groupId"`
}

type SlashCommandHandlerFunc func(*discordgo.Session, *discordgo.InteractionCreate, context.Context)

func (slashCommandFuncContext SlashCommandHandlerFunc) Adapt() func(*discordgo.Session, *discordgo.InteractionCreate) {
	return func (session *discordgo.Session, interaction *discordgo.InteractionCreate) {
		slashCommandFuncContext(session, interaction, context.Background())
	}
}

func CreateRegisterCommandHandler(botconfig botconfig.BotConfig) SlashCommandHandlerFunc {
	backendURL, err := botconfig.BackendURL()
	if err != nil {
		log.Fatal(backendURL)
	}
	
	return func(session *discordgo.Session, interaction *discordgo.InteractionCreate, ctx context.Context) {
		slogger, ok := ctxutil.Value[*slog.Logger](ctx)
		if !ok {
			slogger = slog.Default()
		}

		serverId := interaction.GuildID
		userId := interaction.Member.User.ID
		
		requestStruct := &RegistrationRequest{ ServerID: serverId, UserID: userId }
		
		requestBody, err := json.Marshal(requestStruct)
		if err != nil {
			log.Print(err)
			return
		}
		
		requestURL := backendURL.JoinPath("databases")
		request, err := http.NewRequestWithContext(ctx, "POST", requestURL.String(), bytes.NewBuffer(requestBody))
		if err != nil {
			slogger.ErrorContext(ctx,"Backend Request Creation", "Error", err)
			return
		}
		
		response, err := session.Client.Do(request)
		if (err != nil) {
			slogger.ErrorContext(ctx, "Backend Request Call", "Error", err)
			return
		}

		if response.StatusCode >= 300 || response.StatusCode < 200 {
				body, _ := io.ReadAll(response.Body)
				defer response.Body.Close()
			slogger.ErrorContext(ctx, "Unexpected Response", "Status", response.StatusCode, "Body", string(body))
			return
		}
		
		body, err := io.ReadAll(response.Body)
		defer response.Body.Close()
		if err != nil {
			slogger.ErrorContext(ctx, "Couldn't Read Body", "Error", err)
			return
		}
		
		var registrationResponse RegistrationResponse
		if err := json.Unmarshal(body, &registrationResponse); err != nil {
			slogger.ErrorContext(ctx, "Couldn't Unmarshal Body", "Error", err)
			return
		}
		
		if err = session.InteractionRespond(interaction.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: fmt.Sprintf("Created group %s for this server.", registrationResponse.GroupID),
			},
		}); err != nil {
			slogger.ErrorContext(ctx, "Couldn't Write Discord Response", "Error", err)
			return
		}
	}
}
