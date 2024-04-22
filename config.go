package main

import (
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

type GitLabConfig struct {
	URL             string   `yaml:"URL"`
	PersonalToken   string   `yaml:"PersonalToken"`
	ProjectId       uint64   `yaml:"ProjectId"`
	TrackLabels     []string `yaml:"TrackLabels"`
	TrackIssueTypes []string `yaml:"TrackIssueTypes"`
}

type AppConfig struct {
	Version      string       `yaml:"Version"`
	GitLabConfig GitLabConfig `yaml:"gitLab"`
}

var configFolder = ".gitlab-sprint-report"
var configFilename = "gitlab-sprint-config.yml"
var version = "v1"

func readConfigFile(filePath string, appConfig *AppConfig) error {
	f, err := os.Open(filePath)
	if err != nil {
		return err
	}
	defer f.Close()

	decoder := yaml.NewDecoder(f)
	err = decoder.Decode(appConfig)
	if err != nil {
		return err
	}
	return nil
}

func saveNewConfigFile(filePath string) error {
	emptyConfig := AppConfig{
		Version: version,
		GitLabConfig: GitLabConfig{
			URL:             "https://gitlab.com",
			PersonalToken:   "your-personal-token",
			ProjectId:       0,
			TrackLabels:     []string{"To do", "In Progress", "QA", "Done"},
			TrackIssueTypes: []string{"issue", "incident"},
		},
	}
	fileContent, err := yaml.Marshal(emptyConfig)
	if err != nil {
		return err
	}

	err = os.WriteFile(filePath, fileContent, os.ModePerm)
	if err != nil {
		return err
	}
	return nil
}

func validateConfig(config AppConfig) error {
	if config.Version != version {
		return fmt.Errorf("config file has the wrong version. Please, backup the content and delete it, so you can generate a new one")
	}
	_, err := url.ParseRequestURI(config.GitLabConfig.URL)
	if err != nil {
		return fmt.Errorf("GitLabConfig.URL is not a valid URL")
	}
	if strings.HasSuffix(config.GitLabConfig.URL, "/") {
		return fmt.Errorf("GitLabConfig.URL can't end with a '/'")
	}
	if len(config.GitLabConfig.PersonalToken) == 0 {
		return fmt.Errorf("GitLabConfig.PersonalToken can't be empty")
	}
	if len(config.GitLabConfig.TrackLabels) == 0 {
		return fmt.Errorf("GitLabConfig.TrackLabels shouldn't be empty because it is going to be used to filter issues")
	}
	if len(config.GitLabConfig.TrackIssueTypes) == 0 {
		return fmt.Errorf("GitLabConfig.TrackIssueTypes shouldn't be empty because it is going to be used to filter issues")
	}
	return nil
}

func loadConfig() (*AppConfig, error) {
	usrDir, err := os.UserHomeDir()
	if err != nil {
		return nil, err
	}
	configDir := filepath.Join(usrDir, configFolder)
	_, err = os.Stat(configDir)
	if err != nil {
		os.Mkdir(configDir, os.ModePerm)
	}
	configFilePath := filepath.Join(configDir, configFilename)
	_, err = os.Stat(configFilePath)
	if err != nil {
		err := saveNewConfigFile(configFilePath)
		if err != nil {
			return nil, err
		}
		return nil, fmt.Errorf("config file does not exist. Created a new config file at '%s'. Update that file with the correct configuration", configFilePath)
	}
	var appConfig AppConfig
	err = readConfigFile(configFilePath, &appConfig)
	if err != nil {
		return nil, err
	}
	err = validateConfig(appConfig)
	if err != nil {
		return nil, err
	}
	return &appConfig, nil
}
