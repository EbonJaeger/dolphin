package dolphin

import (
	"testing"
)

var watcher = NewWatcher()

func TestParseChatLine(t *testing.T) {
	// Given
	input := "[12:32:45] [Server thread/INFO]: <TestUser> Sending a chat message"
	expected := &MinecraftMessage{
		Username: "TestUser",
		Message:  "Sending a chat message",
	}
	// When
	actual := watcher.ParseLine(input)
	// Then
	if actual.Username != expected.Username {
		t.Errorf("Parsing chat line got incorrect username, got: %s, expected: %s", actual.Username, expected.Username)
	}
	if actual.Message != expected.Message {
		t.Errorf("Parsing chat line got incorrect message, got: %s, expected: %s", actual.Message, expected.Message)
	}
}

func TestParseJoinLine(t *testing.T) {
	// Given
	input := "[12:32:45] [Server thread/INFO]: TestUser joined the game"
	Config = SetDefaults(RootConfig{})
	expected := &MinecraftMessage{
		Username: "Dolphin",
		Message:  "TestUser joined the game",
	}
	// When
	actual := watcher.ParseLine(input)
	// Then
	if actual.Username != expected.Username {
		t.Errorf("Parsing chat line got incorrect username, got: %s, expected: %s", actual.Username, expected.Username)
	}
	if actual.Message != expected.Message {
		t.Errorf("Parsing chat line got incorrect message, got: %s, expected: %s", actual.Message, expected.Message)
	}
}

func TestParseLeaveLine(t *testing.T) {
	// Given
	input := "[12:32:45] [Server thread/INFO]: TestUser left the game"
	Config = SetDefaults(RootConfig{})
	expected := &MinecraftMessage{
		Username: "Dolphin",
		Message:  "TestUser left the game",
	}
	// When
	actual := watcher.ParseLine(input)
	// Then
	if actual.Username != expected.Username {
		t.Errorf("Parsing chat line got incorrect username, got: %s, expected: %s", actual.Username, expected.Username)
	}
	if actual.Message != expected.Message {
		t.Errorf("Parsing chat line got incorrect message, got: %s, expected: %s", actual.Message, expected.Message)
	}
}

func TestParseAdvancement1Line(t *testing.T) {
	// Given
	input := "[12:32:45] [Server thread/INFO]: TestUser has made the advancement [MonsterHunter]"
	Config = SetDefaults(RootConfig{})
	expected := &MinecraftMessage{
		Username: "Dolphin",
		Message:  ":partying_face: TestUser has made the advancement [MonsterHunter]",
	}
	// When
	actual := watcher.ParseLine(input)
	// Then
	if actual.Username != expected.Username {
		t.Errorf("Parsing chat line got incorrect username, got: %s, expected: %s", actual.Username, expected.Username)
	}
	if actual.Message != expected.Message {
		t.Errorf("Parsing chat line got incorrect message, got: %s, expected: %s", actual.Message, expected.Message)
	}
}

func TestParseAdvancement2Line(t *testing.T) {
	// Given
	input := "[12:32:45] [Server thread/INFO]: TestUser has completed the challenge [MonsterHunter]"
	Config = SetDefaults(RootConfig{})
	expected := &MinecraftMessage{
		Username: "Dolphin",
		Message:  ":partying_face: TestUser has completed the challenge [MonsterHunter]",
	}
	// When
	actual := watcher.ParseLine(input)
	// Then
	if actual.Username != expected.Username {
		t.Errorf("Parsing chat line got incorrect username, got: %s, expected: %s", actual.Username, expected.Username)
	}
	if actual.Message != expected.Message {
		t.Errorf("Parsing chat line got incorrect message, got: %s, expected: %s", actual.Message, expected.Message)
	}
}

func TestParseServerStartLine(t *testing.T) {
	// Given
	input := "[12:32:45] [Server thread/INFO]: Done (21.3242s)! For help, type \"help\""
	Config = SetDefaults(RootConfig{})
	expected := &MinecraftMessage{
		Username: "Dolphin",
		Message:  ":white_check_mark: Server has started",
	}
	// When
	actual := watcher.ParseLine(input)
	// Then
	if actual.Username != expected.Username {
		t.Errorf("Parsing chat line got incorrect username, got: %s, expected: %s", actual.Username, expected.Username)
	}
	if actual.Message != expected.Message {
		t.Errorf("Parsing chat line got incorrect message, got: %s, expected: %s", actual.Message, expected.Message)
	}
}

func TestParseServerStopLine(t *testing.T) {
	// Given
	input := "[12:32:45] [Server thread/INFO]: Stopping the server"
	Config = SetDefaults(RootConfig{})
	expected := &MinecraftMessage{
		Username: "Dolphin",
		Message:  ":x: Server is shutting down",
	}
	// When
	actual := watcher.ParseLine(input)
	// Then
	if actual.Username != expected.Username {
		t.Errorf("Parsing chat line got incorrect username, got: %s, expected: %s", actual.Username, expected.Username)
	}
	if actual.Message != expected.Message {
		t.Errorf("Parsing chat line got incorrect message, got: %s, expected: %s", actual.Message, expected.Message)
	}
}
