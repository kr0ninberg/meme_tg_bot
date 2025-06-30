package main

import (
	"database/sql"
	_ "github.com/lib/pq"
	"log"
	"os"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

func main() {
	botToken := os.Getenv("TELEGRAM_TOKEN") // Или вставь токен напрямую
	if botToken == "" {
		log.Panic("INVALID API KEY")
	}

	bot, err := tgbotapi.NewBotAPI(botToken)
	if err != nil {
		log.Panic(err)
	}

	log.Printf("Авторизован как %s", bot.Self.UserName)

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates := bot.GetUpdatesChan(u)

	keyboard := tgbotapi.NewReplyKeyboard(
		tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton("next"),
		),
	)
	connStr := "user=memes_user dbname=memes sslmode=disable password=memes_pass"
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	for update := range updates {
		if update.Message != nil {

			var filename string
			err = db.QueryRow("SELECT filename FROM memes_filenames ORDER BY RANDOM() LIMIT 1").Scan(&filename)
			if err != nil {
				log.Fatal(err)
			}

			photo := tgbotapi.NewPhoto(update.Message.Chat.ID, tgbotapi.FilePath("memes/"+filename))
			//photo.Caption = "Ты написал: " + update.Message.Text
			photo.ParseMode = "Markdown" // or "HTML" if needed
			photo.ReplyMarkup = keyboard
			bot.Send(photo)
		}
	}
}
