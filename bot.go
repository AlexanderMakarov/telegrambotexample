package main

import (
	"context"
	"fmt"
	"log"
	"math/rand"
	"os"
	"strconv"
	"strings"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/joho/godotenv"
)

func loadDotEnv() {
	err := godotenv.Load()
	if err != nil {
		log.Fatalf("Can't find .env file nearby to get secrets.")
	}
}

func main() {
	loadDotEnv()
	bot, err := tgbotapi.NewBotAPI(os.Getenv("TELEGRAM_APITOKEN"))
	if err != nil {
		log.Fatalf("Can't connect to Telegram: %v", err)
	}
	bot.Debug, err = strconv.ParseBool(os.Getenv("TELEGRAM_DEBUG"))
	if err != nil {
		log.Fatalf("Can't read TELEGRAM_DEBUG environment variable: %v", err)
	}

	updateConfig := tgbotapi.NewUpdate(0)
	updateConfig.Timeout = 30
	ctx := context.Background()
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	// Start bot in separate goroutine.
	updates := bot.GetUpdatesChan(updateConfig)
	go receiveUpdates(ctx, updates, bot)
	log.Println("Start listening for updates. Press Enter key to stop.")

	// Wait Enter to stop bot.
	fmt.Scanln()
	log.Println("Bot stopped.")
	os.Exit(0)
}

func receiveUpdates(ctx context.Context, updates tgbotapi.UpdatesChannel, bot *tgbotapi.BotAPI) {
	for {
		select {
		case <-ctx.Done(): // Stop if someone ("main") stopped context.
			return
		case update := <-updates:
			handleUpdate(update, bot)
		}
	}
}

var (
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

func handleUpdate(update tgbotapi.Update, bot *tgbotapi.BotAPI) {
	switch {
	case update.Message != nil: // Messages.
		handleMessage(update.Message, bot)
		break
	case update.CallbackQuery != nil: // Buttons (aka InlineKeyboard events).
		handleButton(update.CallbackQuery, bot)
		break
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
			rand.Seed(time.Now().UnixNano())
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
