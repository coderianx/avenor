package config

import (
	"fmt"
	"os"
	"path/filepath"
)

// default providers
var (
	GroqBaseURL       string = "https://api.groq.com/openai/v1"
	OpenrouterBaseURL string = "https://openrouter.ai/api/v1"
)

var avenorPath string

func InitAvenor() error {
	home, err := os.UserHomeDir()

	if err != nil {
		return fmt.Errorf("config: get user home dir: %w", err)
	}

	avenorPath = filepath.Join(home, ".avenor")

	err = os.MkdirAll(avenorPath, 0700)

	if err != nil {
		return fmt.Errorf("config: create avenor path folder: %w", err)
	}

	// Tighten permissions on directories created by older versions.
	if err := os.Chmod(avenorPath, 0700); err != nil {
		return fmt.Errorf("config: set avenor path permissions: %w", err)
	}

	return nil
}

func GetAvenorPath() string {
	return avenorPath
}
