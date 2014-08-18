package main

import (
	"bufio"
	"gopod/config"
	"gopod/opml"
	"io"
	"log"
	"os"
	"fmt"
	"gopod/rss"
)

func backupConfigFile(configDirPath string) {
	configFile := config.ConfigFile(configDirPath)
	defer configFile.Close()

	backupFile := config.BackupConfigFile(configDirPath)
	defer backupFile.Close()

	n, err := bufio.NewReader(configFile).WriteTo(backupFile)
	if err != nil {
		panic("Error backing up subscriptions file: " + err.Error())
	}

	log.Printf("Write %d bytes from original subscription file to backup file.\n", n)
}

func loadConfig() (configModel *opml.Opml, configFilePath string) {
	configDirPath := config.ConfigPathInUserHome()

	backupConfigFile(configDirPath)

	configFile := config.ConfigFile(configDirPath)
	defer configFile.Close()

	opmlModel, err := opml.ParseOpml(configFile)
	if err == io.EOF {
		log.Fatal("Subscriptions file is not valid, file terminated unexpectedly.  Make sure that the file has a valid OPML format: " + configFile.Name())
	} else if err != nil {
		panic("Unable to parse the subscriptions file: " + configFile.Name() + " due to " + err.Error())
	}

	return opmlModel, configFile.Name()
}
func download(index int, configModel *opml.Opml, doneChannel chan error) {
	var err error
	defer func() { doneChannel <- err }()

	subscription := &configModel.Body.Outline[index]
	_, err = rss.Download(configModel.Head, subscription);
}
func writeUpdatedConfig(configModel *opml.Opml, configFile string) {
	file, err := os.Create(configFile)

	if err != nil {
		panic("Unable to create/truncate config file: " + err.Error())
	}

	if _, err = configModel.Write(file); err != nil {
		panic("Unable to write updated config file: " + err.Error())
	}
}
func main() {
	configModel, configFile := loadConfig()

	if configModel.Head.DownloadDir == "" {
		log.Fatalf("There is no DownloadDir element defined in head of %s", configFile)
	}

	if configModel.Head.DefaultKeep == 0 {
		configModel.Head.DefaultKeep = 1
	}
	done := make(chan error)
	for i, _ := range configModel.Body.Outline {
		go download(i, configModel, done)
	}

	errors := []error{}
	for i := 0; i < len(configModel.Body.Outline); i++ {
		if err := <-done; err != nil {
			errors = append(errors, err)
		}
	}

	writeUpdatedConfig(configModel, configFile)

	if len(errors) > 0 {
		fmt.Printf("%d errors occurred during execution: \n", len(errors))
		for _, err := range errors {
			fmt.Printf("\t%v\n", err)
		}
	}
}
