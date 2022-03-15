package main

import (
	gitbackup "git-backup"
	"log"
	"os"
	"path/filepath"
)

func main() {
	config := loadConfig()
	for _, githubConfig := range config.Github {
		me, err := githubConfig.GetMe()
		if err != nil {
			log.Printf("Communication Error: %s", err)
			os.Exit(100)
		}
		log.Printf("=== %s ===", githubConfig.JobName)
		log.Printf("Authenticated as: %s", *me.Login)
		repos, err := githubConfig.GetRepos()
		if err != nil {
			log.Printf("Communication Error: %s", err)
			os.Exit(100)
		}
		for _, repo := range repos {
			log.Println(*repo.FullName)
			targetPath := filepath.Join("backup", githubConfig.JobName, *repo.FullName)
			err := os.MkdirAll(targetPath, os.ModePerm)
			if err != nil {
				log.Printf("Failed to create directory: %s", err)
				os.Exit(100)
			}
			err = githubConfig.CloneInto(repo, targetPath)
			if err != nil {
				log.Printf("Failed to clone: %s", err)
				os.Exit(100)
			}

		}
	}
}

func loadConfig() gitbackup.Config {
	// try config file in working directory
	config, err := gitbackup.LoadFile("./git-backup.yml")
	if os.IsNotExist(err) {
		log.Println("No config file found. Exiting...")
		os.Exit(1)
	} else if err != nil {
		log.Printf("Error: %s", err)
		os.Exit(1)
	}
	return config
}