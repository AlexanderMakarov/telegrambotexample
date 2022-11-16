package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"

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

	// Start polling Telegram for updates.
	// updates := bot.GetUpdatesChan(updateConfig)
	// for update := range updates {
	// 	if update.Message == nil {
	// 		continue
	// 	}

	// 	msg := tgbotapi.NewMessage(update.Message.Chat.ID, update.Message.Text)
	// 	msg.ReplyToMessageID = update.Message.MessageID

	// 	if _, err := bot.Send(msg); err != nil {
	// 		panic(err)
	// 	}
	// }
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
	`
	commands = map[string]func(chatId int64, bot *tgbotapi.BotAPI) error{
		"help": sendHelp,
		"menu": sendMenu,
	}

	// Menu texts
	firstMenu  = "<b>Menu 1</b>\n\nA beautiful menu with a shiny inline button."
	secondMenu = "<b>Menu 2</b>\n\nA better menu with even more shiny inline buttons."

	// Button texts
	nextButton     = "Next"
	backButton     = "Back"
	tutorialButton = "Tutorial"

	// Keyboard layout for the first menu. One button, one row
	firstMenuMarkup = tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData(nextButton, nextButton),
		),
	)

	// Keyboard layout for the second menu. Two buttons, one per row
	secondMenuMarkup = tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData(backButton, backButton),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonURL(tutorialButton, "https://core.telegram.org/bots/api"),
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

	if user == nil { // Answer only real users.
		return
	}

	log.Printf("%s wrote %s", user.FirstName, text)

	var err error
	if strings.HasPrefix(text, "/") {
		handler, ok := commands[text]
		if !ok {
			msg := tgbotapi.NewMessage(chatId, fmt.Sprintf("Command '%s' is unknown.\n%s", text, help))
			_, err := bot.Send(msg)
		} else {
			err = handler(chatId, bot)
		}
	} else {
		msg := tgbotapi.NewMessage(message.Chat.ID, fmt.Sprintf("Got it at %v", message.Time()))
		msg.ReplyToMessageID = message.MessageID
		_, err := bot.Send(msg)
		// copyMsg := tgbotapi.NewCopyMessage(message.Chat.ID, message.Chat.ID, message.MessageID)
		// _, err = bot.CopyMessage(copyMsg)
	}

	if err != nil {
		log.Printf("Can't handle input message %v: %s", message, err.Error())
	}
}

func handleButton(query *tgbotapi.CallbackQuery, bot *tgbotapi.BotAPI) {
	var text string

	markup := tgbotapi.NewInlineKeyboardMarkup()
	message := query.Message

	if query.Data == nextButton {
		text = secondMenu
		markup = secondMenuMarkup
	} else if query.Data == backButton {
		text = firstMenu
		markup = firstMenuMarkup
	}

	callbackCfg := tgbotapi.NewCallback(query.ID, "")
	bot.Send(callbackCfg)

	// Replace menu text and keyboard
	msg := tgbotapi.NewEditMessageTextAndMarkup(message.Chat.ID, message.MessageID, text, markup)
	msg.ParseMode = tgbotapi.ModeHTML
	bot.Send(msg)
}

func sendMenu(chatId int64, bot *tgbotapi.BotAPI) error {
	msg := tgbotapi.NewMessage(chatId, firstMenu)
	msg.ParseMode = tgbotapi.ModeHTML
	msg.ReplyMarkup = firstMenuMarkup
	_, err := bot.Send(msg)
	return err
}

func sendHelp(chatId int64, bot *tgbotapi.BotAPI) error {
	msg := tgbotapi.NewMessage(chatId, help)
	_, err := bot.Send(msg)
	return err
}
