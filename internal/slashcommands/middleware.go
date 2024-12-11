package slashcommands

import (
	"context"
	"log/slog"
	"os"
	"runtime/debug"
	"time"

	"github.com/bwmarrin/discordgo"
	"snitch/snitchbot/pkg/ctxutil"
)

func WithTimeout(next SlashCommandHandlerFunc, duration time.Duration) SlashCommandHandlerFunc {
	return func(session *discordgo.Session, interaction *discordgo.InteractionCreate, ctx context.Context) {
		timeoutCtx, cancel := context.WithTimeout(ctx, duration)
		defer cancel()
		next(session, interaction, timeoutCtx)
	}
}

func Log(next SlashCommandHandlerFunc) SlashCommandHandlerFunc {
	return func(session *discordgo.Session, interaction *discordgo.InteractionCreate, ctx context.Context) {
		slogger := slog.New(slog.NewTextHandler(os.Stdout, nil)).With(
			slog.String("User ID", interaction.Member.User.ID),
			slog.String("Guild ID", interaction.GuildID),
			slog.String("Command", interaction.ApplicationCommandData().Name),
		)

		ctx = ctxutil.WithValue(ctx, slogger)
		next(session, interaction, ctx)
	}
}

func ResponseTime(next SlashCommandHandlerFunc) SlashCommandHandlerFunc {
	return func(session *discordgo.Session, interaction *discordgo.InteractionCreate, ctx context.Context) {
		slogger, ok := ctxutil.Value[*slog.Logger](ctx)
		if !ok {
			slogger = slog.Default()
		}

		start := time.Now()
		next(session, interaction, ctx)
		elapsed := time.Since(start)

		slogger.InfoContext(ctx, "Command Time", "Time Elapsed", elapsed)
	}
}

func Recovery(next SlashCommandHandlerFunc) SlashCommandHandlerFunc {
	return func(session *discordgo.Session, interaction *discordgo.InteractionCreate, ctx context.Context) {
		defer func() {
			if err := recover(); err != nil {
				slogger, ok := ctxutil.Value[*slog.Logger](ctx)
				if !ok {
					slogger = slog.Default()
				}
				stack := debug.Stack()

				slogger.ErrorContext(ctx, "Recovery", "Panic", err, "Stack", stack)
			}
		}()
		next(session, interaction, ctx)
	}
}
