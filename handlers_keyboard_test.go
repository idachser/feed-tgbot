package main

import (
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

	if len(keyboard.Keyboard) != 2 {
		t.Fatalf("expected 2 keyboard rows, got %d", len(keyboard.Keyboard))
	}

	if len(keyboard.Keyboard[0]) != 2 {
		t.Fatalf("expected first row to have 2 buttons, got %d", len(keyboard.Keyboard[0]))
	}

	if keyboard.Keyboard[0][0].Text != "News" {
		t.Fatalf("expected first button to be News, got %q", keyboard.Keyboard[0][0].Text)
	}

	if keyboard.Keyboard[0][1].Text != "List" {
		t.Fatalf("expected second button to be List, got %q", keyboard.Keyboard[0][1].Text)
	}

	if len(keyboard.Keyboard[1]) != 1 {
		t.Fatalf("expected second row to have 1 button, got %d", len(keyboard.Keyboard[1]))
	}

	if keyboard.Keyboard[1][0].Text != "Help" {
		t.Fatalf("expected third button to be Help, got %q", keyboard.Keyboard[1][0].Text)
	}
}

func TestHelpTextIncludesCoreCommands(t *testing.T) {
	required := []string{"/add", "/list", "/news", "/remove", "/help"}

	for _, command := range required {
		if !strings.Contains(helpText, command) {
			t.Fatalf("helpText must contain %q", command)
		}
	}
}
