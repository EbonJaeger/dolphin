package dolphin

// Flags holds our command line flags.
type Flags struct {
	Config string `short:"c" long:"config" description:"Specify the path to the configuration file to use"`
	Debug  bool   `long:"debug" description:"Print additional debug lines to stdout"`
}

// MinecraftMessage represents a message from Minecraft to be sent to Discord.
type MinecraftMessage struct {
	Username string
	Message  string
}
