package bot

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"math"
	"strconv"
	"strings"
	"time"

	htmlgenerator "github.com/BalanceBalls/report-generator/internal/generator/html"
	"github.com/BalanceBalls/report-generator/internal/gitlab"
	"github.com/BalanceBalls/report-generator/internal/logger"
	"github.com/BalanceBalls/report-generator/internal/report"
	"github.com/BalanceBalls/report-generator/internal/storage"
	"github.com/BalanceBalls/report-generator/internal/storage/postgres"

	tg "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type ReportsBot struct {
	Bot tg.BotAPI

	config    *Config
	storage   Storage
	builder   Builder
	generator Generator
}

const empty = ""

// commands
const (
	helpCmd    = "help"
	regCmd     = "reg"
	unregCmd   = "unreg"
	genDayCmd  = "gen_day"
	startCmd   = "start"
	profileCmd = "profile"
)

// user input prefixes
const (
	setTokenPrefix    = "token:"
	setOffsetPrefix   = "offset:"
	setGitlabIdPrefix = "id:"
)

func New(cfg *Config) *ReportsBot {
	bot, err := tg.NewBotAPI(cfg.BotToken)
	if err != nil {
		panic(err)
	}

	pgConnString := cfg.GetPostgresConnectionString()
	pgSql, err := postgres.New(pgConnString)
	if err != nil {
		panic(err)
	}

	html := htmlgenerator.New(cfg.ReportFileDir, cfg.ReportTemplate, cfg.GenerateFile)
	gitlabClient := gitlab.NewClient(cfg.GitHost, cfg.GitBasePath)
	reportBuilder := gitlab.NewReportBuilder(*gitlabClient)

	return &ReportsBot{
		Bot: *bot,

		config:    cfg,
		storage:   pgSql,
		generator: html,
		builder:   reportBuilder,
	}
}

func (b *ReportsBot) Serve(ctx context.Context) {
	slog.Info("bot authorized to telegram", "user", b.Bot.Self.UserName)

	if err := b.storage.Up(ctx); err != nil {
		slog.ErrorContext(ctx, err.Error())
		panic(err)
	}
	updateConfig := tg.NewUpdate(0)
	updateConfig.Timeout = b.config.CommandsTimeout
	updates := b.Bot.GetUpdatesChan(updateConfig)

	slog.Info("bot is now ready to serve commands")

	for update := range updates {
		// ignore any non-Message updates
		if update.Message == nil {
			continue
		}

		updateCtx, cancel := context.WithTimeout(ctx, time.Second*time.Duration(b.config.CommandsTimeout))
		defer cancel()

		commandLogger := slog.With(
			slog.Group("context",
				slog.Int("trace_id", update.UpdateID),
				slog.Int64("tg_chat_id", update.Message.Chat.ID),
				slog.Int64("tg_user_id", update.Message.From.ID),
				slog.String("text", update.Message.Text),
			))
		updateCtx = logger.AddToContext(updateCtx, commandLogger)

		// Handling user input
		if !update.Message.IsCommand() {
			userId := update.Message.From.ID
			chatId := update.Message.Chat.ID

			userInput := update.Message.Text
			dbUser, err := b.storage.User(updateCtx, userId)

			if err != nil {
				commandLogger.ErrorContext(updateCtx, "failed to get user info", "reason", err)
				if errors.Is(err, storage.ErrUserNotFound) {
					b.sendText(userNotRegisteredMsg, chatId); continue
				}
			}

			if strings.HasPrefix(userInput, setTokenPrefix) {
				go b.setUserToken(updateCtx, userInput, chatId, dbUser); continue
			} else if strings.HasPrefix(userInput, setOffsetPrefix) {
				go b.setUserTimezoneOffset(updateCtx, userInput, chatId, dbUser); continue
			} else if strings.HasPrefix(userInput, setGitlabIdPrefix) {
				go b.setUserGitlabId(updateCtx, userInput, chatId, dbUser); continue
			} else {
				commandLogger.WarnContext(updateCtx, "user input was not recognized", "input", userInput)
			}

			continue
		}

		chatId := update.Message.Chat.ID
		userId := update.Message.From.ID

		// Extract the command from the Message.
		switch update.Message.Command() {
		case startCmd:
			commandLogger.InfoContext(updateCtx, "/start cmd received")
			go b.sendText(helloMsg, chatId)
		case helpCmd:
			commandLogger.InfoContext(updateCtx, "/help cmd received")
			go b.sendText(helpMsg, chatId)
		case regCmd:
			commandLogger.InfoContext(updateCtx, "/reg cmd received")
			go b.handleRegistration(updateCtx, userId, chatId)
		case unregCmd:
			commandLogger.InfoContext(updateCtx, "/unreg cmd received")
			go b.handleUnregistration(updateCtx, userId, chatId)
		case genDayCmd:
			commandLogger.InfoContext(updateCtx, "/genDay cmd received")
			go b.handleReportGeneration(updateCtx, userId, chatId)
		case profileCmd:
			commandLogger.InfoContext(updateCtx, "/profile cmd received")
			go b.handleProfileInfo(updateCtx, userId, chatId)
		default:
			commandLogger.WarnContext(updateCtx, "command was not recognized")
		}
	}
}

func (b *ReportsBot) handleReportGeneration(ctx context.Context, userId int64, chatId int64) {
	logger := logger.GetFromContext(ctx)
	user, err := b.storage.User(ctx, userId)

	if err != nil {
		logger.ErrorContext(ctx, "report generation failed", "reason", err)
		if errors.Is(err, storage.ErrUserNotFound) {
			b.sendText(userNotRegisteredMsg, chatId)
			return
		}
		b.sendText(reportGenerationFailedMsg, chatId)
		return
	}

	if user.GitlabId == 0 {
		logger.ErrorContext(ctx, "gitlab id is not set for user")
		b.sendText(gitlabIdNotSetErrorMsg, chatId)
		return
	}

	if user.UserToken == empty {
		logger.ErrorContext(ctx, "user token is not set for user")
		b.sendText(tokenNotSetErrorMsg, chatId)
		return
	}

	b.sendText(reportInProgressMsg, chatId)

	respch := make(chan report.Channel)
	go b.builder.Build(ctx, user, respch)

	select {
	case <-ctx.Done():
		logger.ErrorContext(ctx, "update cancelled", "reason", ctx.Err())
	case reportData := <-respch:
		b.processReportResult(ctx, reportData, chatId, user.Id)
	}
}

func (b *ReportsBot) processReportResult(ctx context.Context, reportData report.Channel, chatId int64, userId int64) {
	logger := logger.GetFromContext(ctx)

	if reportData.Err != nil {
		logger.ErrorContext(ctx, "failed to get report data", "reason", reportData.Err)
		if errors.Is(reportData.Err, gitlab.ErrNoGitActions) {
			b.sendText(emptyReportMsg, chatId)
			return
		}
		b.sendText(reportGenerationFailedMsg, chatId)
		return
	}

	if err := b.storage.SaveReport(ctx, reportData.Report, userId); err != nil {
		logger.ErrorContext(ctx, "failed to save report to DB", "reason", err)
	}

	reportBytes, err := b.generator.Generate(reportData.Report)
	if err != nil {
		logger.ErrorContext(ctx, "report generation failed", "reason", err)
		return
	}

	file := tg.FileBytes{
		Name:  reportBytes.Name,
		Bytes: reportBytes.Data,
	}

	msg := tg.NewDocument(chatId, file)
	msg.Caption = reportFileCaption
	if _, err = b.Bot.Send(msg); err != nil {
		logger.ErrorContext(ctx, "failed to send report", "reason", err.Error())
	}
}

func (b *ReportsBot) handleRegistration(ctx context.Context, userId int64, chatId int64) {
	logger := logger.GetFromContext(ctx)
	alreadyRegistered := b.storage.UserExists(ctx, userId)
	if alreadyRegistered {
		b.sendText(userAlreadyRegisteredMsg, chatId)
		return
	}

	newUser := report.User{
		Id:             userId,
		GitlabId:       0,
		UserEmail:      empty,
		UserToken:      empty,
		IsActive:       true,
		TimezoneOffset: 0,
	}

	if err := b.storage.AddUser(ctx, newUser); err != nil {
		logger.ErrorContext(ctx, "failed to add new user", "reason", err)
		b.sendText(userRegistrationErrorMsg, chatId)

		return
	}

	b.sendText(userHasBeenRegisteredMsg, chatId)
}

func (b *ReportsBot) handleProfileInfo(ctx context.Context, userId int64, chatId int64) {
	logger := logger.GetFromContext(ctx)
	user, err := b.storage.User(ctx, userId)

	if err != nil {
		logger.ErrorContext(ctx, "failed to fetch user info", "reason", err)
		if errors.Is(err, storage.ErrUserNotFound) {
			b.sendText(userNotRegisteredMsg, chatId)
			return
		}
		b.sendText(fetchUserInfoFailedMsg, chatId)
		return
	}

	var tokenMsg string
	if user.UserToken == empty {
		tokenMsg = tokenIsNotSetMsg
	} else {
		tokenMsg = tokenIsSetMsg
	}

	responseMsg := fmt.Sprintf(profileCmdTemplate, user.TimezoneOffset, user.GitlabId, tokenMsg)

	b.sendText(responseMsg, chatId)
}

func (b *ReportsBot) handleUnregistration(ctx context.Context, userId int64, chatId int64) {
	logger := logger.GetFromContext(ctx)
	isRegistered := b.storage.UserExists(ctx, userId)

	if isRegistered {
		if err := b.storage.RemoveUser(ctx, userId); err != nil {
			logger.ErrorContext(ctx, "failed to remove user", "reason", err)
			b.sendText(userDataUpdateErrorMsg, chatId)
		}
		b.sendText(userHasBeenRemovedMsg, chatId)
		return
	}

	b.sendText(userNotRegisteredMsg, chatId)
}

func (b *ReportsBot) setUserGitlabId(ctx context.Context, userInput string, chatId int64, dbUser report.User) {
	logger := logger.GetFromContext(ctx)
	inputGitlabId := strings.TrimPrefix(userInput, setGitlabIdPrefix)
	updatedGitlabId, err := strconv.Atoi(inputGitlabId)

	if err != nil {
		logger.ErrorContext(ctx, "could not parse user input for gitlab id", "reason", err)
		b.sendText(gitlabIdBadInputErrorMsg, chatId)
		return
	}

	dbUser.GitlabId = updatedGitlabId

	if err := b.storage.UpdateUser(ctx, dbUser); err != nil {
		logger.ErrorContext(ctx, "failed to update user's gitlab id", "reason", err)
		b.sendText(userDataUpdateErrorMsg, chatId)
		return
	}
	logger.InfoContext(ctx, "gitlab id updated successfully")
	b.sendText(gitlabIdHasBeenSavedMsg, chatId)
}

func (b *ReportsBot) setUserToken(ctx context.Context, userInput string, chatId int64, dbUser report.User) {
	logger := logger.GetFromContext(ctx)
	updatedToken := strings.TrimPrefix(userInput, setTokenPrefix)
	dbUser.UserToken = updatedToken

	if err := b.storage.UpdateUser(ctx, dbUser); err != nil {
		logger.ErrorContext(ctx, "failed to update user's token", "reason", err)
		b.sendText(userDataUpdateErrorMsg, chatId)
		return
	}
	logger.InfoContext(ctx, "user token updated successfully")
	b.sendText(tokenHasBeenSavedMsg, chatId)
}

func (b *ReportsBot) setUserTimezoneOffset(ctx context.Context, userInput string, chatId int64, dbUser report.User) {
	logger := logger.GetFromContext(ctx)
	inputOffset := strings.TrimPrefix(userInput, setOffsetPrefix)
	updatedOffset, err := strconv.ParseInt(inputOffset, 10, 64)

	if math.Abs(float64(updatedOffset)) > 720 {
		logger.ErrorContext(ctx, "offset is larger than half a day")
	}

	if err != nil {
		logger.ErrorContext(ctx, "failed to parse user input", "reason", err)
		b.sendText(userDataUpdateErrorMsg, chatId)
		return
	}

	dbUser.TimezoneOffset = int(updatedOffset)
	if err := b.storage.UpdateUser(ctx, dbUser); err != nil {
		logger.ErrorContext(ctx, "failed to update user's timezone offset", "reason", err)
		b.sendText(userDataUpdateErrorMsg, chatId)
		return
	}

	logger.InfoContext(ctx, "user timezone offset updated successfully")
	b.sendText(timezoneHasBeenSavedMsg, chatId)
}

func (b *ReportsBot) sendText(text string, chatId int64) {
	message := tg.NewMessage(int64(chatId), empty)
	message.Text = text

	if _, err := b.Bot.Send(message); err != nil {
		slog.Error("failed to send a message to bot", "reason", err)
	}
}
