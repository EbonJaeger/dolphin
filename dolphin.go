package dolphin

import (
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/DataDrake/waterlog"
	"github.com/DataDrake/waterlog/format"
	"github.com/DataDrake/waterlog/level"
	"gitlab.com/EbonJaeger/dolphin/command"
	"gitlab.com/EbonJaeger/dolphin/config"
)

// Config is our struct that holds all configuration options.
var Config *config.RootConfig

// Log is our Waterlog instance.
var Log *waterlog.WaterLog

var discordBot *DiscordBot
var parser *command.Parser

// NewDolphin initializes all the things and connects to Discord.
func NewDolphin(cliFlags Flags) {
	// Initialize logging
	Log = waterlog.New(os.Stdout, "", log.Ltime)
	if cliFlags.Debug {
		Log.SetLevel(level.Debug)
	} else {
		Log.SetLevel(level.Info)
	}
	Log.SetFormat(format.Partial)

	// Get default config path if we weren't passed one from the CLI
	configPath := cliFlags.Config

	// Check if we were given a config path
	if configPath == "" {
		// No path given, get the platform-specific config path
		var err error
		configPath, err = config.GetDefaultConfDir()
		if err != nil {
			Log.Fatalf("Unable to get the default config location: %s\n", err)
		}
	}

	// Create the config file if it doesn't exist
	if err := config.CreateConfigFile(configPath); err != nil {
		Log.Fatalf("Error creating config file: %s\n", err)
	}

	// Load our config
	c, err := config.Load()
	if err != nil {
		Log.Fatalf("Error trying to load configuration: %s\n", err)
	}

	// Make sure we have good defaults
	c = config.MergeDefaults(c)
	if err := config.SaveConfig(c); err != nil {
		Log.Fatalf("Error trying to save config: %s\n", err.Error())
	}

	Config = &c

	// Check if a bot token is configured
	if Config.Discord.BotToken == "" {
		Log.Errorln("+-------------------------------------------------------------------------------------+")
		Log.Errorln("| No Discord bot token is configured!                                                 |")
		Log.Errorln("|                                                                                     |")
		Log.Errorln("| Create a Discord bot here:                                                          |")
		Log.Errorln("| https://discordapp.com/developers/applications/me                                   |")
		Log.Errorln("|                                                                                     |")
		Log.Errorln("| Copy the token into your config file, and add the bot to your server with this URL: |")
		Log.Errorln("| https://discordapp.com/oauth2/authorize?client_id=<BOT CLIENT ID>&scope=bot         |")
		Log.Errorln("+-------------------------------------------------------------------------------------+")
		os.Exit(1)
	}

	// Create our Discord client and connect to Discord
	Log.Infoln("Creating Discord session")
	var discordErr error
	discordBot, discordErr = NewDiscordBot()
	if discordErr != nil {
		Log.Fatalf("Error creating Discord bot: %s\n", discordErr.Error())
	}

	// Create our command parser
	parser = command.NewParser(Config, Log)

	Log.Goodln("Connected to Discord! Press CTRL+C to exit")

	// Start watching Minecraft for messages
	go discordBot.WaitForMessages()

	// Wait until told to close
	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt)
	<-sc

	// Newline to keep things pretty
	Log.Println("")

	// Close everything on exit
	if err := discordBot.Close(); err != nil {
		Log.Fatalf("Error while closing: %s\n", err.Error())
	} else {
		Log.Goodln("Dolphin shut down successfully!")
	}
}
