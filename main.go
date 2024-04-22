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
	gitLabConfig := gitlab.NewConfig("**", "**", 0)
	issues, err := gitlab.IssuesFromSprint(httpClient, gitLabConfig, "Sprint-4")
	if err != nil {
		panic(err)
	}
	fmt.Printf("FOUND %+v\n", issues)
	config := report.NewIssueProgressConfig([]string{"To do", "In Progress", "PR", "QA"})
	issueProgressReport := report.NewIssueProgressReport(issues, config)
	issueProgressReport.PrintTable(os.Stdout)

}
