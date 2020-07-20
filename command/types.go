package command

import (
	"github.com/diamondburned/arikawa/discord"
	"github.com/diamondburned/arikawa/state"
	"gitlab.com/EbonJaeger/dolphin/config"
)

// Color code for embed colors.
const (
	ErrorColor = 0xcc0000
	InfoColor  = 0x0099ff
)

// Handler is the interface the each Discord command handler implements.
type Handler struct {
	Name string
	Run  func(state *state.State, cmd DiscordCommand, config config.RootConfig)
}

// Cmd is the type that all command handlers are.
type Cmd struct {
	Name string
}

// DiscordCommand is a command sent by a user in Discord to be parsed and handled.
type DiscordCommand struct {
	Sender    discord.User
	Command   string
	Args      []string
	MessageID discord.Snowflake
}

// Parser is a command parser that handles sending commands to the appropriate handler.
type Parser struct {
	AwaitingResponse map[discord.User]string
	Handlers         []Handler
}
