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
	"path/filepath"
	"sort"
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
func list(filename string) ([]os.FileInfo, error) {
	downloadDir, err := os.Open(filename)
	if err != nil {
		return nil, err
	}

	dirs, err := downloadDir.Readdir(0)
	if err != nil {
		return nil, err
	}

	return dirs, nil
}
type SortablePodcast []os.FileInfo
func (s SortablePodcast) Len() int {
	return len(s)
}
func (s SortablePodcast) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}
func (s SortablePodcast) Less(i, j int) bool {
	return s[i].ModTime().After(s[j].ModTime())
}
func deleteOutOfDateFiles(configModel *opml.Opml) error {
	typeNames, err := list(configModel.Head.DownloadDir)
	if err != nil {
		return fmt.Errorf("deleteOutOfDateFile: Unable to list directories in downloadDir: %v", err)
	}

	for  _, dirName := range typeNames {

		channels, err := list (dirName.Name())
		if err != nil {
			return fmt.Errorf("deleteOutOfDateFile: Unable to list channels in %v: %v", dirName.Name(), err)
		}

		for _, channel := range channels {

			title := filepath.Base(channel.Name())
			outline := configModel.Body.Get(title)
			podcasts, err := list(channel.Name())

			if err != nil {
				return fmt.Errorf("deleteOutOfDateFile: Unable to list podcasts in %v: %v", channel.Name(), err)
			}
			sort.Sort(SortablePodcast(podcasts))
			for  i := range podcasts {
				if i >= outline.Keep {
					fmt.Println("Deleting old podcast: %d\n", i)
				}
			}
		}
	}
	return nil
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

	if err := deleteOutOfDateFiles(configModel); err != nil {
		panic(err.Error())
	}

	if len(errors) > 0 {
		fmt.Printf("%d errors occurred during execution: \n", len(errors))
		for _, err := range errors {
			fmt.Printf("\t%v\n", err)
		}
	}
}
