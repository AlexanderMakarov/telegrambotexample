package telegrambot

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/joho/godotenv"

	"github.com/GoogleCloudPlatform/functions-framework-go/functions"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

var (
	bot  *tgbotapi.BotAPI
	port int16

	help = `Hi. Bot supports following commands:
	/help - prints this help.
	/menu - shows menu.
	Also it rephrases eny string you sent to it.
	`
	commands = map[string]func(chatId int64, bot *tgbotapi.BotAPI) error{
		"help": sendHelp,
		"menu": sendMenu,
	}

	// Menu texts
	menuText = "<b>Menu 1</b>\n\nA beautiful menu with a shiny inline buttons."

	// Button texts
	siteButton     = "Site"
	tutorialButton = "Tutorial"
	closeButton    = "Close"

	menuMarkup = tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonURL(siteButton, "https://core.telegram.org/bots/api"),
			tgbotapi.NewInlineKeyboardButtonData(tutorialButton, tutorialButton),
		),
	)
)

func init() {

	// If executing not in GCP then need to load environment variables from .env file.
	// https://stackoverflow.com/a/61692563/1535127
	if _, ok := os.LookupEnv("X_GOOGLE_FUNCTION_REGION"); !ok {
		log.Println("Running not on GCP, loading '.env' file.")
		loadDotEnv()
	}

	telegramApiToken := os.Getenv("TELEGRAM_APITOKEN")
	if len(telegramApiToken) <= 0 {
		log.Fatal("Can't find TELEGRAM_APITOKEN environment variable.")
	}
	telegramDebug := false
	var err error
	telegramDebug, err = strconv.ParseBool(os.Getenv("TELEGRAM_DEBUG"))
	if err != nil {
		log.Printf("Can't read TELEGRAM_DEBUG environment variable, will use FALSE value: %v", err)
	}
	log.Printf("Got from environment len(TELEGRAM_APITOKEN)=%d, DEBUG=%t", len(telegramApiToken), telegramDebug)

	// Create Telegram Bot instance.
	bot, err = tgbotapi.NewBotAPI(telegramApiToken)
	if err != nil {
		log.Fatalf("Can't connect to Telegram: %v", err)
	}
	bot.Debug = telegramDebug
	log.Println("Initialized NewBotAPI")

	// Add seed to "rand" module.
	rand.Seed(time.Now().UnixNano())

	// Check for errors left during bot inactivtiy.
	info, err := bot.GetWebhookInfo()
	if err != nil {
		log.Fatal(err)
	}
	if info.LastErrorDate != 0 {
		log.Printf("Telegram last fail on delivering webhhook at %v: %s", info.LastErrorDate, info.LastErrorMessage)
	}

	// Start serve locally (onlyf for local).
	functions.HTTP("HandleTelegramWebHook", HandleTelegramWebHook)
	log.Println("Initialization is completed.")
}

func loadDotEnv() {
	err := godotenv.Load()
	if err != nil {
		log.Fatalf("Can't find .env file nearby to get secrets.")
	}
}

func StartPollingLocally() {
	updateConfig := tgbotapi.NewUpdate(0)
	updateConfig.Timeout = 30
	ctx := context.Background()
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	// Start bot in separate goroutine.
	updates := bot.GetUpdatesChan(updateConfig)
	go receiveUpdates(ctx, updates)
	log.Println("Start listening for updates. Press Enter key to stop.")

	// Wait Enter to stop bot.
	fmt.Scanln()
	log.Println("Bot stopped.")
	os.Exit(0)
}

// Entry point for "webhook" connection.
func HandleTelegramWebHook(w http.ResponseWriter, r *http.Request) {
	var update tgbotapi.Update
	if err := json.NewDecoder(r.Body).Decode(&update); err != nil {
		log.Printf("Could not decode incoming update %s", err.Error())
	}
	handleUpdate(update, bot)
}

// Entry point for "polling" connection.
func receiveUpdates(ctx context.Context, updates tgbotapi.UpdatesChannel) {
	for {
		select {
		case <-ctx.Done(): // Stop if someone ("main") stopped context.
			return
		case update := <-updates:
			handleUpdate(update, bot)
		}
	}
}

func handleUpdate(update tgbotapi.Update, bot *tgbotapi.BotAPI) {
	switch {
	case update.Message != nil: // Messages.
		handleMessage(update.Message, bot)
		break
	case update.CallbackQuery != nil: // Buttons (aka InlineKeyboard events).
		handleButton(update.CallbackQuery, bot)
		break
	default:
		log.Printf("Received unsupported update: %v", update)
	}
}

func handleMessage(message *tgbotapi.Message, bot *tgbotapi.BotAPI) {
	user := message.From
	text := message.Text
	chatId := message.Chat.ID

	if user == nil { // Answer only to real users.
		return
	}

	log.Printf("%s wrote %s", user.FirstName, text)

	var err error
	if strings.HasPrefix(text, "/") {
		handler, ok := commands[text[1:]]
		if !ok {
			msg := tgbotapi.NewMessage(chatId, fmt.Sprintf("Command '%s' is unknown.\n%s", text, help))
			_, err = bot.Send(msg)
		} else {
			err = handler(chatId, bot)
		}
	} else {
		words := strings.Split(text, " ")
		if len(words) > 1 {
			rand.Shuffle(len(words), func(i, j int) { words[i], words[j] = words[j], words[i] })
			text = strings.Join(words, " ")
		}

		msg := tgbotapi.NewMessage(message.Chat.ID, fmt.Sprintf("At %v:\n%s", message.Time(), text))
		msg.ReplyToMessageID = message.MessageID
		_, err = bot.Send(msg)
	}

	if err != nil {
		log.Printf("Can't handle input message %v: %s", message, err.Error())
	}
}

func handleButton(query *tgbotapi.CallbackQuery, bot *tgbotapi.BotAPI) {
	if query.Data == tutorialButton {
		err := sendHelp(query.Message.Chat.ID, bot)
		if err != nil {
			log.Printf("Can't send help: %s", err.Error())
		}
	}

	// Answer to Telegram that button was handled.
	callbackCfg := tgbotapi.NewCallback(query.ID, "")
	bot.Send(callbackCfg)
}

func sendMenu(chatId int64, bot *tgbotapi.BotAPI) error {
	msg := tgbotapi.NewMessage(chatId, menuText)
	msg.ParseMode = tgbotapi.ModeHTML
	msg.ReplyMarkup = menuMarkup
	_, err := bot.Send(msg)
	return err
}

func sendHelp(chatId int64, bot *tgbotapi.BotAPI) error {
	msg := tgbotapi.NewMessage(chatId, help)
	_, err := bot.Send(msg)
	return err
}
