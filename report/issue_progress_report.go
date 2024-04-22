package report

import (
	"fmt"
	"gitlabsprintreport/gitlab"
	"io"
	"slices"
	"text/tabwriter"
	"time"
)

type IssueProgressConfig struct {
	ProgressLabels []string
}

func NewIssueProgressConfig(progressLabels []string) IssueProgressConfig {
	return IssueProgressConfig{ProgressLabels: progressLabels}
}

type progress struct {
	status       string
	duration     time.Duration
	isStillGoing bool
}

type issueProgress struct {
	title    string
	progress []progress
}

type IssueProgressReport struct {
	issues []issueProgress
}

func formatDuration(d time.Duration) string {
	duration := d
	day := time.Hour * 24
	days := duration / day
	duration -= day * days
	hours := duration / time.Hour
	duration -= time.Hour * hours
	minutes := duration / time.Minute
	duration -= time.Minute * minutes
	seconds := duration / time.Second
	duration -= time.Second * seconds
	return fmt.Sprintf("%d day(s) %02d:%02d:%02d", days.Nanoseconds(), hours.Nanoseconds(), minutes.Nanoseconds(), seconds.Nanoseconds())
}

func durationBetweenWorkDates(start time.Time, end time.Time) time.Duration {
	nonWorkingDuration := time.Duration(0)
	oneDay, _ := time.ParseDuration("24h")
	for i := start; i.Before(end); i = i.Add(oneDay) {
		switch i.Weekday() {
		case time.Saturday:
			nonWorkingDuration = nonWorkingDuration + oneDay
		case time.Sunday:
			nonWorkingDuration = nonWorkingDuration + oneDay
		}
	}
	duration := end.Sub(start)
	totalDuration := duration - nonWorkingDuration
	return totalDuration
}

func newIssueProgress(issue gitlab.Issue, config IssueProgressConfig) issueProgress {
	title := issue.Title
	result := []progress{}
	for _, status := range issue.StatusChanges {
		if slices.Contains(config.ProgressLabels, status.Label) {
			totalDuration := time.Duration(0)
			isStillGoing := false
			status.SortEvents()
			eventsQnt := len(status.Events)
			if eventsQnt%2 != 0 {
				event := status.Events[eventsQnt-1]
				now := time.Now() //Need to receive as input to test properly
				duration := durationBetweenWorkDates(event.Timestamp, now)
				isStillGoing = true
				totalDuration = totalDuration + duration
			}
			if eventsQnt >= 2 {
				for i := 1; i < eventsQnt; i = i + 2 {
					start := status.Events[i-1]
					end := status.Events[i]
					duration := durationBetweenWorkDates(start.Timestamp, end.Timestamp)
					totalDuration = totalDuration + duration
				}
			}
			result = append(result, progress{status: status.Label, duration: totalDuration, isStillGoing: isStillGoing})
		}
	}
	slices.SortFunc(result, func(i, j progress) int {
		ii := slices.Index(config.ProgressLabels, i.status)
		ij := slices.Index(config.ProgressLabels, j.status)
		switch {
		case ii > ij:
			return 1
		case ii < ij:
			return -1
		}
		return 0
	})

	return issueProgress{title: title, progress: result}
}

func NewIssueProgressReport(issues []gitlab.Issue, config IssueProgressConfig) IssueProgressReport {
	result := []issueProgress{}
	for _, issue := range issues {
		ip := newIssueProgress(issue, config)
		result = append(result, ip)
	}
	return IssueProgressReport{issues: result}
}

func (r IssueProgressReport) PrintTable(stdout io.Writer) {
	w := tabwriter.NewWriter(stdout, 0, 0, 3, ' ', tabwriter.TabIndent|tabwriter.Debug)
	fmt.Fprintln(w, "Issue Title\tStatus\tDuration")
	for _, issue := range r.issues {
		fmt.Fprintf(w, "%s\t\t\n", issue.title)
		for _, sd := range issue.progress {
			duration := formatDuration(sd.duration)
			fmt.Fprintf(w, " \t%s\t%s", sd.status, duration)
			if sd.isStillGoing {
				fmt.Fprintf(w, ", Still going")
			}
			fmt.Fprintf(w, "\n")
		}
	}
	w.Flush()
}
