package dolphin

import (
	"fmt"
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
	bot := &DiscordBot{
		watcher: &MinecraftWatcher{},
	}
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
			// Create RCON connection
			host := Config.Minecraft.RconIP
			port := Config.Minecraft.RconPort
			password := Config.Minecraft.RconPassword
			conn, err := rcon.Dial(host, port, password)
			if err != nil {
				Log.Errorf("Error opening RCON connection: %s\n", err.Error())
				return
			}
			// Authenticate to RCON
			if err = conn.Authenticate(); err != nil {
				Log.Errorf("Error authenticating with RCON: %s\n", err.Error())
				conn.Close()
				return
			}
			// Format the command to send to Minecraft
			cmd := fmt.Sprintf("tellraw @a %s", Config.Minecraft.TellrawTemplate)
			cmd = strings.Replace(cmd, "%username%", e.Author.Username, -1)
			cmd = strings.Replace(cmd, "%message%", e.Content, -1)
			// Send the command to Minecraft
			if _, err := conn.SendCommand(cmd); err != nil {
				Log.Errorf("Error sending command to Minecraft: %s\n", err.Error())
			}
			// Close the connection
			conn.Close()
		}
	}
}

func (d *DiscordBot) formatMessage(m *MinecraftMessage) string {
	// Insert Discord mentions
	if Config.Discord.AllowMentions {
		m.Message = d.insertMentions(m.Message)
	}
	return fmt.Sprintf("**%s**: %s", m.Username, m.Message)
}

// getUserFromName gets the Discord user from a mention or username. The username
// can be only a partial username
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

func (d *DiscordBot) sendToDiscord(msg *MinecraftMessage) {
	if Config.Discord.UseWebhooks {
		// Get the configured webhook
		id, token := matchWebhookURL(Config.Discord.WebhookURL)
		Log.Debugf("Sending to webhook: id='%s', token='%s'\n", id, token)
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
		params := d.setWebhookParams(msg)
		// Semd tp the webhook
		if _, err := d.session.WebhookExecute(webhook.ID, token, false, params); err != nil {
			Log.Errorf("Error sending data to Discord webhook: %s\n", err.Error())
		}
	} else {
		formatted := d.formatMessage(msg)
		if _, err := d.session.ChannelMessageSend(Config.Discord.ChannelID, formatted); err != nil {
			Log.Errorf("Error sending a message to Discord: %s\n", err.Error())
		}
	}
}

func matchWebhookURL(url string) (string, string) {
	wm := webhookRegex.FindStringSubmatch(url)
	if len(wm) != 3 {
		return "", ""
	}
	return wm[1], wm[2]
}

func (d *DiscordBot) setWebhookParams(m *MinecraftMessage) *discordgo.WebhookParams {
	if Config.Discord.AllowMentions {
		// Insert Discord mentions
		m.Message = d.insertMentions(m.Message)
	}
	// Get the avatar for this user
	var avatarURL string
	// Check if the message is from a player, or the server
	if m.Username == Config.Discord.BotName {
		avatarURL = "https://cdn6.aptoide.com/imgs/8/e/d/8ede957333544a11f75df4518b501bdb_icon.png?w=256"
	} else {
		avatarURL = fmt.Sprintf("https://minotar.net/helm/%s/256.png", m.Username)
	}
	return &discordgo.WebhookParams{
		Content:   m.Message,
		Username:  m.Username,
		AvatarURL: avatarURL,
	}
}
