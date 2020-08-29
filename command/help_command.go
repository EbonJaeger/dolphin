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
	if len(cmd.Args) == 1 {
		topic := strings.ToLower(cmd.Args[0])

		// Show a more detailed help page if one if available
		switch topic {
		case "config":
			return showConfigHelp(state, cmd)
		default:
			{
				embed := CreateEmbed(WarnColor, "Unknown Help Topic", fmt.Sprintf("There is no help page for `%s`. Use `!help` for commands.", topic), "")
				return SendCommandEmbed(state, cmd, embed)
			}
		}
	} else {
		// Show the default help page
		return showDefaultHelp(state, cmd)
	}

}

func showConfigHelp(state *state.State, cmd DiscordCommand) error {
	// Create our help embed
	embed := discord.Embed{
		Title: "Config Command Help",
		Type:  discord.NormalEmbed,
		Color: InfoColor,
	}

	// Create our description text with all of the commands
	b := strings.Builder{}
	b.WriteString("Here is a list of all configuration options that can be set via the `!config` command.\n")
	b.WriteString("The configuration can be updated by entering `!config <option> <value>`.\n")
	b.WriteString("Bools are `true` or `false`, while `strings` are normal text.\n")
	b.WriteString("\n")
	b.WriteString("**Name — Value Type**\n")

	b.WriteString("`allowmentions` ** — ** bool\n")
	b.WriteString("`channelid` ** — ** string\n")
	b.WriteString("`showadvancements` ** — ** bool\n")
	b.WriteString("`showdeaths` ** — ** bool\n")
	b.WriteString("`showjoinleave` ** — ** bool\n")
	b.WriteString("`usemembernicks` ** — ** bool\n")

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
		return SendCommandEmbed(state, cmd, embed)
	}

	return nil
}

func showDefaultHelp(state *state.State, cmd DiscordCommand) error {
	// Create our help embed
	embed := discord.Embed{
		Title: "Bot Command Help",
		Type:  discord.NormalEmbed,
		Color: InfoColor,
		Footer: &discord.EmbedFooter{
			Text: "Use !help config to view possible options",
		},
	}

	// Create our description text with all of the commands
	b := strings.Builder{}
	b.WriteString("Here is a list of all available commands:\n")
	b.WriteString("\n")
	for _, handler := range handlers {
		b.WriteString(fmt.Sprintf("`!%s`  **—**  %s\n", handler.Name, handler.Desc))
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
		return SendCommandEmbed(state, cmd, embed)
	}

	return nil
}
