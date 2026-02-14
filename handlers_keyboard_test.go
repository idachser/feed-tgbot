package main

import (
	"fmt"
	"strings"
	"testing"
)

func TestCommandReplyKeyboard(t *testing.T) {
	keyboard := commandReplyKeyboard()

	if !keyboard.ResizeKeyboard {
		t.Fatal("expected ResizeKeyboard to be true")
	}

	if keyboard.OneTimeKeyboard {
		t.Fatal("expected OneTimeKeyboard to be false")
	}

	if len(keyboard.Keyboard) != 3 {
		t.Fatalf("expected 3 keyboard rows, got %d", len(keyboard.Keyboard))
	}

	if len(keyboard.Keyboard[0]) != 2 {
		t.Fatalf("expected first row to have 2 buttons, got %d", len(keyboard.Keyboard[0]))
	}

	if keyboard.Keyboard[0][0].Text != "Add" {
		t.Fatalf("expected first button to be Add, got %q", keyboard.Keyboard[0][0].Text)
	}

	if keyboard.Keyboard[0][1].Text != "Remove" {
		t.Fatalf("expected second button to be Remove, got %q", keyboard.Keyboard[0][1].Text)
	}

	if len(keyboard.Keyboard[1]) != 2 {
		t.Fatalf("expected second row to have 2 buttons, got %d", len(keyboard.Keyboard[1]))
	}

	if keyboard.Keyboard[1][0].Text != "News" {
		t.Fatalf("expected third button to be News, got %q", keyboard.Keyboard[1][0].Text)
	}

	if keyboard.Keyboard[1][1].Text != "List" {
		t.Fatalf("expected fourth button to be List, got %q", keyboard.Keyboard[1][1].Text)
	}

	if len(keyboard.Keyboard[2]) != 2 {
		t.Fatalf("expected third row to have 2 buttons, got %d", len(keyboard.Keyboard[2]))
	}

	if keyboard.Keyboard[2][0].Text != "Help" {
		t.Fatalf("expected fifth button to be Help, got %q", keyboard.Keyboard[2][0].Text)
	}

	if keyboard.Keyboard[2][1].Text != "Updates" {
		t.Fatalf("expected sixth button to be Updates, got %q", keyboard.Keyboard[2][1].Text)
	}
}

func TestHelpTextIncludesCoreCommands(t *testing.T) {
	required := []string{"/add", "/list", "/news", "/remove", "/updates", "/help"}

	for _, command := range required {
		if !strings.Contains(helpText, command) {
			t.Fatalf("helpText must contain %q", command)
		}
	}
}

func TestNewsSourceSelectionKeyboard(t *testing.T) {
	userID := int64(42)
	feeds := []string{
		"https://example.com/feed1.xml",
		"https://example.com/feed2.xml",
	}

	keyboard := newsSourceSelectionKeyboard(userID, feeds)

	if len(keyboard.InlineKeyboard) != len(feeds) {
		t.Fatalf("expected %d rows, got %d", len(feeds), len(keyboard.InlineKeyboard))
	}

	for i, row := range keyboard.InlineKeyboard {
		if len(row) != 1 {
			t.Fatalf("expected 1 button in row %d, got %d", i, len(row))
		}

		btn := row[0]
		expectedCallback := fmt.Sprintf("%s:%d:%d", callbackNewsFeed, userID, i)
		if btn.CallbackData != expectedCallback {
			t.Fatalf("unexpected callback data at row %d: got %q want %q", i, btn.CallbackData, expectedCallback)
		}
		if btn.Text == "" {
			t.Fatalf("button text at row %d should not be empty", i)
		}
	}
}
