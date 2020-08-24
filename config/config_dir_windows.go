// +build windows

package config

import (
	"os/user"
	"path/filepath"
)

// GetDefaultConfDir attempts to get the default configuration directory.
// If the directory does not exist, we will attempt to create it. The default
// directory location varies by platform. For each, they are:
//
// Windows: C:\Users\<username>\AppData\Local\dolphin
// macOS: ~/Library/Application Support/dolphin
// Linux: $HOME/.config/dolphin
func GetDefaultConfDir() (string, error) {
	// Get the current user
	user, err := user.Current()
	if err != nil {
		return "", err
	}

	// Get our default config directory
	return filepath.Join(user.HomeDir, "AppData", "Local", "Dolphin"), nil
}
