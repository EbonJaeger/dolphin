package dolphin

import (
	"bufio"
	"bytes"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/BurntSushi/toml"
)

// RootConfig is our root config struct.
type RootConfig struct {
	Discord   DiscordConfig
	Minecraft MinecraftConfig
}

// DiscordConfig holds all settings for the Discord side of the application.
type DiscordConfig struct {
	BotToken      string
	ChannelID     string
	AllowMentions bool
	UseNick       bool
	Webhook       WebhookConfig
}

// WebhookConfig holds settings for using Discord webhooks to send messages.
type WebhookConfig struct {
	Enabled bool
	URL     string
}

// MinecraftConfig holds all settings for the Minecraft server side of the application.
type MinecraftConfig struct {
	RconIP              string
	RconPort            int
	RconPassword        string
	TellrawTemplate     string
	CustomDeathKeywords *[]string
	UseLogFile          bool
	LogFilePath         string
}

var filePath string

// LoadConfig loads the configuration from disk.
func LoadConfig(path string) (RootConfig, error) {
	var conf = RootConfig{}
	// Make sure our path ends in the name of the file
	if !strings.HasSuffix(path, ".conf") {
		path = filepath.Join(path, "dolphin.conf")
	}
	Log.Infof("Using configuration at '%s'\n", path)
	// Create the file if it doesn't exist
	if err := createFile(path); err != nil {
		return conf, err
	}
	// Parse the file
	if _, err := toml.DecodeFile(path, &conf); err != nil {
		return conf, err
	}
	return conf, nil
}

// SaveConfig saves the current configuration to disk.
func SaveConfig(data interface{}) error {
	var (
		buf     bytes.Buffer
		saveErr error
	)
	// Create our buffer and encoder
	writer := bufio.NewWriter(&buf)
	encoder := toml.NewEncoder(writer)
	// Encode the struct as TOML
	if saveErr = encoder.Encode(data); saveErr == nil {
		// Write to the file
		saveErr = ioutil.WriteFile(filePath, buf.Bytes(), 0644)
	}
	return saveErr
}

// SetDefaults sets sane config defaults and returns the resulting config.
func SetDefaults(config RootConfig) RootConfig {
	if config.Discord == (DiscordConfig{}) {
		config.Discord = DiscordConfig{
			BotToken:      "",
			ChannelID:     "",
			AllowMentions: true,
			UseNick:       false,
			Webhook: WebhookConfig{
				Enabled: false,
				URL:     "",
			},
		}
	}
	if config.Minecraft == (MinecraftConfig{}) {
		config.Minecraft = MinecraftConfig{
			RconIP:              "localhost",
			RconPort:            25575,
			RconPassword:        "",
			TellrawTemplate:     `[{"color": "white", "text": "<%username%> %message%"}]`,
			CustomDeathKeywords: &[]string{},
			UseLogFile:          true,
			LogFilePath:         "/home/minecraft/server/logs/latest.log",
		}
	}
	return config
}

// createFile creates a blank file if it does not exist.
func createFile(path string) error {
	// Check if the file exists
	if _, err := os.Stat(path); err != nil {
		if os.IsNotExist(err) {
			// Create the file
			file, createErr := os.Create(path)
			if createErr != nil {
				return createErr
			}
			// Set the file permissions
			if chmodErr := file.Chmod(0644); chmodErr != nil {
				return chmodErr
			}
		} else {
			// Other error
			return err
		}
	}
	return nil
}
