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
	ServerID string `json:"serverId"` // we need to tell go that our number is encoded as a string, hence ',string'
	GroupID  string `json:"groupId"`
}

type SlashCommandHandlerFunc func(*discordgo.Session, *discordgo.InteractionCreate, context.Context)

func (slashCommandFuncContext SlashCommandHandlerFunc) Adapt() func(*discordgo.Session, *discordgo.InteractionCreate) {
	return func(session *discordgo.Session, interaction *discordgo.InteractionCreate) {
		slashCommandFuncContext(session, interaction, context.Background())
	}
}

func handleNewReport(session *discordgo.Session, interaction *discordgo.InteractionCreate, ctx context.Context) {
	slogger, ok := ctxutil.Value[*slog.Logger](ctx)
	if !ok {
		slogger = slog.Default()
	}

	options := interaction.ApplicationCommandData().Options[0].Options
	optionMap := make(map[string]*discordgo.ApplicationCommandInteractionDataOption, len(options))
	for _, opt := range options {
		optionMap[opt.Name] = opt
	}

	reportedUserOption, ok := optionMap["reported-user"]
	if !ok {
		slogger.ErrorContext(ctx, "Failed to get reported user option")
		return
	}
	reportedUser := reportedUserOption.UserValue(session)

	reportReason := ""
	reportReasonOption, ok := optionMap["report-reason"]
	if ok {
		reportReason = reportReasonOption.StringValue()
	}

	responseContent := fmt.Sprintf("Reported user: %s; Report reason: %s", reportedUser.Username, reportReason)

	if err := session.InteractionRespond(interaction.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: responseContent,
		},
	}); err != nil {
		slogger.ErrorContext(ctx, "Failed to respond", "Error", err)
	}
}

func handleListReports(session *discordgo.Session, interaction *discordgo.InteractionCreate, ctx context.Context) {
	slogger, ok := ctxutil.Value[*slog.Logger](ctx)
	if !ok {
		slogger = slog.Default()
	}

	options := interaction.ApplicationCommandData().Options[0].Options
	optionMap := make(map[string]*discordgo.ApplicationCommandInteractionDataOption, len(options))
	for _, opt := range options {
		optionMap[opt.Name] = opt
	}

	reportedUserName := ""
	reportedUserOption, ok := optionMap["reported-user"]
	if ok {
		reportedUserName = reportedUserOption.UserValue(session).Username
	}

	reporterUserName := ""
	reporterUserOption, ok := optionMap["reporter-user"]
	if ok {
		reporterUserName = reporterUserOption.UserValue(session).Username
	}

	responseContent := fmt.Sprintf("Reported user: %s; Reporter user: %s", reportedUserName, reporterUserName)

	if err := session.InteractionRespond(interaction.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: responseContent,
		},
	}); err != nil {
		slogger.ErrorContext(ctx, "Failed to respond", "Error", err)
	}
}

func handleDeleteReport(session *discordgo.Session, interaction *discordgo.InteractionCreate, ctx context.Context) {
	slogger, ok := ctxutil.Value[*slog.Logger](ctx)
	if !ok {
		slogger = slog.Default()
	}

	options := interaction.ApplicationCommandData().Options[0].Options
	optionMap := make(map[string]*discordgo.ApplicationCommandInteractionDataOption, len(options))
	for _, opt := range options {
		optionMap[opt.Name] = opt
	}

	reportIdOption, ok := optionMap["report-id"]
	if !ok {
		slogger.ErrorContext(ctx, "Failed to get reported user option")
		return
	}

	responseContent := fmt.Sprintf("Delete report %s", reportIdOption.StringValue())

	if err := session.InteractionRespond(interaction.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: responseContent,
		},
	}); err != nil {
		slogger.ErrorContext(ctx, "Failed to respond", "Error", err)
	}
}

func CreateReportCommandHandler(botconfig botconfig.BotConfig) SlashCommandHandlerFunc {
	backendURL, err := botconfig.BackendURL()
	if err != nil {
		log.Fatal(backendURL)
	}

	return func(session *discordgo.Session, interaction *discordgo.InteractionCreate, ctx context.Context) {
		slogger, ok := ctxutil.Value[*slog.Logger](ctx)
		if !ok {
			slogger = slog.Default()
		}

		options := interaction.ApplicationCommandData().Options

		switch options[0].Name {
		case "new":
			handleNewReport(session, interaction, ctx)
		case "list":
			handleListReports(session, interaction, ctx)
		case "delete":
			handleDeleteReport(session, interaction, ctx)
		default:
			slogger.ErrorContext(ctx, "Invalid subcommand", "Subcommand Name", options[0].Name)
		}
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

		requestStruct := &RegistrationRequest{ServerID: serverId, UserID: userId}

		requestBody, err := json.Marshal(requestStruct)
		if err != nil {
			log.Print(err)
			return
		}

		requestURL := backendURL.JoinPath("databases")
		request, err := http.NewRequestWithContext(ctx, "POST", requestURL.String(), bytes.NewBuffer(requestBody))
		if err != nil {
			slogger.ErrorContext(ctx, "Backend Request Creation", "Error", err)
			return
		}

		response, err := session.Client.Do(request)
		if err != nil {
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
