package dolphin

import (
	"log"
	"os"
	"os/signal"
	"os/user"
	"path/filepath"
	"syscall"

	"github.com/DataDrake/waterlog"
	"github.com/DataDrake/waterlog/format"
	"github.com/DataDrake/waterlog/level"
	"github.com/bwmarrin/discordgo"
)

// Config is our struct that holds all configuration options.
var Config RootConfig

// Log is our Waterlog instance.
var Log *waterlog.WaterLog

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
	if configPath == "" {
		user, err := user.Current()
		if err != nil {
			Log.Fatalf("Error while getting the current user: %s\n", err.Error())
		}
		confDir := filepath.Join(user.HomeDir, ".config", "dolphin")
		// Check if the directory exists
		if _, dirErr := os.Stat(confDir); dirErr != nil {
			if os.IsNotExist(dirErr) {
				// Attempt to create the directory
				if createErr := os.Mkdir(confDir, 0755); err != nil {
					Log.Fatalf("Error creating default config directory: %s\n", createErr.Error())
				}
			} else {
				Log.Fatalf("Error getting config directory: %s\n", dirErr.Error())
			}
		}
		configPath = confDir
	}
	// Load our config
	var readErr error
	if Config, readErr = LoadConfig(configPath); readErr != nil {
		Log.Fatalf("Error trying to load configuration: %s\n", readErr.Error())
	}
	// Make sure we have good defaults
	if Config == (RootConfig{}) {
		Config = SetDefaults(Config)
		if err := SaveConfig(Config); err != nil {
			Log.Fatalf("Error trying to save config: %s\n", err.Error())
		}
	}
	// Create our Discord client
	Log.Infoln("Creating Discord session")
	s, err := discordgo.New("Bot" + Config.Discord.BotToken)
	if err != nil {
		Log.Fatalf("Unable to initialize Discord session: %s\n", err.Error())
	}
	// Add our Discord handlers
	s.AddHandler(OnReady)
	s.AddHandler(OnGuildCreate)
	s.AddHandler(OnMessageCreate)
	// Open Discord websocet
	Log.Infoln("Connecting to Discord websocket")
	if err := s.Open(); err != nil {
		Log.Fatalf("Unable to connect to Discord websocket: %s\n", err.Error())
	}
	// Wait until told to close
	Log.Goodln("Connected to Discord! Press CTRL+C to exit")
	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt, os.Kill)
	<-sc
	// Newline to keep things pretty
	Log.Println("")
	// Close the Discord session on exit
	if err = s.Close(); err != nil {
		Log.Fatalf("Error while closing Discord connection: %s\n", err.Error())
	}
	Log.Goodln("Dolphin shut down successfully!")
}
