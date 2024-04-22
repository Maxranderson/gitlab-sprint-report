package gitlab

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"slices"
	"strconv"
	"strings"
	"time"
)

type Label struct {
	Name string
}

type ResourceLabelEvent struct {
	CreatedAt time.Time `json:"created_at"`
	Label     Label     `json:"label"`
	Action    string    `json:"action"`
}

type EventType uint8

const (
	Add EventType = iota
	Remove
)

func (s EventType) String() string {
	switch s {
	case Add:
		return "add"
	case Remove:
		return "remove"
	}
	return "unknown"
}

type StatusEvent struct {
	Timestamp time.Time
	Type      EventType
}

type Status struct {
	Label  string
	Events []StatusEvent
}

func (status *Status) SortEvents() {
	slices.SortFunc(status.Events, func(a, b StatusEvent) int {
		return a.Timestamp.Compare(b.Timestamp)
	})
}

type User struct {
	Name string `json:"name"`
}

type Issue struct {
	Id            uint64 `json:"iid"`
	Title         string `json:"title"`
	Assignee      User
	StatusChanges []Status
}

type issuesPage struct {
	issues      []Issue
	currentPage uint64
	lastPage    uint64
}

type Config struct {
	URL             string
	PersonalToken   string
	ProjectId       uint64
	TrackLabels     []string
	TrackIssueTypes []string
}

func NewConfig(url string, personalToken string, projectId uint64, trackLabels []string, trackIssueTypes []string) Config {
	return Config{url, personalToken, projectId, trackLabels, trackIssueTypes}
}

func newGitLabRequest(config Config, path string) (*http.Request, error) {
	url := fmt.Sprintf("%s/api/v4/%s", config.URL, path)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Add("PRIVATE-TOKEN", config.PersonalToken)
	return req, nil
}

func doRequest(client *http.Client, req *http.Request) (*http.Response, []byte, error) {
	res, err := client.Do(req)
	if err != nil {
		return nil, nil, err
	}
	defer res.Body.Close()
	bodyContent, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, nil, err
	}
	if res.StatusCode == 401 {
		return nil, nil, fmt.Errorf("Unauthorized error! Please check your Personal Token.")
	}
	if res.StatusCode == 403 {
		return nil, nil, fmt.Errorf("Forbidden error! Please check your Personal Token permissions. It needs to have read_api.")
	}
	if res.StatusCode >= 400 {
		bodyError := string(bodyContent)
		return nil, nil, fmt.Errorf("Unknown error! Request code: %d. Request body: %s.", res.StatusCode, bodyError)
	}
	return res, bodyContent, nil
}

func fetchStatusChange(client *http.Client, config Config, issues []Issue) ([]Issue, error) {
	withStatusChange := []Issue{}
	for _, i := range issues {

		req, err := newGitLabRequest(config, fmt.Sprintf("projects/%d/issues/%d/resource_label_events", config.ProjectId, i.Id))
		if err != nil {
			return nil, err
		}
		_, body, err := doRequest(client, req)
		if err != nil {
			return nil, err
		}

		var resourceLabelEvents []ResourceLabelEvent
		err = json.Unmarshal(body, &resourceLabelEvents)
		if err != nil {
			return nil, err
		}
		statuss := make(map[string]Status)
		for _, e := range resourceLabelEvents {
			label := e.Label.Name
			_, contains := statuss[label]
			if !contains {
				statuss[label] = Status{Label: label}
			}

			status := statuss[label]
			var eventType EventType
			if e.Action == "add" {
				eventType = Add
			} else {
				eventType = Remove
			}
			status.Events = append(status.Events, StatusEvent{Timestamp: e.CreatedAt, Type: eventType})
			statuss[label] = status
		}
		changes := []Status{}
		for _, s := range statuss {
			changes = append(changes, s)
		}
		i.StatusChanges = changes
		withStatusChange = append(withStatusChange, i)
	}
	return withStatusChange, nil
}

func fetchIssuesByPage(client *http.Client, request http.Request, currentPage uint64) (*issuesPage, error) {
	query := request.URL.Query()
	query.Add("per_page", "50")
	query.Add("page", strconv.FormatUint(currentPage, 10))
	request.URL.RawQuery = query.Encode()

	res, body, err := doRequest(client, &request)
	if err != nil {
		return nil, err
	}

	totalPages, err := strconv.Atoi(res.Header.Get("X-Total-Pages"))
	if err != nil {
		return nil, err
	}

	var issues []Issue
	err = json.Unmarshal(body, &issues)
	if err != nil {
		return nil, err
	}

	return &issuesPage{currentPage: currentPage, lastPage: uint64(totalPages), issues: issues}, nil
}

func fetchIssues(client *http.Client, config Config, issueType string, labels string) ([]Issue, error) {
	req, err := newGitLabRequest(config, fmt.Sprintf("projects/%d/issues", config.ProjectId))
	if err != nil {
		return nil, err
	}
	query := req.URL.Query()
	query.Add("issue_type", issueType)
	query.Add("labels", labels)
	req.URL.RawQuery = query.Encode()

	resp, err := fetchIssuesByPage(client, *req, 1)
	if err != nil {
		return nil, err
	}

	var issues = resp.issues
	if resp.lastPage != 1 {
		for pn := 2; pn <= int(resp.lastPage); pn++ {
			page, err := fetchIssuesByPage(client, *req, uint64(pn))
			if err != nil {
				return nil, err
			}
			issues = append(issues, page.issues...)
		}
	}

	withStatusChange, err := fetchStatusChange(client, config, issues)
	if err != nil {
		return nil, err
	}
	return withStatusChange, nil
}

func IssuesFromSprint(client *http.Client, config Config, sprintLabel string) ([]Issue, error) {
	fmt.Printf("Fetching issues from: %s\n", config.URL)
	var issues = []Issue{}
	for _, l := range config.TrackLabels {
		ls := strings.Join([]string{l, sprintLabel}, ",")
		for _, it := range config.TrackIssueTypes {
			fetchedIssues, err := fetchIssues(client, config, it, ls)
			if err != nil {
				return nil, err
			}
			issues = append(issues, fetchedIssues...)
		}
	}

	return issues, nil
}
