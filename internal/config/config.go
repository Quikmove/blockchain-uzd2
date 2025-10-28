package config

import (
	"log"
	"os"
	"path/filepath"
	"strconv"
)

type Config struct {
	Version      uint32
	Difficulty   uint32
	Port         string
	NameListPath string
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
	cfg := &Config{
		Version:    uint32(parsedVersion),
		Difficulty: uint32(parsedDifficulty),
		Port:       port,
	}
	log.Println("Version:", cfg.Version)
	log.Println("Difficulty:", cfg.Difficulty)
	if root, err := findModuleRoot(); err == nil {
		cfg.NameListPath = filepath.Join(root, "assets", "name_list.txt")
	}
	return cfg
}

func findModuleRoot() (string, error) {
	dir, err := os.Getwd()
	if err != nil {
		return "", err
	}
	dir, err = filepath.Abs(dir)
	if err != nil {
		return "", err
	}
	for {
		mod := filepath.Join(dir, "go.mod")
		if _, err := os.Stat(mod); err == nil {
			return dir, nil
		} else if !os.IsNotExist(err) {
			return "", err
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			return "", os.ErrNotExist
		}
		dir = parent
	}
}
