package dolphin

import (
	"testing"
)

var watcher = NewWatcher("TestBot")

func TestParseChatLine(t *testing.T) {
	// Given
	input := "[12:32:45] [Server thread/INFO]: <TestUser> Sending a chat message"
	expected := &MinecraftMessage{
		Username: "TestUser",
		Message:  "Sending a chat message",
	}
	// When
	actual := watcher.ParseLine("TestBot", input)
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
	expected := &MinecraftMessage{
		Username: "TestBot",
		Message:  "TestUser joined the game",
	}
	// When
	actual := watcher.ParseLine("TestBot", input)
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
	expected := &MinecraftMessage{
		Username: "TestBot",
		Message:  "TestUser left the game",
	}
	// When
	actual := watcher.ParseLine("TestBot", input)
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
	expected := &MinecraftMessage{
		Username: "TestBot",
		Message:  ":partying_face: TestUser has made the advancement [MonsterHunter]",
	}
	// When
	actual := watcher.ParseLine("TestBot", input)
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
	expected := &MinecraftMessage{
		Username: "TestBot",
		Message:  ":partying_face: TestUser has completed the challenge [MonsterHunter]",
	}
	// When
	actual := watcher.ParseLine("TestBot", input)
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
	expected := &MinecraftMessage{
		Username: "TestBot",
		Message:  ":white_check_mark: Server has started",
	}
	// When
	actual := watcher.ParseLine("TestBot", input)
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
	expected := &MinecraftMessage{
		Username: "TestBot",
		Message:  ":x: Server is shutting down",
	}
	// When
	actual := watcher.ParseLine("TestBot", input)
	// Then
	if actual.Username != expected.Username {
		t.Errorf("Parsing chat line got incorrect username, got: %s, expected: %s", actual.Username, expected.Username)
	}
	if actual.Message != expected.Message {
		t.Errorf("Parsing chat line got incorrect message, got: %s, expected: %s", actual.Message, expected.Message)
	}
}
