package main

import (
	"io"
	"log"
	"gopod/opml"
	"gopod/config"
	"bufio"
	"os"
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
	configDirPath := config.ConfigPathInUserHome();

	backupConfigFile(configDirPath)

	configFile := config.ConfigFile(configDirPath)
	defer configFile.Close()


	opmlModel, err := opml.ParseOpml(configFile)
	if err == io.EOF {
		log.Fatal("Subscriptions file is not valid, file terminated unexpectedly.  Make sure that the file has a valid OPML format: " + configFile.Name())
	} else if err != nil {
		panic("Unable to parse the subscriptions file: " + configFile.Name() + " due to " + err.Error())
	}

	return opmlModel, configFile.Name();
}
func download(index int, configModel *opml.Opml, doneChannel chan bool) {
	defer func() {doneChannel <- true}()

	subscription := &configModel.Body.Outline[index]
	rss.Download(configModel.Head, subscription)
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
		configModel.Head.DefaultKeep = 1;
	}
	done := make(chan bool)
	for i, _ := range configModel.Body.Outline {
		go download(i, configModel, done)
	}

	for i := 0; i < len(configModel.Body.Outline); i++ {
		<-done
	}

	writeUpdatedConfig(configModel, configFile)
}
