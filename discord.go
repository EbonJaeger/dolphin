package dolphin

import (
	"fmt"
	"os"
	"regexp"
	"strings"

	"github.com/EbonJaeger/dolphin/rcon"
	"github.com/bwmarrin/discordgo"
)

var webhookRegex = regexp.MustCompile("https://discordapp.com/api/webhooks/(.*)/(.*)")

// DiscordBot holds our Discord session and Minecraft log watcher.
type DiscordBot struct {
	guildID string
	session *discordgo.Session
	watcher *MinecraftWatcher
}

// NewDiscordBot creates a new DiscordBot with a MinecraftWatcher and
// connects to Discord.
func NewDiscordBot() (*DiscordBot, error) {
	bot := &DiscordBot{}
	var discordErr error
	// Create Discord session
	bot.session, discordErr = discordgo.New("Bot " + Config.Discord.BotToken)
	bot.session.StateEnabled = true
	if discordErr != nil {
		return nil, discordErr
	}
	// Add our Discord handlers
	bot.session.AddHandler(bot.onReady)
	bot.session.AddHandler(bot.onGuildCreate)
	bot.session.AddHandler(bot.onMessageCreate)
	// Connect to Discord websocket
	discordErr = bot.session.Open()
	// Create our Minecraft log watcher
	bot.watcher = NewWatcher(bot.session.State.User.Username)
	return bot, discordErr
}

// Close cleans up the watcher and closes the Discord session.
func (d *DiscordBot) Close() error {
	var closeErr error
	if err := d.watcher.Close(); err != nil {
		closeErr = err
	}
	if err := d.session.Close(); err != nil {
		closeErr = err
	}
	return closeErr
}

// WaitForMessages starts the Minecraft log watcher and waits for messages
// on a messages channel.
func (d *DiscordBot) WaitForMessages() {
	// Make our messages channel
	mc := make(chan *MinecraftMessage)
	// Start our Minecraft watcher
	go d.watcher.Watch(mc)
	for {
		// Read message from the channel
		msg := <-mc
		Log.Debugf("Received a line from Minecraft: Username='%s', Text='%s'\n", msg.Username, msg.Message)
		// Send the message to the Discord channel
		d.sendToDiscord(msg)
	}
}

// onReady starts our Minecraft watcher when the bot is ready.
func (d *DiscordBot) onReady(s *discordgo.Session, e *discordgo.Ready) {
	// Set the bot gaming status
	s.UpdateStatus(0, "Bridging the Minecraft/Discord gap")
}

// onGuildCreate handles when the bot joins a Guild.
func (d *DiscordBot) onGuildCreate(s *discordgo.Session, e *discordgo.GuildCreate) {
	// Make sure the guild is available
	if e.Guild.Unavailable {
		Log.Warnf("Attempted to join Guild '%s', but it was unavailable\n", e.Guild.Name)
		return
	}
	// Make sure we are only connected to one Guild
	if d.guildID != "" {
		Log.Errorln("Already connected to a guild! Aborting...")
		d.Close()
		os.Exit(1)
	}
	// Set our Guild ID
	d.guildID = e.Guild.ID
	Log.Infof("Connected to guild named '%s'\n", e.Guild.Name)
}

// onMessageCreate handles messages that the bot receives, and sends them
// to Minecraft via RCON.
func (d *DiscordBot) onMessageCreate(s *discordgo.Session, e *discordgo.MessageCreate) {
	// Ignore messages that aren't from the configured channel
	if e.ChannelID == Config.Discord.ChannelID {
		// Ignore messages from ourselves
		if e.Author.ID != s.State.User.ID && e.Message.WebhookID == "" {
			Log.Debugln("Received a message from Discord")
			// Get the name to use
			var name string
			if Config.Discord.UseMemberNicks {
				name = d.getNickname(e.Author.ID)
			} else {
				name = e.Author.Username
			}
			// Format the command to send to Minecraft
			cmd := fmt.Sprintf("tellraw @a %s", Config.Minecraft.TellrawTemplate)
			cmd = strings.Replace(cmd, "%username%", name, -1)
			cmd = strings.Replace(cmd, "%message%", e.Content, -1)
			// Create RCON connection
			conn, err := rcon.Dial(Config.Minecraft.RconIP, Config.Minecraft.RconPort, Config.Minecraft.RconPassword)
			if err != nil {
				Log.Errorf("Error opening RCON connection: %s\n", err.Error())
				return
			}
			defer conn.Close()
			// Authenticate to RCON
			if err = conn.Authenticate(); err != nil {
				Log.Errorf("Error authenticating with RCON: %s\n", err.Error())
				return
			}
			// Send the command to Minecraft
			if _, err := conn.SendCommand(cmd); err != nil {
				Log.Errorf("Error sending command to RCON: %s\n", err.Error())
			}
		}
	}
}

// sendToDiscord sends a message from Minecraft to the configured
// Discord channel.
func (d *DiscordBot) sendToDiscord(m *MinecraftMessage) {
	// Insert Discord mentions if configured and present
	if Config.Discord.AllowMentions {
		// Insert Discord mentions
		m.Message = d.insertMentions(m.Message)
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
		webhook, err := d.session.WebhookWithToken(id, token)
		if err != nil {
			Log.Errorf("Error getting Discord webhook: %s\n", err.Error())
			return
		}
		// Form our webhook params
		params := d.setWebhookParams(m)
		// Semd tp the webhook
		Log.Debugf("Sending to webhook: id='%s', token='%s'\n", id, token)
		if _, err := d.session.WebhookExecute(webhook.ID, token, false, params); err != nil {
			Log.Errorf("Error sending data to Discord webhook: %s\n", err.Error())
		}
	} else {
		// Format the message for Discord
		formatted := fmt.Sprintf("**%s**: %s", m.Username, m.Message)
		// Send to the configured Discord channel
		if _, err := d.session.ChannelMessageSend(Config.Discord.ChannelID, formatted); err != nil {
			Log.Errorf("Error sending a message to Discord: %s\n", err.Error())
		}
	}
}

// getNickname gets the nickname of a Discord user in a Guild.
func (d *DiscordBot) getNickname(id string) string {
	var m *discordgo.Member
	// Look in the cached state for the Member
	m, _ = d.session.State.Member(d.guildID, id)
	// Check if we got it
	if m == nil {
		// Look up the member normally
		m, _ = d.session.GuildMember(d.guildID, id)
	}
	// Make sure we do have a user
	if m == nil {
		return ""
	}
	return m.Nick
}

// getUserFromName gets the Discord user from a mention or username. The username
// can be only a partial username.
func (d *DiscordBot) getUserFromName(text string) *discordgo.User {
	var target *discordgo.User
	// Look through all cached guild members in the state
	gCache, _ := d.session.State.Guild(d.guildID)
	for _, u := range gCache.Members {
		// Check if the name matches or is a partial
		if strings.Contains(strings.ToLower(u.User.Username), strings.ToLower(text)) {
			target = u.User
			break
		}
	}
	if target == nil {
		// Look through all guild members
		g, _ := d.session.Guild(d.guildID)
		for _, u := range g.Members {
			// Check if the name matches or is a partial
			if strings.Contains(strings.ToLower(u.User.Username), strings.ToLower(text)) {
				target = u.User
				break
			}
		}
	}
	return target
}

// insertMentions looks for potential Discord mentions in a Minecraft chat
// message. If there are any, we will attempt to get the user being mentioned
// to get their mention string to put into the chat message.
func (d *DiscordBot) insertMentions(msg string) string {
	// Split the message into words
	words := strings.Split(msg, " ")
	// Iterate over each word
	for _, word := range words {
		// Check if the word might be a mention
		if strings.HasPrefix(word, "@") {
			// Attempt to get the user
			user := d.getUserFromName(word[1:])
			if user != nil {
				// Replace the word with the mention
				strings.Replace(msg, word, user.Mention(), 1)
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
func (d *DiscordBot) setWebhookParams(m *MinecraftMessage) *discordgo.WebhookParams {
	// Get the avatar to use for this message
	var avatarURL string
	if m.Username == d.session.State.User.Username {
		// Use the bot's avatar
		avatarURL = d.session.State.User.AvatarURL("256")
	} else {
		// Player's Minecraft head as the avatar
		avatarURL = fmt.Sprintf("https://minotar.net/helm/%s/256.png", m.Username)
	}
	return &discordgo.WebhookParams{
		Content:   m.Message,
		Username:  m.Username,
		AvatarURL: avatarURL,
	}
}
