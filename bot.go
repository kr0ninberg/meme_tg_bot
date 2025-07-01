package main

import (
	"context"
	"database/sql"
	"fmt"
	_ "github.com/lib/pq"
	"log"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

func main() {
	cfg, err := loadConfig()
	if err != nil {
		log.Fatalf("config: %v", err)
	}

	db, err := initDB(cfg)
	if err != nil {
		log.Fatalf("db: %v", err)
	}
	defer db.Close()

	bot, err := tgbotapi.NewBotAPI(cfg.TelegramToken)
	if err != nil {
		log.Fatalf("bot: %v", err)
	}
	log.Printf("Logged in as @%s", bot.Self.UserName)

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	go handleUpdates(ctx, bot, db, cfg)

	<-ctx.Done()
	log.Println("graceful shutdown")
}

type Config struct {
	TelegramToken string
	MemeDir       string
	DBDSN         string
	AllowedChatID int64 // 0 = любой
}

func loadConfig() (*Config, error) {
	token := os.Getenv("TELEGRAM_TOKEN")
	dsn := os.Getenv("DATABASE_DSN")
	if token == "" || dsn == "" {
		return nil, fmt.Errorf("env TELEGRAM_TOKEN or DATABASE_DSN not set")
	}
	return &Config{
		TelegramToken: token,
		MemeDir:       "memes",
		DBDSN:         dsn,
	}, nil
}

func initDB(cfg *Config) (*sql.DB, error) {
	db, err := sql.Open("postgres", cfg.DBDSN)
	if err != nil {
		return nil, err
	}
	db.SetMaxOpenConns(5)
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(time.Hour)
	if err := db.Ping(); err != nil {
		return nil, err
	}
	return db, nil
}

func handleUpdates(ctx context.Context, bot *tgbotapi.BotAPI, db *sql.DB, cfg *Config) {
	keyboard := tgbotapi.NewReplyKeyboard(
		tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton("next"),
		),
	)

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 30
	updates := bot.GetUpdatesChan(u)

	for {
		select {
		case <-ctx.Done():
			return
		case upd := <-updates:
			if upd.Message != nil {
				if upd.Message.IsCommand() && upd.Message.Command() == "start" {
					msg := tgbotapi.NewMessage(upd.Message.Chat.ID, "Привет! Жми кнопку «next» 👇")
					msg.ReplyMarkup = keyboard
					bot.Send(msg)
					continue
				}
				filename, err := getRandomMeme(ctx, db)
				if err != nil {
					log.Printf("random meme: %v", err)
					continue
				}

				photo := tgbotapi.NewPhoto(upd.Message.Chat.ID,
					tgbotapi.FilePath(filepath.Join(cfg.MemeDir, filename)))
				photo.ReplyMarkup = keyboard
				bot.Send(photo)
			}
		}
	}
}

func getRandomMeme(ctx context.Context, db *sql.DB) (string, error) {
	var filename string
	err := db.QueryRowContext(ctx,
		"SELECT filename FROM memes_filenames ORDER BY RANDOM() LIMIT 1").
		Scan(&filename)
	return filename, err
}
