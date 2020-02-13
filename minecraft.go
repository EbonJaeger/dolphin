package dolphin

import (
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/nxadm/tail"
)

// MinecraftWatcher watches for log lines from a Minecraft server.
type MinecraftWatcher struct {
	botName       string
	deathKeywords []string
	tail          *tail.Tail
}

// NewWatcher creates a new watcher with all of the Minecraft death message keywords.
func NewWatcher(botName string) *MinecraftWatcher {
	var deathKeywords = []string{"shot", "pricked", "walked into a cactus", "roasted", "drowned", "kinetic", "blew up", "blown up", "killed", "hit the ground", "fell", "doomed", "squashed", "magic", "flames", "burned", "walked into fire", "burnt", "bang", "lava", "lightning", "danger", "slain", "fireballed", "stung", "starved", "suffocated", "squished", "poked", "imapled", "didn't want to live", "withered", "pummeled", "died", "slain"}
	// Append any custom death keywords
	if Config.Minecraft.CustomDeathKeywords != nil {
		deathKeywords = append(deathKeywords, *Config.Minecraft.CustomDeathKeywords...)
	}
	return &MinecraftWatcher{
		botName:       botName,
		deathKeywords: deathKeywords,
	}
}

// Close stops the tail process and cleans up inotify file watches.
func (w *MinecraftWatcher) Close() error {
	err := w.tail.Stop()
	w.tail.Cleanup()
	return err
}

// Watch watches a log file for changes and sends Minecraft messages
// to the given channel.
func (w *MinecraftWatcher) Watch(c chan<- *MinecraftMessage) {
	if Config.Minecraft.UseLogFile {
		// Check that the log file exists
		if _, err := os.Stat(Config.Minecraft.LogFilePath); err == nil {
			Log.Infof("Using Minecraft log file at '%s'\n", Config.Minecraft.LogFilePath)
			// Start tailing the file
			var tailErr error
			w.tail, tailErr = tail.TailFile(Config.Minecraft.LogFilePath, tail.Config{
				Location: &tail.SeekInfo{
					Whence: io.SeekEnd,
				},
				ReOpen: true,
				Follow: true,
			})
			if tailErr != nil {
				Log.Fatalf("Error trying to tail log file: %s\n", tailErr.Error())
			}
			for {
				// Read line from the Tail channel
				if line := <-w.tail.Lines; line != nil {
					// Parse the line to see if it's a message we care about
					if msg := w.ParseLine(w.botName, line.Text); msg != nil {
						// Send the message through the channel
						c <- msg
					}
				}
			}
		} else {
			Log.Fatalf("Error opening log file: %s\n", err.Error())
		}
	}
}

// ParseLine parses a log line for various types of messages and
// returns a MinecraftMessage struct if it is a message we care about.
func (w *MinecraftWatcher) ParseLine(botName string, line string) *MinecraftMessage {
	// Trim the time and thread prefix
	line = line[33:len(line)]
	// Trim trailing whitespace
	line = strings.TrimSpace(line)
	// Check if the line is a chat message
	if strings.HasPrefix(line, "<") {
		// Split the message into parts
		parts := strings.SplitN(line, " ", 2)
		username := strings.TrimPrefix(parts[0], "<")
		username = strings.TrimSuffix(username, ">")
		message := parts[1]
		return &MinecraftMessage{
			Username: username,
			Message:  message,
		}
	}
	// Check for player join or leave
	if strings.Contains(line, "joined the game") || strings.Contains(line, "left the game") {
		return &MinecraftMessage{
			Username: botName,
			Message:  line,
		}
	}
	// Check if the line is an advancement message
	if isAdvancement(line) {
		return &MinecraftMessage{
			Username: botName,
			Message:  fmt.Sprintf(":partying_face: %s", line),
		}
	}
	// Check if the line is a death message
	for _, word := range w.deathKeywords {
		if strings.Contains(line, word) {
			return &MinecraftMessage{
				Username: botName,
				Message:  fmt.Sprintf(":skull: %s", line),
			}
		}
	}
	// Check if the server just finished starting
	if strings.HasPrefix(line, "Done (") {
		return &MinecraftMessage{
			Username: botName,
			Message:  ":white_check_mark: Server has started",
		}
	}
	// Check if the server is shutting down
	if strings.HasPrefix(line, "Stopping the server") {
		return &MinecraftMessage{
			Username: botName,
			Message:  ":x: Server is shutting down",
		}
	}
	// Doesn't match anything we care about
	return nil
}

func isAdvancement(line string) bool {
	return strings.Contains(line, "has made the advancement") ||
		strings.Contains(line, "has completed the challenge") ||
		strings.Contains(line, "has reached the goal")
}
