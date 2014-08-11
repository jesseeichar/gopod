package main

import (
	"os"
	"os/user"
	"path"
	"fmt"
	"gopod/opml"
	"io"
	"log"
)

const (
	GO_POD_DIR                = ".gopod"
	GO_POD_SUBSCRIPTIONS_FILE = "subscriptions.opml"
)

func main() {
	user, lookupErr := user.Current()
	if lookupErr != nil {
		panic("Unable to find current user due to error: " + lookupErr.Error())
	}
	configDir := openOrCreate(path.Join(user.HomeDir, GO_POD_DIR), func(path string) (*os.File, error) {
			os.MkdirAll(path, os.ModeDir)
			return os.Open(path)
		})
	defer configDir.Close()

	subscriptionsFile := openOrCreate(path.Join(configDir.Name(), GO_POD_SUBSCRIPTIONS_FILE), func(path string) (*os.File, error) {
			file, err := os.Create(path)
			if err != nil {
				return file, err
			}
			emptySubscription := opml.New()
			_, err = emptySubscription.Write(file)

			if err != nil {
				return nil, err
			}
			if err := file.Close(); err != nil {
				panic("Unable to flush subscriptions file on creation of an empty file: " + file.Name() + ". Error: " + err.Error())
			}

			return os.Open(path)
		})
	defer subscriptionsFile.Close()

	subscriptions, err := opml.ParseOpml(subscriptionsFile)
	if err == io.EOF {
		log.Fatal("Subscriptions file is not valid, file terminated unexpectedly.  Make sure that the file has a valid OPML format: " + subscriptionsFile.Name())
	} else if err != nil {
		panic("Unable to parse the subscriptions file: " + subscriptionsFile.Name() + " due to " + err.Error())
	}

	fmt.Print(subscriptions)
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
