package main

import (
	"context"
)

type TelegramSender interface {
	SendMessage(ctx context.Context, chatID int64, text string) error
}

type OpenRouterClient interface {
	ChatCompletion(ctx context.Context, prompt string) (string, error)
}

type ResultSaver interface {
	SaveResult(ctx context.Context, chatID int64, data string) (int64, error)
}

// handleClaim processes user claim: sends prompt to OpenRouter, saves the result and sends it back to Telegram.
func handleClaim(ctx context.Context, tg TelegramSender, or OpenRouterClient, repo ResultSaver, chatID int64, prompt string) error {
	resp, err := or.ChatCompletion(ctx, prompt)
	if err != nil {
		return err
	}
	if _, err := repo.SaveResult(ctx, chatID, resp); err != nil {
		return err
	}
	if err := tg.SendMessage(ctx, chatID, resp); err != nil {
		return err
	}
	return nil
}
