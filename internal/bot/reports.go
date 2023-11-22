package bot

import (
	"context"
	"errors"
	"log"
	"strconv"
	"strings"

	"github.com/BalanceBalls/report-generator/internal/builder"
	"github.com/BalanceBalls/report-generator/internal/clients/gitlab"
	"github.com/BalanceBalls/report-generator/internal/generator"
	htmlgenerator "github.com/BalanceBalls/report-generator/internal/generator/html"
	"github.com/BalanceBalls/report-generator/internal/storage"
	"github.com/BalanceBalls/report-generator/internal/storage/sqlite"
	tg "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type ReportsBot struct {
	Bot tg.BotAPI

	storage   storage.Storage
	generator generator.Generator
	gitClient gitlab.GitlabClient
}

const empty = ""

// commands
const (
	helpCmd   = "help"
	regCmd    = "reg"
	unregCmd  = "unreg"
	genDayCmd = "gen_day"
	startCmd  = "start"
)

// user input prefixes
const (
	setTokenPrefix  = "token:"
	setOffsetPrefix = "offset:"
)

func New(token string) *ReportsBot {
	bot, err := tg.NewBotAPI(token)
	if err != nil {
		panic(err)
	}

	sqlite, err := sqlite.New("bot.sqlite")
	if err != nil {
		panic(err)
	}

	html := htmlgenerator.New("./reports", "html_report.tmpl", true)

	gitHost := "localhost:4443"
	gitBasePath := "api/v4"
	gitlabClient := gitlab.New(gitHost, gitBasePath)

	return &ReportsBot{
		Bot: *bot,

		storage:   sqlite,
		generator: html,
		gitClient: *gitlabClient,
	}
}

func (b *ReportsBot) Serve(ctx context.Context) {
	log.Printf("authorized on account %s", b.Bot.Self.UserName)

	updateConfig := tg.NewUpdate(0)
	updateConfig.Timeout = 60
	updates := b.Bot.GetUpdatesChan(updateConfig)

	for update := range updates {
		// ignore any non-Message updates
		if update.Message == nil {
			continue
		}

		// Handling user input
		if !update.Message.IsCommand() {
			userId := update.Message.From.ID
			chatId := update.Message.Chat.ID

			userInput := update.Message.Text
			dbUser, err := b.storage.User(ctx, userId)

			if err != nil {
				log.Print(err)
				if errors.Is(err, storage.ErrUserNotFound) {
					b.sendText(userNotRegisteredMsg, chatId)
					continue
				}
			}

			if strings.HasPrefix(userInput, setTokenPrefix) {
				b.setUserToken(ctx, userInput, chatId, *dbUser)
			} else if strings.HasPrefix(userInput, setOffsetPrefix) {
				b.setUserTimezoneOffset(ctx, userInput, chatId, *dbUser)
			} else {
				log.Printf("user input was not recognized: %s", userInput)
			}

			continue
		}

		chatId := update.Message.Chat.ID
		userId := update.Message.From.ID

		// Extract the command from the Message.
		switch update.Message.Command() {
		case startCmd:
			b.sendText(helloMsg, chatId)
		case helpCmd:
			b.sendText(helpMsg, chatId)
		case regCmd:
			b.handleRegistration(ctx, userId, chatId)
		case unregCmd:
			b.handleUnregistration(ctx, userId, chatId)
		case genDayCmd:
			user, err := b.storage.User(ctx, userId)

			if err != nil {
				log.Print(err)
			}

			reportBuilder := builder.New(b.gitClient, user.Id, user.UserToken, user.TimezoneOffset)
			respch := make(chan builder.BuildResult)
			go reportBuilder.Build(ctx, respch)

			select {
			case <-ctx.Done():
				log.Print(ctx.Err())
			case reportData := <-respch:
				reportBytes, err := b.generator.Generate(reportData.Report)
				if err != nil {
					log.Print(err)
					continue
				}

				file := tg.FileBytes{
					Name:  reportBytes.Name,
					Bytes: reportBytes.Data,
				}
				
				msg := tg.NewDocument(chatId, file)
				msg.Caption = "Отчет за сегодняшний день"
				b.Bot.Send(msg)
			}
			b.sendText("Репорты пока в разработке", chatId)
		default:
			log.Print("command was not recognized:", update.Message.Command())
		}
	}
}

func (b *ReportsBot) handleRegistration(ctx context.Context, userId int64, chatId int64) {
	alreadyRegistered := b.storage.UserExists(ctx, userId)
	if alreadyRegistered {
		b.sendText(userAlreadyRegisteredMsg, chatId)
		return
	}

	newUser := storage.User{
		Id:             userId,
		UserEmail:      "test@q.com",
		UserToken:      "qweqweqwe",
		IsActive:       true,
		TimezoneOffset: 200,
	}

	if err := b.storage.AddUser(ctx, newUser); err != nil {
		log.Print("failed to add new user", err)
		b.sendText(userRegistrationErrorMsg, chatId)

		return
	}

	b.sendText(userHasBeenRegisteredMsg, chatId)
}

func (b *ReportsBot) handleUnregistration(ctx context.Context, userId int64, chatId int64) {
	isRegistered := b.storage.UserExists(ctx, userId)

	if isRegistered {
		if err := b.storage.RemoveUser(ctx, userId); err != nil {
			log.Print(err)
			b.sendText(userDataUpdateErrorMsg, chatId)
		}
		b.sendText(userHasBeenRemovedMsg, chatId)
		return
	}

	b.sendText(userNotRegisteredMsg, chatId)
}

func (b *ReportsBot) setUserToken(ctx context.Context, userInput string, chatId int64, dbUser storage.User) {
	updatedToken := strings.TrimPrefix(userInput, setTokenPrefix)
	dbUser.UserToken = updatedToken

	if err := b.storage.UpdateUser(ctx, dbUser); err != nil {
		log.Print(err)
		b.sendText(userDataUpdateErrorMsg, chatId)
	}
	b.sendText(tokenHasBeenSavedMsg, chatId)
}

func (b *ReportsBot) setUserTimezoneOffset(ctx context.Context, userInput string, chatId int64, dbUser storage.User) {
	inputOffset := strings.TrimPrefix(userInput, setOffsetPrefix)
	updatedOffset, err := strconv.ParseInt(inputOffset, 10, 64)

	if err != nil {
		log.Print("failed to parse user input: ", err)
		b.sendText(userDataUpdateErrorMsg, chatId)
	}

	dbUser.TimezoneOffset = int(updatedOffset)
	if err := b.storage.UpdateUser(ctx, dbUser); err != nil {
		log.Print(err)
		b.sendText(userDataUpdateErrorMsg, chatId)
	}
	b.sendText(timezoneHasBeenSavedMsg, chatId)
}

func (b *ReportsBot) sendText(text string, chatId int64) {
	message := tg.NewMessage(int64(chatId), empty)
	message.Text = text

	if _, err := b.Bot.Send(message); err != nil {
		log.Panic(err)
	}
}
