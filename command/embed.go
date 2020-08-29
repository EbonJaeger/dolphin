package command

import (
	"time"

	"github.com/diamondburned/arikawa/discord"
	"github.com/diamondburned/arikawa/state"
)

// CreateEmbed makes a new Discord embed with the given parameters.
func CreateEmbed(color discord.Color, title string, bodyText string, footerText string) discord.Embed {
	embed := discord.Embed{
		Color:       color,
		Description: bodyText,
		Title:       title,
		Type:        discord.NormalEmbed,
	}

	// Add a footer if we were given footer text
	if footerText != "" {
		embed.Footer = &discord.EmbedFooter{
			Text: footerText,
		}
	}

	return embed
}

// RemoveEmbed attempts to remove an embed after 30 seconds. If a message
// is unable to be deleted, an error is returned.
func RemoveEmbed(state *state.State, channelID discord.ChannelID, commandID, embedID discord.MessageID) error {
	// Remove the embed after 30 seconds
	time.Sleep(30 * time.Second)
	if err := state.Client.DeleteMessage(channelID, embedID); err != nil {
		return err
	}
	if err := state.Client.DeleteMessage(channelID, commandID); err != nil {
		return err
	}

	return nil
}

// SendCommandEmbed sends the given embed to the specified Discord channel,
// and removes it after 30 seconds. If there was an error sending or removing
// the embed, an error is returned.
func SendCommandEmbed(state *state.State, cmd DiscordCommand, embed discord.Embed) error {
	message, err := state.Client.SendEmbed(cmd.ChannelID, embed)
	if err != nil {
		return err
	}

	return RemoveEmbed(state, cmd.ChannelID, message.ID, cmd.MessageID)
}

// SendMissingPermsEmbed creates a new embed for a player missing a permission.
// This embed is then sent to the channel, and the command and embed are removed
// after 30 seconds.
func SendMissingPermsEmbed(state *state.State, channelID discord.ChannelID, messageID discord.MessageID) error {
	embed := CreateEmbed(ErrorColor, "Insufficient Permissions", ":no_entry: You don't have permission to run that command!", "Please contact a server administrator for help.")
	message, err := state.Client.SendEmbed(channelID, embed)
	if err != nil {
		return err
	}

	return RemoveEmbed(state, channelID, message.ID, messageID)
}
