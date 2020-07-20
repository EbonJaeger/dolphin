package command

import (
	"strings"

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
			go handler.Run(state, cmd, config)
		}
	}
}
