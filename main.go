package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"time"

	"github.com/go-telegram/bot"
)

var storage *Storage

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()

	if err := InitDB("bot.db"); err != nil {
		log.Fatalf("failed to initialize database: %v", err)
	}
	defer CloseDB()

	storage = NewStorage(DB)

	opts := []bot.Option{
		bot.WithDefaultHandler(defaultHandler),
	}

	b, err := bot.New(os.Getenv("TG_BOT_TOKEN"), opts...)
	if err != nil {
		panic(err)
	}

	b.RegisterHandler(bot.HandlerTypeMessageText, "/start", bot.MatchTypeExact, startHandler)
	b.RegisterHandler(bot.HandlerTypeMessageText, "/add", bot.MatchTypePrefix, addHandler)
	b.RegisterHandler(bot.HandlerTypeMessageText, "/list", bot.MatchTypeExact, listHandler)
	b.RegisterHandler(bot.HandlerTypeMessageText, "/news", bot.MatchTypeExact, newsHandler)
	b.RegisterHandler(bot.HandlerTypeMessageText, "/remove", bot.MatchTypePrefix, removeHandler)

	go startScheduler(ctx, b, 30*time.Minute)

	b.Start(ctx)
}
