package config

import (
	"os"

	"github.com/BurntSushi/toml"
)

type Config struct {
	DefaultCommand string `toml:"default_command"`
	StartPort      int    `toml:"start_port"`
	Browser        string `toml:"browser"`  // "chrome" | "chromium" | "auto"
	ProfilesDir    string `toml:"profiles_dir"`
	BrowserPath    string `toml:"browser_path"` // optional override
}

func defaults() Config {
	return Config{
		DefaultCommand: "pnpm run dev",
		StartPort:      3000,
		Browser:        "auto",
		ProfilesDir:    "~/.devbrowser/profiles",
	}
}

func Load() (Config, error) {
	cfg := defaults()

	path, err := ConfigFile()
	if err != nil {
		return cfg, err
	}

	if _, err := os.Stat(path); os.IsNotExist(err) {
		// Write default config on first run
		_ = writeDefaults(path, cfg)
	} else if _, err := toml.DecodeFile(path, &cfg); err != nil {
		return cfg, err
	}

	cfg.ProfilesDir = ExpandHome(cfg.ProfilesDir)
	return cfg, nil
}

func writeDefaults(path string, cfg Config) error {
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()
	return toml.NewEncoder(f).Encode(cfg)
}
