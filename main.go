package main

import (
	"fmt"
	"gitlabsprintreport/gitlab"
	"gitlabsprintreport/report"
	"net/http"
	"os"
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
	issues, err := gitlab.IssuesFromSprint(httpClient, gitLabConfig, "Sprint-1")
	if err != nil {
		fmt.Printf("Error: %s\n", err.Error())
		os.Exit(2)
	}
	fmt.Printf("FOUND %d issues\n", len(issues))
	reportConfig := report.NewIssueProgressConfig([]string{"To do", "In Progress", "PR", "QA"})
	fmt.Println("Generating Issue progress report...")
	issueProgressReport := report.NewIssueProgressReport(issues, reportConfig)
	issueProgressReport.PrintTable(os.Stdout)

}
