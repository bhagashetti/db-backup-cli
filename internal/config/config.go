package config

import (
	"encoding/json"
	"fmt"
	"os"
)

// BackupConfig represents backup configuration loaded from JSON file.
type BackupConfig struct {
	DBType       string `json:"dbType"`
	Host         string `json:"host"`
	Port         int    `json:"port"`
	User         string `json:"user"`
	Password     string `json:"password"`
	DBName       string `json:"dbName"`
	Output       string `json:"out"`
	Compress     bool   `json:"compress"`
	UseTimestamp bool   `json:"useTimestamp"`
	Encrypt      bool   `json:"encrypt"`
	EncryptKey   string `json:"encryptKey"`
	UploadS3     bool   `json:"uploadS3"`
	S3Bucket     string `json:"s3Bucket"`
	S3Region     string `json:"s3Region"`
	S3Prefix     string `json:"s3Prefix"`
}

// RestoreConfig represents restore configuration loaded from JSON file.
type RestoreConfig struct {
	DBType   string `json:"dbType"`
	Host     string `json:"host"`
	Port     int    `json:"port"`
	User     string `json:"user"`
	Password string `json:"password"`
	DBName   string `json:"dbName"`
	Input    string `json:"input"`
}

// LoadBackup reads and parses a backup config file.
func LoadBackup(path string) (*BackupConfig, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read backup config file: %w", err)
	}

	var cfg BackupConfig
	if err := json.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("parse backup config JSON: %w", err)
	}

	return &cfg, nil
}

// LoadRestore reads and parses a restore config file.
func LoadRestore(path string) (*RestoreConfig, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read restore config file: %w", err)
	}

	var cfg RestoreConfig
	if err := json.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("parse restore config JSON: %w", err)
	}

	return &cfg, nil
}
