package main

import (
	"flag"
	gitbackup "git-backup"
	"log"
	"os"
	"path/filepath"
)

var configFilePath = flag.String("config.file", "git-backup.yml", "The path to your config file.")
var targetPath = flag.String("backup.path", "backup", "The target path to the backup folder.")

func main() {
	flag.Parse()

	config := loadConfig()
	for _, source := range config.GetSources() {
		sourceName := source.GetName()
		log.Printf("=== %s ===", sourceName)
		if err := source.Test(); err != nil {
			log.Printf("Failed to verify connection to job [%s]: %s", sourceName, err)
			os.Exit(110)
		}
		repos, err := source.ListRepositories()
		if err != nil {
			log.Printf("Communication Error: %s", err)
			os.Exit(100)
		}
		for _, repo := range repos {
			log.Printf("Discovered %s", repo.FullName)
			targetPath := filepath.Join(*targetPath, sourceName, repo.FullName)
			err := os.MkdirAll(targetPath, os.ModePerm)
			if err != nil {
				log.Printf("Failed to create directory: %s", err)
				os.Exit(100)
			}
			err = repo.CloneInto(targetPath)
			if err != nil {
				log.Printf("Failed to clone: %s", err)
				os.Exit(100)
			}
		}
	}
}

func loadConfig() gitbackup.Config {
	// try config file in working directory
	config, err := gitbackup.LoadFile(*configFilePath)
	if os.IsNotExist(err) {
		log.Println("No config file found. Exiting...")
		os.Exit(1)
	} else if err != nil {
		log.Printf("Error: %s", err)
		os.Exit(1)
	}
	return config
}