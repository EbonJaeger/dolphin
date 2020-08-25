package config

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestCreateConfigFileGivenPath(t *testing.T) {
	// given
	dir := os.TempDir()
	fullPath := filepath.Join(dir, "dolphin.conf")

	// when/then
	err := CreateConfigFile(dir)
	if err != nil {
		t.Errorf("Failed to create config file: %s", err)
	}
	defer os.Remove(fullPath)

	t.Logf("Created file at '%s'\n", fullPath)

	// Try to stat the file
	if _, err := os.Stat(fullPath); err != nil {
		t.Errorf("Failed to create config file: %s", err)
	}
}

func TestCreateConfigFileGivenAll(t *testing.T) {
	// given
	dir := filepath.Join(os.TempDir(), "dolphin_testing")
	fullPath := filepath.Join(dir, "dolphin_given.conf")

	// when/then
	err := CreateConfigFile(fullPath)
	if err != nil {
		t.Errorf("Failed to create config file: %s", err)
	}
	defer os.RemoveAll(dir)

	t.Logf("Created file at '%s'\n", fullPath)

	// Try to stat the file
	if _, err := os.Stat(fullPath); err != nil {
		t.Errorf("Failed to create config file: %s", err)
	}
}

func TestSaveAndLoadConfig(t *testing.T) {
	// given
	data := RootConfig{
		DiscordConfig{
			BotToken:       "bot-token",
			ChannelID:      "2576235623",
			AllowMentions:  true,
			UseMemberNicks: false,
			Webhook: WebhookConfig{
				Enabled: false,
				URL:     "",
			},
		},

		MinecraftConfig{
			RconIP:              "localhost",
			RconPort:            25575,
			RconPassword:        "igb348grt348fg",
			TellrawTemplate:     `[{"color": "white", "text": "<%username%> %message%"}]`,
			CustomDeathKeywords: &[]string{},
			UseLogFile:          true,
			LogFilePath:         "/home/minecraft/server/logs/latest.log",
		},
	}

	// Create a temp config file
	dir := filepath.Join(os.TempDir(), "dolphin-testing")
	path := filepath.Join(dir, "dolphin_save_load_test.conf")
	if err := CreateConfigFile(path); err != nil {
		t.Errorf("Error creating testing directory: %s\n", err)
	}
	defer os.RemoveAll(dir)

	// when
	err := SaveConfig(data)
	if err != nil {
		t.Errorf("Failed to save config: %s\n", err)
	}

	// then
	loaded, err := Load()
	if err != nil {
		t.Errorf("Failed to load config: %s\n", err)
	}

	// Check if the loaded config matches the original
	if !cmp.Equal(loaded, data) {
		diff := cmp.Diff(loaded, data)
		t.Errorf("Loaded config is different from the given config: Diff: %s\n", diff)
	}
}

func TestSetDefaults(t *testing.T) {
	// given
	expected := RootConfig{
		DiscordConfig{
			BotToken:       "",
			ChannelID:      "",
			AllowMentions:  true,
			UseMemberNicks: false,
			Webhook: WebhookConfig{
				Enabled: false,
				URL:     "",
			},
		},

		MinecraftConfig{
			RconIP:              "localhost",
			RconPort:            25575,
			RconPassword:        "",
			TellrawTemplate:     `[{"color": "white", "text": "<%username%> %message%"}]`,
			CustomDeathKeywords: &[]string{},
			UseLogFile:          true,
			LogFilePath:         "/home/minecraft/server/logs/latest.log",
		},
	}

	// when
	actual := SetDefaults()

	// then
	if !cmp.Equal(actual, expected) {
		diff := cmp.Diff(actual, expected)
		t.Errorf("Setting config defaults is incorrect: Diff: %s", diff)
	}
}
