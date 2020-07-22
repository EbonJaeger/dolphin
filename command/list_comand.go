package command

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/diamondburned/arikawa/discord"
	"github.com/diamondburned/arikawa/state"
	"gitlab.com/EbonJaeger/dolphin/rcon"
)

// ListPlayers sends an RCON command to the Minecraft server to list all online players.
func ListPlayers(state *state.State, cmd DiscordCommand) error {
	// Create RCON connection
	conn, err := rcon.Dial(conf.Minecraft.RconIP, conf.Minecraft.RconPort, conf.Minecraft.RconPassword)
	if err != nil {
		return err
	}
	defer conn.Close()

	// Authenticate to RCON
	if err = conn.Authenticate(); err != nil {
		return err
	}

	// Send the command to Minecraft
	resp, err := conn.SendCommand("minecraft:list")
	if err != nil {
		return err
	}

	// Vanilla servers dont support the 'minecraft:' command prefix
	if strings.HasPrefix(resp, "Unknown or incomplete command") {
		resp, err = conn.SendCommand("list")
		if err != nil {
			return err
		}
	}

	embed := createListEmbed(strings.Split(resp, ":"))
	channel, _ := discord.ParseSnowflake(conf.Discord.ChannelID)
	message, err := state.Client.SendEmbed(channel, embed)
	if err != nil {
		return err
	}

	// Remove the embed after 30 seconds
	removeEmbed(state, channel, cmd.MessageID, message.ID)

	return nil
}

func createListEmbed(resp []string) discord.Embed {
	online, max := getPlayerCount(resp[0])

	embed := discord.Embed{
		Color:       InfoColor,
		Description: fmt.Sprintf("There are **%d** out of **%d** players online.", online, max),
		Title:       "Online Players",
		Type:        discord.NormalEmbed,
	}

	if len(resp) > 1 {
		embed.Footer = &discord.EmbedFooter{
			Text: strings.TrimSpace(resp[1]),
		}
	}

	return embed
}

func getPlayerCount(text string) (online, max int) {
	parts := strings.Split(text, " ")
	gotOnline := false
	for _, part := range parts {
		if num, err := strconv.Atoi(part); err == nil {
			if gotOnline {
				max = num
			} else {
				online = num
				gotOnline = true
			}
		}
	}

	return
}
