package config

import (
	"gopod/opml"
	"os"
	"os/user"
	"path/filepath"
)

const (
	GO_POD_DIR         = ".gopod"
	CONFIG_FILE        = "config.xml"
	CONFIG_BACKUP_FILE = "config-backup.xml"
)

func ConfigPathInUserHome() string {
	currentUser, lookupErr := user.Current()
	if lookupErr != nil {
		panic("Unable to find current user due to error: " + lookupErr.Error())
	}
	return filepath.Join(currentUser.HomeDir, GO_POD_DIR)
}

func openOrCreate(configDirName string, createFunc func(string) (*os.File, error)) *os.File {
	if dir, err := os.Open(configDirName); os.IsNotExist(err) {
		if dir, err = createFunc(configDirName); err != nil {
			panic("Unable to create: " + configDirName)
		}
		return dir
	} else if err == nil {
		return dir
	} else {
		panic("Unable to open " + configDirName)
	}
}

func openConfigDir(configDirPath string) *os.File {
	configDir := openOrCreate(configDirPath, func(path string) (*os.File, error) {
		os.MkdirAll(path, os.ModeDir)
		return os.Open(path)
	})
	return configDir
}

func openOrCreateConfigFile(path string) (*os.File, error) {
	file, err := os.Create(path)
	if err != nil {
		return file, err
	}
	emptyConfig := opml.New()
	_, err = emptyConfig.Write(file)

	if err != nil {
		return nil, err
	}
	if err := file.Close(); err != nil {
		panic("Unable to flush config file on creation of an empty file: " + file.Name() + ". Error: " + err.Error())
	}

	return os.Open(path)
}

func ConfigFile(configDirPath string) *os.File {
	configDir := openConfigDir(configDirPath)
	defer configDir.Close()

	return openOrCreate(filepath.Join(configDir.Name(), CONFIG_FILE), openOrCreateConfigFile)
}

func BackupConfigFile(configDirPath string) *os.File {
	configDir := openConfigDir(configDirPath)
	defer configDir.Close()

	backupFile := filepath.Join(configDirPath, CONFIG_BACKUP_FILE)
	file, err := os.Create(backupFile)
	if err != nil {
		panic("Unable to create backup file: " + backupFile)
	}
	return file
}
