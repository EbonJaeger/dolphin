package command

import (
	"fmt"
	"strings"

	"github.com/diamondburned/arikawa/discord"
	"github.com/diamondburned/arikawa/state"
)

// ShowHelp creates and sends an embed listing all of the commands that
// can be used on Discord.
func ShowHelp(state *state.State, cmd DiscordCommand) error {
	// Create our help embed
	embed := discord.Embed{
		Title: "Bot Command Help",
		Type:  discord.NormalEmbed,
		Color: InfoColor,
	}

	// Create our description text with all of the commands
	b := strings.Builder{}
	b.WriteString("Here is a list of all available commands:\n")
	b.WriteString("\n")
	for _, handler := range handlers {
		b.WriteString(fmt.Sprintf("`!%s`  **-**  %s\n", handler.Name, handler.Desc))
	}
	embed.Description = b.String()

	// Create a DM channel with the sender
	dm, err := state.Client.CreatePrivateChannel(cmd.Sender.ID)
	if err != nil {
		return err
	}

	// Attempt to send the help embed to the DM
	if _, err := state.Client.SendEmbed(dm.ID, embed); err != nil {
		// An error happened; Probably the sender doesn't allow DM's from randos.
		// So, send it to the channel instead, and remove after 30 seconds.
		channel, _ := discord.ParseSnowflake(conf.Discord.ChannelID)
		message, err := state.Client.SendEmbed(channel, embed)
		if err != nil {
			return err
		}

		removeEmbed(state, channel, cmd.MessageID, message.ID)
	}

	return nil
}
