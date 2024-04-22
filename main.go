package main

import (
	"fmt"
	"gitlabsprintreport/gitlab"
	"gitlabsprintreport/report"
	"net/http"
	"os"
	"path/filepath"
)

func main() {

	httpClient := &http.Client{}
	fmt.Println("Starting...")
	config, err := loadConfig()
	if err != nil {
		fmt.Printf("Error: %s\n", err.Error())
		os.Exit(2)
	}
	gitLabConfig := gitlab.NewConfig(
		config.GitLabConfig.URL,
		config.GitLabConfig.PersonalToken,
		config.GitLabConfig.ProjectId,
		config.GitLabConfig.TrackLabels,
		config.GitLabConfig.TrackIssueTypes,
	)
	sprintLabel := "Sprint-5"
	issues, err := gitlab.IssuesFromSprint(httpClient, gitLabConfig, sprintLabel)
	if err != nil {
		fmt.Printf("Error: %s\n", err.Error())
		os.Exit(2)
	}
	fmt.Printf("Found %d issues\n", len(issues))
	reportConfig := report.NewIssueProgressConfig([]string{"To do", "In Progress", "PR", "QA"}, "Done", []string{sprintLabel})
	fmt.Println("Generating Issue progress report...")
	issueProgressReport := report.NewIssueProgressReport(issues, reportConfig)
	path, err := os.Getwd()
	if err != nil {
		fmt.Printf("Error: %s\n", err.Error())
		os.Exit(2)
	}
	filename := fmt.Sprintf("%s-report.txt", sprintLabel)
	outputFile := filepath.Join(path, filename)
	file, err := os.Create(outputFile)
	if err != nil {
		fmt.Printf("Error: %s\n", err.Error())
		os.Exit(2)
	}
	issueProgressReport.PrintTable(file)
	file.Close()
	fmt.Printf("File generated at: %s", outputFile)
}
