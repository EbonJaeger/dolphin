package config

import (
	"bufio"
	"bytes"
	"errors"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/BurntSushi/toml"
	log "github.com/DataDrake/waterlog"
)

// RootConfig is our root config struct.
type RootConfig struct {
	Discord   DiscordConfig
	Minecraft MinecraftConfig
}

// DiscordConfig holds all settings for the Discord side of the application.
type DiscordConfig struct {
	BotToken       string
	ChannelID      string
	AllowMentions  bool
	UseMemberNicks bool
	Webhook        WebhookConfig
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

var configPath string

// CreateConfigFile attempts to create the given config dir+file
// if it doesn't yet exist.
func CreateConfigFile(path string) error {
	var dir string
	var file string

	// See if we're given a specific file to use
	if filepath.Ext(path) != "" {
		dir, file = filepath.Split(path)
	} else {
		dir = filepath.Clean(path)
		file = "dolphin.conf"
	}

	configPath = filepath.Join(dir, file)

	// Check if the path exists
	if _, err := os.Stat(configPath); err != nil {
		if errors.Is(err, os.ErrNotExist) {
			// Attempt to create the config directory
			if mkdirErr := os.Mkdir(dir, 0750); mkdirErr != nil {
				return mkdirErr
			}

			// Attempt to create the config file
			if _, createErr := os.Create(path); createErr != nil {
				return createErr
			}

			// Set the file permissions
			if chmodErr := os.Chmod(path, 0644); chmodErr != nil {
				return chmodErr
			}
		}
	}

	return nil
}

// Load loads the configuration from disk.
func Load() (RootConfig, error) {
	var conf = RootConfig{}
	log.Infof("Loading configuration from '%s'\n", configPath)

	// Parse the file
	if _, err := toml.DecodeFile(configPath, &conf); err != nil {
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
		saveErr = ioutil.WriteFile(configPath, buf.Bytes(), 0600)
	}
	return saveErr
}

// SetDefaults sets sane config defaults and returns the resulting config.
func SetDefaults(config RootConfig) RootConfig {
	if config.Discord == (DiscordConfig{}) {
		config.Discord = DiscordConfig{
			BotToken:       "",
			ChannelID:      "",
			AllowMentions:  true,
			UseMemberNicks: false,
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
