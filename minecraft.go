package dolphin

import (
	"fmt"
	"os"
	"strings"

	"github.com/hpcloud/tail"
)

// ParseLine parses a log line for various types of messages and
// returns a MinecraftMessage struct if it is a message we care about.
func ParseLine(line string) MinecraftMessage {
	// Trim the time and thread prefix
	line = line[33 : len(line)-1]
	// Trim trailing whitespace
	line = strings.TrimSpace(line)
	// Check if the line is a chat message
	if strings.HasPrefix(line, "<") {
		Log.Debugln("Matched a chat message from Minecraft")
		// Split the message into parts
		parts := strings.SplitN(line, " ", 2)
		username := strings.TrimPrefix(parts[0], "<")
		username = strings.TrimSuffix(username, ">")
		message := parts[1]
		return MinecraftMessage{
			Username: username,
			Message:  message,
		}
	}
	// Check for player join or leave
	if strings.Contains(line, "joined the game") || strings.Contains(line, "left the game") {
		Log.Debugln("Matched a join or leave message from Minecraft")
		return MinecraftMessage{
			Username: Config.Discord.BotName,
			Message:  line,
		}
	}
	// Check if the line is an advancement message
	if isAdvancement(line) {
		Log.Debugln("Matched an advancement message from Minecraft")
		return MinecraftMessage{
			Username: Config.Discord.BotName,
			Message:  fmt.Sprintf(":partying_face: %s", line),
		}
	}
	// Check if the line is a player death
	for _, word := range *Config.Minecraft.DeathKeywords {
		if strings.Contains(line, word) {
			Log.Debugf("Matched a death message from Minecraft on the word '%s'\n", word)
			return MinecraftMessage{
				Username: Config.Discord.BotName,
				Message:  fmt.Sprintf(":skull: %s", line),
			}
		}
	}
	// Check if the server just finished starting
	if strings.HasPrefix(line, "Done (") {
		// TODO: debug log
		return MinecraftMessage{
			Username: Config.Discord.BotName,
			Message:  ":white_check_mark: Server has started",
		}
	}
	// Check if the server is shutting down
	if strings.HasPrefix(line, "Stopping the server") {
		// TODO: debug log
		return MinecraftMessage{
			Username: Config.Discord.BotName,
			Message:  ":x: Server is shutting down",
		}
	}
	// Doesn't match anything we care about
	return MinecraftMessage{}
}

// Watch watches a log file for changes and calls a provided
// callback function with the message in the line, or an
// empty message.
func Watch(callback func(msg MinecraftMessage)) {
	if Config.Minecraft.UseLogFile {
		// Check that the log file exists
		if _, err := os.Stat(Config.Minecraft.LogFilePath); err == nil {
			t, err := tail.TailFile(Config.Minecraft.LogFilePath, tail.Config{
				MustExist: true,
				Follow:    true,
				Logger:    Log,
			})
			if err != nil {
				Log.Fatalf("Error trying to tail log file: %s\n", err.Error())
			}
			for line := range t.Lines {
				message := ParseLine(line.Text)
				callback(message)
			}
		} else {
			Log.Fatalf("Error opening log file: %s\n", err.Error())
		}
	}
}

func isAdvancement(line string) bool {
	return strings.Contains(line, "has made the advancement") ||
		strings.Contains(line, "has completed the challenge") ||
		strings.Contains(line, "has reached the goal")
}
