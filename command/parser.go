package command

import (
	"fmt"
	"strings"
	"time"

	"github.com/DataDrake/waterlog"
	"github.com/diamondburned/arikawa/discord"
	"github.com/diamondburned/arikawa/state"
	"gitlab.com/EbonJaeger/dolphin/config"
)

// log is our package-level Waterlog instance from the main package.
var log *waterlog.WaterLog

// NewParser creates a new command parser with our commands registered.
func NewParser(logger *waterlog.WaterLog) *Parser {
	log = logger

	// Register our commands
	handlers := make([]Handler, 0)

	handlers = append(handlers, Handler{
		Name: "list",
		Run:  ListPlayers,
	})

	return &Parser{
		AwaitingResponse: make(map[discord.User]string),
		Handlers:         handlers,
	}
}

// Parse will turn a Discord message into a DiscordCommand to be
// passed on to a command handler.
func (p *Parser) Parse(message discord.Message, state *state.State, config config.RootConfig) {
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
		message.ID,
	}

	// Send the command to all handlers
	for _, handler := range p.Handlers {
		// Only send to the correct handler
		if handler.Name == cmd.Command {
			log.Debugf("Running Discord bot command: %s\n", handler.Name)
			if err := handler.Run(state, cmd, config); err != nil {
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
				embed := newErrorEmbed(cmd, errorMessage)
				channel, _ := discord.ParseSnowflake(config.Discord.ChannelID)
				message, _ := state.Client.SendEmbed(channel, embed)
				log.Errorf("Error running the '%s' command: %s\n", cmd.Command, err)
				go removeEmbed(state, channel, cmd.MessageID, message.ID)
			}
		}
	}
}

func newErrorEmbed(cmd DiscordCommand, err string) discord.Embed {
	return discord.Embed{
		Color:       ErrorColor,
		Description: fmt.Sprintf(":no_entry: An error occurred while running the `%s` command.", cmd.Command),
		Footer: &discord.EmbedFooter{
			Text: fmt.Sprintf("err: %s", err),
		},
		Type: discord.NormalEmbed,
	}
}

func removeEmbed(state *state.State, channelID, commandID, embedID discord.Snowflake) {
	// Remove the embed after 30 seconds
	time.Sleep(30 * time.Second)
	if err := state.Client.DeleteMessage(channelID, embedID); err != nil {
		log.Errorf("Error removing embed: %s\n", err)
	}
	if err := state.Client.DeleteMessage(channelID, commandID); err != nil {
		log.Errorf("Error removing command message: %s\n", err)
	}
}
