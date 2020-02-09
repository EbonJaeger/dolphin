package dolphin

import (
	"bufio"
	"bytes"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/BurntSushi/toml"
)

// RootConfig is our root config struct.
type RootConfig struct {
	Discord   DiscordConfig
	Minecraft MinecraftConfig
}

// DiscordConfig holds all settings for the Discord side of the application.
type DiscordConfig struct {
	BotToken        string
	BotName         string
	ChannelID       string
	MessageTemplate string
	AllowMentions   bool
	UseWebhooks     bool
	WebhookID       string
}

// MinecraftConfig holds all settings for the Minecraft server side of the application.
type MinecraftConfig struct {
	RconIP          string
	RconPort        int
	RconPassword    string
	TellrawTemplate string
	DeathKeywords   *[]string
	UseLogFile      bool
	LogFilePath     string
}

// LoadConfig loads the configuration from disk.
func LoadConfig() (RootConfig, error) {
	var conf = RootConfig{}
	// Get our config file
	path := filepath.Join("", "dolphin.conf")
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
	path := filepath.Join("", "dolphin.conf")
	// Create our buffer and encoder
	writer := bufio.NewWriter(&buf)
	encoder := toml.NewEncoder(writer)
	// Encode the struct as TOML
	if saveErr = encoder.Encode(data); saveErr != nil {
		// Write to the file
		saveErr = ioutil.WriteFile(path, buf.Bytes(), 0644)
	}
	return saveErr
}

// SetDefaults sets sane config defaults and returns the resulting config.
func SetDefaults(config RootConfig) RootConfig {
	if config.Discord == (DiscordConfig{}) {
		config.Discord = DiscordConfig{
			BotToken:        "",
			BotName:         "Dolphin",
			ChannelID:       "",
			MessageTemplate: "`%username%`: %message%",
			AllowMentions:   true,
			UseWebhooks:     false,
			WebhookID:       "",
		}
	}
	if config.Minecraft == (MinecraftConfig{}) {
		config.Minecraft = MinecraftConfig{
			RconIP:          "localhost",
			RconPort:        25575,
			RconPassword:    "",
			TellrawTemplate: `[{"color": "white", "text": "<%username%> %message%"}]`,
			DeathKeywords:   &[]string{"shot", "fell", "death", "died", "doomed", "pummeled", "removed", "didn't want", "withered", "squashed", "flames", "burnt", "walked into", "bang", "roasted", "squished", "drowned", "killed", "slain", "blown", "blew", "suffocated", "struck", "lava", "impaled", "speared", "fireballed", "finished", "kinetic"},
			UseLogFile:      true,
			LogFilePath:     "/home/minecraft/server/logs/latest.log",
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
