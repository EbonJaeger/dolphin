package dolphin

import (
	"fmt"
	"strings"

	"github.com/EbonJaeger/dolphin/rcon"
	"github.com/bwmarrin/discordgo"
)

// OnReady starts our Minecraft watcher when the bot is ready.
func OnReady(s *discordgo.Session, e *discordgo.Ready) {
	// Set the bot gaming status
	s.UpdateStatus(0, "Bridging the Minecraft/Discord gap")
}

// OnGuildCreate handles when the bot joins a Guild.
func OnGuildCreate(s *discordgo.Session, e *discordgo.GuildCreate) {
	// Make sure the guild is available
	if e.Guild.Unavailable {
		Log.Warnf("Attempted to join Guild '%s', but it was unavailable\n", e.Guild.Name)
		return
	}
	Log.Infof("Connected to guild named '%s'", e.Guild.Name)
	// Start our Minecraft watcher with our callback
	Watch(func(msg MinecraftMessage) {
		Log.Debugf("Received a line from Minecraft: Username='%s', Text='%s'", msg.Username, msg.Message)
		if Config.Discord.UseWebhooks {
			// Try to get our webhook by ID
			webhook, _ := s.Webhook(Config.Discord.WebhookID)
			if webhook == nil {
				// Try to create a new webhook for us to use
				var err error
				webhook, err = s.WebhookCreate(Config.Discord.ChannelID, Config.Discord.BotName, "")
				if err != nil {
					Log.Errorf("Error creating a new webhook: %s\n", err.Error())
					return
				}
				// Save the webhook ID in the config
				Config.Discord.WebhookID = webhook.ID
				SaveConfig(Config)
			}
			// Form our webhook params
			params := setWebhookParams(s, e.Guild, msg)
			// Semd tp the webhook
			if _, err := s.WebhookExecute(webhook.ID, Config.Discord.BotToken, false, params); err != nil {
				Log.Errorf("Error sending data to Discord webhook: %s\n", err.Error())
			}
		} else {
			formatted := formatMessage(s, e.Guild, msg)
			if _, err := s.ChannelMessageSend(Config.Discord.ChannelID, formatted); err != nil {
				Log.Errorf("Error sending a message to Discord: %s\n", err.Error())
			}
		}
	})
}

// OnMessageCreate handles messages that the bot receives, and sends them
// to Minecraft via RCON.
func OnMessageCreate(s *discordgo.Session, e *discordgo.MessageCreate) {
	// Ignore messages that aren't from the configured channel
	if e.ChannelID == Config.Discord.ChannelID {
		// Ignore messages from ourselves
		if e.Author.ID != s.State.User.ID {
			Log.Debugln("Received a message from Discord")
			// Create RCON connection
			host := Config.Minecraft.RconIP
			port := Config.Minecraft.RconPort
			password := Config.Minecraft.RconPassword
			conn, err := rcon.Dial(host, port, password)
			if err != nil {
				Log.Errorf("Error opening RCON connectino to '%s:%d': %s\n", host, port, err.Error())
				return
			}
			// Authenticate to RCON
			if err = conn.Authenticate(); err != nil {
				Log.Errorf("Error authenticating with RCON at '%s:%d': %s\n", host, port, err.Error())
				conn.Close()
				return
			}
			// Format the command to send to Minecraft
			cmd := fmt.Sprintf("tellraw @a %s", Config.Discord.MessageTemplate)
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

func formatMessage(s *discordgo.Session, g *discordgo.Guild, m MinecraftMessage) string {
	// Insert Discord mentions
	if Config.Discord.AllowMentions {
		m.Message = insertMentions(s, g, m.Message)
	}
	// Format message using the configured template
	f := Config.Discord.MessageTemplate
	strings.Replace(f, "%username%", m.Username, -1)
	strings.Replace(f, "%message%", m.Message, 1)
	return f
}

// getUserFromName gets the Discord user from a mention or username. The username
// can be only a partial username
func getUserFromName(s *discordgo.Session, g *discordgo.Guild, t string) *discordgo.User {
	var target *discordgo.User
	// Check if it's a mention
	if strings.HasPrefix(t, "<@!") {
		// Trim the mention prefix and suffix
		id := strings.TrimPrefix(t, "<@!")
		id = strings.TrimSuffix(id, ">")
		// Get the user
		target, _ = s.User(id)
	} else {
		// Look through all users in the guild
		for _, u := range g.Members {
			// Check if the name matches or is a partial
			if strings.Contains(strings.ToLower(u.User.Username), strings.ToLower(t)) {
				target = u.User
				break
			}
		}
	}

	return target
}

func insertMentions(s *discordgo.Session, g *discordgo.Guild, m string) string {
	// Split the message into words
	words := strings.Split(m, " ")
	// Iterate over each word
	for _, word := range words {
		// Check if the word might be a mention
		if strings.HasPrefix(word, "@") {
			// Attempt to get the user
			user := getUserFromName(s, g, word[1:])
			if user != nil {
				// Replace the word with the mention
				strings.Replace(m, word, user.Mention(), 1)
			}
		}
	}
	return m
}

func setWebhookParams(s *discordgo.Session, g *discordgo.Guild, m MinecraftMessage) *discordgo.WebhookParams {
	if Config.Discord.AllowMentions {
		// Insert Discord mentions
		m.Message = insertMentions(s, g, m.Message)
	}
	// Get the avatar for this user
	var avatarURL string
	// Check if the message is from a player, or the server
	if m.Username == Config.Discord.BotName {
		avatarURL = "https://cdn6.aptoide.com/imgs/8/e/d/8ede957333544a11f75df4518b501bdb_icon.png?w=256"
	} else {
		avatarURL = fmt.Sprintf("https://minotaur.net/helm/%s/256.png", m.Username)
	}
	return &discordgo.WebhookParams{
		Content:   m.Message,
		Username:  m.Username,
		AvatarURL: avatarURL,
	}
}
