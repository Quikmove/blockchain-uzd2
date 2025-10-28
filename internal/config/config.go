package config

import (
	"os"
	"strconv"
)

type Config struct {
	Version    uint32
	Difficulty uint32
	Port       string
}

func LoadConfig() *Config {
	version := os.Getenv("BLOCK_VERSION")
	difficulty := os.Getenv("BLOCK_DIFFICULTY")
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	parsedVersion, err := strconv.ParseUint(version, 10, 32)
	if err != nil {
		parsedVersion = 1
	}

	parsedDifficulty, err := strconv.ParseUint(difficulty, 10, 32)
	if err != nil {
		parsedDifficulty = 3
	}
	return &Config{
		Version:    uint32(parsedVersion),
		Difficulty: uint32(parsedDifficulty),
		Port:       port,
	}
}
