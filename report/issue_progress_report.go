package report

import (
	"fmt"
	"gitlabsprintreport/gitlab"
	"io"
	"slices"
	"sort"
	"strings"
	"text/tabwriter"
	"time"
)

type IssueProgressConfig struct {
	ProgressLabels []string
	DoneLabel      string
	ExcludedLabels []string
}

func NewIssueProgressConfig(progressLabels []string, doneLabel string, excludedLabels []string) IssueProgressConfig {
	return IssueProgressConfig{ProgressLabels: progressLabels, DoneLabel: doneLabel, ExcludedLabels: excludedLabels}
}

type progress struct {
	status       string
	duration     time.Duration
	isStillGoing bool
}

type issueProgress struct {
	title       string
	id          uint64
	progress    []progress
	otherLabels []string
}

type userProgress struct {
	assignee string
	issues   []issueProgress
}

type IssueProgressReport struct {
	progress []userProgress
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
	issueId := issue.Id
	result := []progress{}
	otherlabels := []string{}

	for _, status := range issue.StatusChanges {
		eventsQnt := len(status.Events)
		if slices.Contains(config.ProgressLabels, status.Label) {
			totalDuration := time.Duration(0)
			isStillGoing := false
			status.SortEvents()
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
		if !slices.Contains(config.ExcludedLabels, status.Label) &&
			!slices.Contains(config.ProgressLabels, status.Label) &&
			status.Label != config.DoneLabel &&
			eventsQnt%2 != 0 &&
			!slices.Contains(otherlabels, status.Label) {
			otherlabels = append(otherlabels, status.Label)
		}
	}

	sort.Strings(otherlabels)

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

	return issueProgress{title: title, id: issueId, progress: result, otherLabels: otherlabels}
}

func NewIssueProgressReport(issues []gitlab.Issue, config IssueProgressConfig) IssueProgressReport {

	ups := make(map[string]userProgress)
	for _, issue := range issues {
		var assignee string
		if issue.Assignee.Name != "" {
			assignee = issue.Assignee.Name
		} else {
			assignee = "Unknown"
		}
		_, contains := ups[assignee]
		if !contains {
			ups[assignee] = userProgress{assignee: assignee, issues: []issueProgress{}}
		}
		up := ups[assignee]
		ip := newIssueProgress(issue, config)
		up.issues = append(up.issues, ip)
		ups[assignee] = up
	}
	result := []userProgress{}
	for _, s := range ups {
		result = append(result, s)
	}

	return IssueProgressReport{progress: result}
}

func (r IssueProgressReport) PrintTable(stdout io.Writer) {
	w := tabwriter.NewWriter(stdout, 0, 0, 3, ' ', tabwriter.TabIndent|tabwriter.Debug)
	fmt.Fprintln(w, "Assignee\tIssue ID\tIssue Title\tStatus\tDuration\tOther Labels")
	for _, up := range r.progress {
		fmt.Fprintf(w, "%s\t \t \t \t \t \n", up.assignee)
		for _, issue := range up.issues {
			otherLabels := strings.Join(issue.otherLabels, ", ")
			fmt.Fprintf(w, " \t%d\t%s\t \t \t%s\n", issue.id, issue.title, otherLabels)
			for _, sd := range issue.progress {
				duration := formatDuration(sd.duration)
				fmt.Fprintf(w, " \t \t \t%s\t%s", sd.status, duration)
				if sd.isStillGoing {
					fmt.Fprintf(w, ", Still going")
				}
				fmt.Fprintf(w, "\t \n")
			}
		}
	}
	w.Flush()
}

func (r IssueProgressReport) PrintCsv(stdout io.Writer) {
	fmt.Fprintln(stdout, "Assignee;Issue ID;Issue Title;Status;Duration;Other Labels")
	for _, up := range r.progress {
		fmt.Fprintf(stdout, "%s;;;;;\n", up.assignee)
		for _, issue := range up.issues {
			otherLabels := strings.Join(issue.otherLabels, ", ")
			fmt.Fprintf(stdout, ";%d;%s;;;%s\n", issue.id, issue.title, otherLabels)
			for _, sd := range issue.progress {
				duration := formatDuration(sd.duration)
				fmt.Fprintf(stdout, ";;;%s;%s", sd.status, duration)
				if sd.isStillGoing {
					fmt.Fprintf(stdout, ", Still going")
				}
				fmt.Fprintf(stdout, ";\n")
			}
		}
	}
}
