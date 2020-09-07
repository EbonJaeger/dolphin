package command

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/diamondburned/arikawa/discord"
	"github.com/diamondburned/arikawa/state"
	"gitlab.com/EbonJaeger/dolphin/config"
)

// ParseConfigCommand checks if the config option specified is valid, and
// updates the config accordingly.
func ParseConfigCommand(state *state.State, cmd DiscordCommand) error {
	// Check if the sender has permission
	if !hasAdminPermisison(state, cmd) {
		log.Warnf("A user tried to update the config without the administrator permissions: '%s'\n", cmd.Sender.Username)
		return SendMissingPermsEmbed(state, cmd.ChannelID, cmd.MessageID)
	}

	// Check the number of args for the command
	if len(cmd.Args) != 2 {
		embed := CreateEmbed(WarnColor, "Incorrect Usage", ":warning: Configuration usage: `!config <option> <value>`", "")
		return SendCommandEmbed(state, cmd, embed)
	}

	// Handle the command and update the config
	confOption := strings.ToLower(cmd.Args[0])
	switch confOption {
	case "allowmentions":
		{
			value, err := strconv.ParseBool(cmd.Args[1])
			if err != nil {
				embed := CreateEmbed(WarnColor, "Invalid Value", ":warning: This option only accepts `true` or `false`.", "")
				return SendCommandEmbed(state, cmd, embed)
			}

			log.Debugf("Updating config property '%s' to '%s'\n", cmd.Args[0], cmd.Args[1])
			conf.Discord.AllowMentions = value
			if err := config.SaveConfig(conf); err != nil {
				return err
			}

			// Let the user know the command was successful
			return sendSuccessEmbed(state, cmd)
		}
	case "channel":
		{
			chanStr := cmd.Args[1]

			// Check if we were given a channel mention or just an ID
			if strings.HasPrefix(chanStr, "<#") && strings.HasSuffix(chanStr, ">") {
				endIndex := len(chanStr) - 1
				chanStr = chanStr[2:endIndex]
			}

			// Get the channel
			snowflake, err := discord.ParseSnowflake(chanStr)
			if err != nil {
				embed := CreateEmbed(WarnColor, "Unknown Channel", fmt.Sprintf(":warning: `%s` does not seem to be an actual channel or channel ID.", chanStr), "")
				return SendCommandEmbed(state, cmd, embed)
			}

			// Update the config
			return updateChannel(state, cmd, snowflake, chanStr)
		}
	case "showadvancements":
		{
			value, err := strconv.ParseBool(cmd.Args[1])
			if err != nil {
				embed := CreateEmbed(WarnColor, "Invalid Value", ":warning: This option only accepts `true` or `false`.", "")
				return SendCommandEmbed(state, cmd, embed)
			}

			log.Debugf("Updating config property '%s' to '%s'\n", cmd.Args[0], cmd.Args[1])
			conf.Discord.MessageOptions.ShowAdvancements = value
			if err := config.SaveConfig(conf); err != nil {
				return err
			}

			// Let the user know the command was successful
			return sendSuccessEmbed(state, cmd)
		}
	case "showdeaths":
		{
			value, err := strconv.ParseBool(cmd.Args[1])
			if err != nil {
				embed := CreateEmbed(WarnColor, "Invalid Value", ":warning: This option only accepts `true` or `false`.", "")
				return SendCommandEmbed(state, cmd, embed)
			}

			log.Debugf("Updating config property '%s' to '%s'\n", cmd.Args[0], cmd.Args[1])
			conf.Discord.MessageOptions.ShowDeaths = value
			if err := config.SaveConfig(conf); err != nil {
				return err
			}

			// Let the user know the command was successful
			return sendSuccessEmbed(state, cmd)
		}
	case "showjoinleave":
		{
			value, err := strconv.ParseBool(cmd.Args[1])
			if err != nil {
				embed := CreateEmbed(WarnColor, "Invalid Value", ":warning: This option only accepts `true` or `false`.", "")
				return SendCommandEmbed(state, cmd, embed)
			}

			log.Debugf("Updating config property '%s' to '%s'\n", cmd.Args[0], cmd.Args[1])
			conf.Discord.MessageOptions.ShowJoinsLeaves = value
			if err := config.SaveConfig(conf); err != nil {
				return err
			}

			// Let the user know the command was successful
			return sendSuccessEmbed(state, cmd)
		}
	case "usemembernicks":
		{
			value, err := strconv.ParseBool(cmd.Args[1])
			if err != nil {
				embed := CreateEmbed(WarnColor, "Invalid Value", ":warning: This option only accepts `true` or `false`.", "")
				return SendCommandEmbed(state, cmd, embed)
			}

			log.Debugf("Updating config property '%s' to '%s'\n", cmd.Args[0], cmd.Args[1])
			conf.Discord.UseMemberNicks = value
			if err := config.SaveConfig(conf); err != nil {
				return err
			}

			// Let the user know the command was successful
			return sendSuccessEmbed(state, cmd)
		}
	default:
		{
			embed := CreateEmbed(WarnColor, "Unknown Option", fmt.Sprintf(":warning: Unknown configuration option `%s`.", confOption), "See `!help config` for configuration options")
			return SendCommandEmbed(state, cmd, embed)
		}
	}
}

func hasAdminPermisison(state *state.State, cmd DiscordCommand) bool {
	user, err := state.Member(cmd.GuildID, cmd.Sender.ID)
	if err != nil {
		return false
	}

	for _, roleID := range user.RoleIDs {
		role, err := state.Role(cmd.GuildID, roleID)
		if err != nil {
			continue
		}

		if role.Permissions.Has(discord.PermissionAdministrator) {
			return true
		}
	}

	return false
}

func sendSuccessEmbed(state *state.State, cmd DiscordCommand) error {
	embed := CreateEmbed(SuccessColor, "Config Updated", ":white_check_mark: Configuration updated successfully!", "")
	return SendCommandEmbed(state, cmd, embed)
}

func updateChannel(state *state.State, cmd DiscordCommand, snowflake discord.Snowflake, newChannel string) error {
	// Get the channel
	channelID := discord.ChannelID(snowflake)
	channel, err := state.Channel(channelID)
	if err != nil {
		embed := CreateEmbed(WarnColor, "Unknown Channel", fmt.Sprintf(":warning: '%s' does not seem to be an actual channel.", newChannel), "")
		return SendCommandEmbed(state, cmd, embed)
	}

	// Get the Guild to check if the Channel is actually in this Guild
	t, _ := state.Message(cmd.ChannelID, cmd.MessageID)
	guild, _ := state.Guild(t.GuildID)
	if channel.GuildID != guild.ID {
		embed := CreateEmbed(WarnColor, "Channel Doesn't Exist", fmt.Sprintf(":warning: No channel with ID `%s` found in this server!", newChannel), "See `!help config` for configuration options")
		return SendCommandEmbed(state, cmd, embed)
	}

	// Update the config
	log.Debugf("Updating config property '%s' to '%s'\n", cmd.Args[0], newChannel)
	conf.Discord.ChannelID = newChannel
	if err := config.SaveConfig(conf); err != nil {
		return err
	}

	// Let the user know the command was successful
	return sendSuccessEmbed(state, cmd)
}
