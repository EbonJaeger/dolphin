package dolphin

import (
	"github.com/diamondburned/arikawa/discord"
	"github.com/diamondburned/arikawa/state"
)

// DiscordBot holds our Discord session info and Minecraft log watcher.
type DiscordBot struct {
	avatarURL string
	channel   discord.ChannelID
	guildID   discord.GuildID
	id        discord.UserID
	name      string
	state     *state.State
	watcher   *MinecraftWatcher
}

// Flags holds our command line flags.
type Flags struct {
	Config  string `short:"c" long:"config" description:"Specify the path to the configuration file to use"`
	Debug   bool   `long:"debug" description:"Print additional debug lines to stdout"`
	Version bool   `short:"v" long:"version" description:"Print version information and exit"`
}

// MinecraftMessage represents a message from Minecraft to be sent to Discord.
type MinecraftMessage struct {
	Username string
	Message  string
}
