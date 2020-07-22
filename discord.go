package dolphin

import (
	"errors"
	"fmt"
	"regexp"
	"strings"

	"github.com/diamondburned/arikawa/api"
	"github.com/diamondburned/arikawa/discord"
	"github.com/diamondburned/arikawa/gateway"
	"github.com/diamondburned/arikawa/state"
	"gitlab.com/EbonJaeger/dolphin/rcon"
)

var webhookRegex = regexp.MustCompile("https://discordapp.com/api/webhooks/(.*)/(.*)")

// DiscordBot holds our Discord session info and Minecraft log watcher.
type DiscordBot struct {
	avatarURL string
	channel   discord.Snowflake
	guildID   discord.Snowflake
	id        discord.Snowflake
	name      string
	state     *state.State
	watcher   *MinecraftWatcher
}

// NewDiscordBot creates a new DiscordBot with a MinecraftWatcher and
// connects to discord.
func NewDiscordBot() (*DiscordBot, error) {
	bot := &DiscordBot{}
	var discordErr error

	// Create Discord session
	s, discordErr := state.New("Bot " + Config.Discord.BotToken)
	if discordErr != nil {
		return nil, discordErr
	}
	bot.state = s

	// Add our Discord handlers
	bot.state.AddHandler(bot.onReady)
	bot.state.AddHandler(bot.onGuildCreate)
	bot.state.AddHandler(bot.onMessageCreate)

	// Connect to Discord websocket
	if discordErr = bot.state.Open(); discordErr != nil {
		return nil, discordErr
	}

	// Get our Discord user
	self, discordErr := bot.state.Me()
	if discordErr != nil {
		return nil, discordErr
	}

	// Set our data and create the Minecraft watcher
	bot.id = self.ID
	bot.name = self.Username
	bot.avatarURL = self.AvatarURL()

	if Config.Discord.ChannelID != "" {
		bot.channel, discordErr = discord.ParseSnowflake(Config.Discord.ChannelID)
		if discordErr != nil {
			return nil, discordErr
		}
	} else {
		return nil, errors.New("no channel ID configured")
	}

	bot.watcher = NewWatcher(self.Username)

	return bot, discordErr
}

// Close cleans up the watcher and closes the Discord session.
func (bot *DiscordBot) Close() error {
	var closeErr error

	if err := bot.watcher.Close(); err != nil {
		closeErr = err
	}
	if err := bot.state.Session.Close(); err != nil {
		closeErr = err
	}

	return closeErr
}

// WaitForMessages starts the Minecraft log watcher and waits for messages
// on a messages channel.
func (bot *DiscordBot) WaitForMessages() {
	// Make our messages channel
	mc := make(chan *MinecraftMessage)

	// Start our Minecraft watcher
	go bot.watcher.Watch(mc)
	for {
		// Read message from the channel
		msg := <-mc
		Log.Debugf("Received a line from Minecraft: Username='%s', Text='%s'\n", msg.Username, msg.Message)
		// Send the message to the Discord channel
		bot.sendToDiscord(msg)
	}
}

// onReady sets the bot's Discord status.
func (bot *DiscordBot) onReady(e *gateway.ReadyEvent) {
	// Set the bot gaming status
	err := bot.state.Gateway.UpdateStatus(gateway.UpdateStatusData{
		Game: &discord.Activity{
			Name: "Bridging the Minecraft/Discord gap",
		},
	})

	if err != nil {
		Log.Errorf("Unable to update Discord status: %s\n", err)
	}
}

// onGuildCreate handles when the bot joins or connects to a Guild.
func (bot *DiscordBot) onGuildCreate(e *gateway.GuildCreateEvent) {
	// Make sure the guild is available
	if e.Unavailable {
		Log.Warnf("Attempted to join Guild '%s', but it was unavailable\n", e.Guild.Name)
		return
	}

	if bot.guildID.String() != "" {
		Log.Warnf("Received a Guild join event for '%s', but we've already joined one\n", e.Guild.Name)
		return
	}

	Log.Infof("Connected to guild named '%s'\n", e.Guild.Name)
	bot.guildID = e.Guild.ID
}

// onMessageCreate handles messages that the bot receives, and sends them
// to Minecraft via RCON.
func (bot *DiscordBot) onMessageCreate(e *gateway.MessageCreateEvent) {
	// Ignore messages that aren't from the configured channel
	if e.ChannelID.String() == Config.Discord.ChannelID {
		// Ignore messages from ourselves
		if e.Author.ID != bot.id && e.Message.WebhookID.String() == "" {
			// Check if the message is a bot command
			if strings.HasPrefix(e.Message.Content, "!") {
				c := make(chan bool)
				go parser.Parse(e.Message, bot.state, Config, c)
				// Don't go any further if the command was found and ran
				if <-c {
					return
				}
			}

			Log.Debugln("Received a message from Discord")

			// Get the name to use
			var name string
			if Config.Discord.UseMemberNicks {
				name = bot.getNickname(e.Author.ID)
			} else {
				name = e.Author.Username
			}

			// Print the URL if message contains an attachement but no message content
			if len(e.Message.Attachments) > 0 {
				if len(e.Content) == 0 {
					e.Content = e.Message.Attachments[0].URL
					if err := sendToMinecraft(e.Content, name); err != nil {
						Log.Errorf("Error sending command to RCON: %s\n", err)
					}
					return
				}
			}

			content := formatMessage(bot.state, e.Message)
			lines := strings.Split(content, "\n")

			// Send a separate message for each line
			for i := 0; i < len(lines); i++ {
				line := lines[i]
				// Split long lines into additional messages
				if len(line) > 100 {
					lines = append(lines, "")
					copy(lines[i+2:], lines[i+1:])
					lines[i+1] = line[100:]
					line = line[:100]
				}

				if err := sendToMinecraft(line, name); err != nil {
					Log.Errorf("Error sending command to RCON: %s\n", err)
				}
			}
		}
	}
}

// sendToDiscord sends a message from Minecraft to the configured
// Discord channel.
func (bot *DiscordBot) sendToDiscord(m *MinecraftMessage) {
	// Insert Discord mentions if configured and present
	if Config.Discord.AllowMentions {
		// Insert Discord mentions
		m.Message = bot.insertMentions(m.Message)
	}

	// Send the message to Discord either via webhook or normal channel message
	if Config.Discord.Webhook.Enabled {
		// Get the configured webhook
		id, token := matchWebhookURL(Config.Discord.Webhook.URL)
		if id == "" || token == "" {
			Log.Warnln("Invalid or undefined Discord webhook URL")
			return
		}

		// Attempt to get the webhook
		snowflake, err := discord.ParseSnowflake(id)
		if err != nil {
			Log.Errorf("Error parsing Webhook Snowflake: %s\n", err.Error())
		}

		webhook, err := bot.state.WebhookWithToken(snowflake, token)
		if err != nil {
			Log.Errorf("Error getting Discord webhook: %s\n", err.Error())
			return
		}

		// Form our webhook params
		params := bot.setWebhookParams(m)

		// Send to the webhook
		Log.Debugf("Sending to webhook: id='%s', token='%s'\n", id, token)
		if _, err := bot.state.ExecuteWebhook(webhook.ID, token, false, params); err != nil {
			Log.Errorf("Error sending data to Discord webhook: %s\n", err.Error())
		}
	} else {
		// Format the message for Discord
		formatted := fmt.Sprintf("**%s**: %s", m.Username, m.Message)

		// Send to the configured Discord channel
		if _, err := bot.state.Client.SendMessage(bot.channel, formatted, nil); err != nil {
			Log.Errorf("Error sending a message to Discord: %s\n", err.Error())
		}
	}
}

// getNickname gets the nickname of a Discord user in a Guild.
func (bot *DiscordBot) getNickname(id discord.Snowflake) string {
	var m *discord.Member

	// Look in the cached state for the Member
	m, _ = bot.state.Member(bot.guildID, id)

	// Make sure we do have a user
	if m == nil {
		return ""
	}

	if m.Nick == "" {
		return m.User.Username
	}

	return m.Nick
}

// getUserFromName gets the Discord user from a mention or username. The username
// can be only a partial username.
func (bot *DiscordBot) getUserFromName(text string) (target *discord.User) {
	// Look through all guild members in the state
	members, _ := bot.state.Members(bot.guildID)
	for _, u := range members {
		// Check if the name matches, case-insensitive
		if strings.EqualFold(u.User.Username, text) {
			target = &u.User
			break
		}
	}

	return target
}

// insertMentions looks for potential Discord mentions in a Minecraft chat
// message. If there are any, we will attempt to get the user being mentioned
// to get their mention string to put into the chat message.
func (bot *DiscordBot) insertMentions(msg string) string {
	// Split the message into words
	words := strings.Split(msg, " ")

	// Iterate over each word
	for _, word := range words {
		// Check if the word might be a mention
		if strings.HasPrefix(word, "@") {
			// Attempt to get the user
			user := bot.getUserFromName(word[1:])
			if user != nil {
				// Replace the word with the mention
				msg = strings.Replace(msg, word, user.Mention(), 1)
			}
		}
	}

	return msg
}

func matchWebhookURL(url string) (string, string) {
	wm := webhookRegex.FindStringSubmatch(url)

	// Make sure we have the correct number of parts (ID and token)
	if len(wm) != 3 {
		return "", ""
	}

	// Return the webhook ID and token
	return wm[1], wm[2]
}

// setWebhookParams sets the avater, username, and message for a webhook request.
func (bot *DiscordBot) setWebhookParams(m *MinecraftMessage) api.ExecuteWebhookData {
	// Get the avatar to use for this message
	var avatarURL string

	if m.Username == bot.name {
		// Use the bot's avatar
		avatarURL = bot.avatarURL
	} else {
		// Player's Minecraft head as the avatar
		avatarURL = fmt.Sprintf("https://minotar.net/helm/%s/256.png", m.Username)
	}

	return api.ExecuteWebhookData{
		Content:   m.Message,
		Username:  m.Username,
		AvatarURL: avatarURL,
	}
}

func formatMessage(state *state.State, message discord.Message) string {
	content := message.Content

	// Replace mentions
	for _, word := range strings.Split(content, " ") {
		if strings.HasPrefix(word, "<#") && strings.HasSuffix(word, ">") {
			// Get the ID from the mention string
			id := word[2 : len(word)-1]
			snowflake, _ := discord.ParseSnowflake(id)

			channel, err := state.Channel(snowflake)
			if err != nil {
				Log.Warnf("Error while getting channel from Discord: %s\n", err)
				continue
			}

			content = strings.Replace(content, fmt.Sprintf("<#%s>", id), fmt.Sprintf("#%s", channel.Name), -1)
		}
	}

	for _, member := range message.Mentions {
		content = strings.Replace(content, fmt.Sprintf("<@!%s>", member.ID), fmt.Sprintf("@%s", member.Username), -1)
	}

	// Escape quote characters
	content = strings.Replace(content, "\"", "\\\"", -1)

	return content
}

func sendToMinecraft(content, username string) error {
	// Format command to send to the Minecraft server
	command := fmt.Sprintf("tellraw @a %s", Config.Minecraft.TellrawTemplate)
	command = strings.Replace(command, "%username%", username, -1)
	command = strings.Replace(command, "%message%", content, -1)

	// Create RCON connection
	conn, err := rcon.Dial(Config.Minecraft.RconIP, Config.Minecraft.RconPort, Config.Minecraft.RconPassword)
	if err != nil {
		return err
	}
	defer conn.Close()

	// Authenticate to RCON
	if err = conn.Authenticate(); err != nil {
		return err
	}

	// Send the command to Minecraft
	if _, err := conn.SendCommand(command); err != nil {
		return err
	}

	return nil
}
