package config

// RootConfig is our root config struct.
type RootConfig struct {
	Discord   DiscordConfig
	Minecraft MinecraftConfig
}

// DiscordConfig holds all settings for the Discord side of the application.
type DiscordConfig struct {
	BotToken       string
	ChannelID      string
	AllowMentions  bool
	UseMemberNicks bool
	Webhook        WebhookConfig
}

// WebhookConfig holds settings for using Discord webhooks to send messages.
type WebhookConfig struct {
	Enabled bool
	URL     string
}

// MinecraftConfig holds all settings for the Minecraft server side of the application.
type MinecraftConfig struct {
	RconIP              string
	RconPort            int
	RconPassword        string
	TellrawTemplate     string
	CustomDeathKeywords *[]string
	UseLogFile          bool
	LogFilePath         string
}
