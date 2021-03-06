package command

import (
	"fmt"
	"strings"

	"github.com/DataDrake/waterlog"
	"github.com/diamondburned/arikawa/discord"
	"github.com/diamondburned/arikawa/state"
	"gitlab.com/EbonJaeger/dolphin/config"
)

var (
	conf     *config.RootConfig
	handlers []Handler
	log      *waterlog.WaterLog
)

// NewParser creates a new command parser with our commands registered.
func NewParser(configuration *config.RootConfig, logger *waterlog.WaterLog) *Parser {
	conf = configuration
	log = logger

	// Register our commands
	handlers = append(handlers, Handler{
		Name: "config",
		Desc: "Configure certain bot settings",
		Run:  ParseConfigCommand,
	})

	handlers = append(handlers, Handler{
		Name: "help",
		Desc: "Show all available bot commands",
		Run:  ShowHelp,
	})

	handlers = append(handlers, Handler{
		Name: "list",
		Desc: "List all online players",
		Run:  ListPlayers,
	})

	return &Parser{}
}

// Parse will turn a Discord message into a DiscordCommand to be
// passed on to a command handler.
func (p *Parser) Parse(message discord.Message, state *state.State, resp chan bool) {
	// Forget about the command prefix
	raw := message.Content[1:]
	parts := strings.Split(raw, " ")

	log.Debugf("Parsing command from Discord: %s\n", parts)

	// Split out any command args
	var args []string
	if len(parts) > 1 {
		args = parts[1:]
	} else {
		args = make([]string, 0)
	}

	cmd := DiscordCommand{
		message.Author,
		parts[0],
		args,
		message.GuildID,
		message.ChannelID,
		message.ID,
	}

	// Send the command to all handlers
	for _, handler := range handlers {
		// Only send to the correct handler
		if handler.Name == cmd.Command {
			log.Debugf("Running Discord bot command: %s\n", handler.Name)
			resp <- true

			if err := handler.Run(state, cmd); err != nil {
				handleCommandError(state, cmd, err)
			}
		}
	}
}

func handleCommandError(state *state.State, cmd DiscordCommand, err error) {
	// Sanitize error from RCON
	errorMessage := err.Error()
	if strings.HasPrefix(errorMessage, "dial tcp") {
		// This is so hacky and I hate it.
		start := strings.Index(errorMessage, ":")
		errorMessage = errorMessage[start+1:]
		start = strings.Index(errorMessage, ":")
		errorMessage = errorMessage[start+1:]
	}

	// Embed an error and log it
	embed := CreateEmbed(ErrorColor, "Error", fmt.Sprintf(":no_entry: An error occurred while running the `%s` command.", cmd.Command), fmt.Sprintf("err: %s", errorMessage))
	snowflake, _ := discord.ParseSnowflake(conf.Discord.ChannelID)
	channel := discord.ChannelID(snowflake)
	message, sendError := state.Client.SendEmbed(channel, embed)
	if sendError != nil {
		log.Errorf("Error while trying to display another error: %s\n", sendError)
		log.Errorf("The previous error was: %s\n", err)
		return
	}

	log.Errorf("Error running the '%s' command: %s\n", cmd.Command, err)
	if err := RemoveEmbed(state, channel, cmd.MessageID, message.ID); err != nil {
		log.Errorf("Error trying to remove an error embed: %s\n", err)
	}
}
