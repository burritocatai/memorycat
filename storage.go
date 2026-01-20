package main

import (
	"encoding/json"
	"os"
	"path/filepath"
)

type Command struct {
	Command     string `json:"command"`
	Description string `json:"description"`
}

type Storage struct {
	Commands []Command `json:"commands"`
}

func getConfigPath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	configDir := filepath.Join(home, ".config", "memorycat")
	if err := os.MkdirAll(configDir, 0755); err != nil {
		return "", err
	}
	return filepath.Join(configDir, "commands.json"), nil
}

func LoadCommands() (*Storage, error) {
	configPath, err := getConfigPath()
	if err != nil {
		return nil, err
	}

	data, err := os.ReadFile(configPath)
	if err != nil {
		if os.IsNotExist(err) {
			return &Storage{Commands: []Command{}}, nil
		}
		return nil, err
	}

	var storage Storage
	if err := json.Unmarshal(data, &storage); err != nil {
		return nil, err
	}

	return &storage, nil
}

func SaveCommands(storage *Storage) error {
	configPath, err := getConfigPath()
	if err != nil {
		return err
	}

	data, err := json.MarshalIndent(storage, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(configPath, data, 0644)
}
