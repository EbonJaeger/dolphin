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
	MessageOptions MessageConfig `toml:"message_options" comment:"Toggle whether certain messages are sent to the Discord channel"`
	Webhook        WebhookConfig
}

// MessageConfig holds settings for the messages that should be sent to Discord from Minecraft
type MessageConfig struct {
	ShowAdvancements bool `toml:"show_advancements"`
	ShowDeaths       bool `toml:"show_deaths"`
	ShowJoinsLeaves  bool `toml:"show_joins_and_leaves"`
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
