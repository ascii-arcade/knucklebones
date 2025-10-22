package config

import (
	"cmp"
	"os"
	"strconv"
	"time"

	"github.com/ascii-arcade/knucklebones/language"
	"github.com/charmbracelet/log"
)

const (
	MinimumHeight = 33
	MinimumWidth  = 120
)

var (
	Language *language.Language = setDefaultLanguage()

	Debug    bool   = cmp.Or(os.Getenv("ASCII_ARCADE_DEBUG"), "false") == "true"
	Host     string = cmp.Or(os.Getenv("ASCII_ARCADE_HOST"), "localhost")
	SSHPort  string = cmp.Or(os.Getenv("ASCII_ARCADE_SSH_PORT"), "2222")
	HTTPPort string = cmp.Or(os.Getenv("ASCII_ARCADE_HTTP_PORT"), "8080")

	Database    string = cmp.Or(os.Getenv("ASCII_ARCADE_DB_NAME"), "knucklebones")
	DatabaseURI string = cmp.Or(os.Getenv("ASCII_ARCADE_DB_URI"), "mongodb://localhost:27017")

	PlayerTimeoutDurationMinutes = cmp.Or(os.Getenv("ASCII_ARCADE_PLAYER_TIMEOUT_DURATION"), "30")

	Version string = "dev"
)

func GetPlayerTimeoutDuration() time.Duration {
	duration, err := strconv.Atoi(PlayerTimeoutDurationMinutes)
	if err != nil {
		log.Warn("Invalid PLAYER_TIMEOUT_DURATION, defaulting to 30 minutes", "value", PlayerTimeoutDurationMinutes)
		return 30 * time.Minute
	}
	return time.Duration(duration) * time.Minute
}

func setDefaultLanguage() *language.Language {
	langCode := cmp.Or(os.Getenv("ASCII_ARCADE_LANG"), "EN")
	lang, exists := language.Languages[langCode]
	if !exists {
		log.Warn("Unknown language code %s, defaulting to English", langCode)
		return language.Languages["EN"]
	}
	return lang
}
