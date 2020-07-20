package command

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/diamondburned/arikawa/discord"
	"github.com/diamondburned/arikawa/state"
	"gitlab.com/EbonJaeger/dolphin/config"
	"gitlab.com/EbonJaeger/dolphin/rcon"
)

// ListPlayers sends an RCON command to the Minecraft server to list all online players.
func ListPlayers(state *state.State, cmd DiscordCommand, config config.RootConfig) {
	// Create RCON connection
	conn, err := rcon.Dial(config.Minecraft.RconIP, config.Minecraft.RconPort, config.Minecraft.RconPassword)
	if err != nil {
		log.Errorf("Error opening RCON connection: %s\n", err)
		return
	}
	defer conn.Close()

	// Authenticate to RCON
	if err = conn.Authenticate(); err != nil {
		log.Errorf("Error opening RCON connection: %s\n", err)
		return
	}

	// Send the command to Minecraft
	resp, err := conn.SendCommand("minecraft:list")
	if err != nil {
		log.Errorf("Error sending RCON command: %s\n", err)
		return
	}

	embed := createListEmbed(strings.Split(resp, ":"))
	channel, _ := discord.ParseSnowflake(config.Discord.ChannelID)
	message, _ := state.Client.SendEmbed(channel, embed)

	// Remove the embed after 30 seconds
	time.Sleep(30 * time.Second)
	if err := state.Client.DeleteMessage(message.ChannelID, message.ID); err != nil {
		log.Errorf("Error removing embedded message: %s\n", err)
	}
	if err := state.Client.DeleteMessage(message.ChannelID, cmd.MessageID); err != nil {
		log.Errorf("Error removing command message: %s\n", err)
	}
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
