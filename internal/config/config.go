package config

import (
	"github.com/caarlos0/env/v11"
)

// Config holds all application configuration.
type Config struct {
	Port          int    `env:"PORT" envDefault:"8080"`
	DatabasePath  string `env:"DB_PATH" envDefault:"./language-app.db"`
	UploadDir     string `env:"UPLOAD_DIR" envDefault:"./uploads"`
	MaxUploadSize int64  `env:"MAX_UPLOAD_SIZE" envDefault:"52428800"` // 50 MB in bytes
	TTSDir        string `env:"TTS_DIR" envDefault:"./uploads/tts-cache"`
	LTEndpoint    string `env:"LIBRETRANSLATE_ENDPOINT" envDefault:"http://localhost:5000"`
	LTApiKey      string `env:"LIBRETRANSLATE_API_KEY" envDefault:""`
}

// Load reads configuration from environment variables.
func Load() (*Config, error) {
	var cfg Config
	if err := env.Parse(&cfg); err != nil {
		return nil, err
	}
	return &cfg, nil
}
