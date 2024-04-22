package report

import (
	"gitlabsprintreport/gitlab"
	"testing"
	"time"
)

/*
	{
		Id:359 Title:Teste Unitário totalizador (UTIL)
		StatusChanges:[
			{
				Label:E-2
				Events:[{Timestamp:2024-02-22 10:28:35.767 -0300 -03 Type:add}]
			}
			{
				Label:In Progress
				Events:[{Timestamp:2024-02-22 14:28:25.749 -0300 -03 Type:add} {Timestamp:2024-02-27 14:16:38.356 -0300 -03 Type:remove}]
			}
			{
				Label:Sprint-1
				Events:[{Timestamp:2024-02-23 15:07:20.435 -0300 -03 Type:add}]
			}
			{
				Label:PR
				Events:[{Timestamp:2024-02-26 12:58:44.447 -0300 -03 Type:add} {Timestamp:2024-02-27 14:16:43.463 -0300 -03 Type:remove}]
			}
			{
				Label:Dev MOD
				Events:[{Timestamp:2024-02-26 15:10:41.18 -0300 -03 Type:add}]}
			{
				Label:QA
				Events:[{Timestamp:2024-02-27 14:16:38.356 -0300 -03 Type:add}]}
			{
				Label:To do
				Events:[{Timestamp:2024-02-19 13:04:01.029 -0300 -03 Type:add} {Timestamp:2024-02-22 10:09:20.66 -0300 -03 Type:remove} {Timestamp:2024-02-22 10:29:10.786 -0300 -03 Type:add} {Timestamp:2024-02-22 14:28:25.749 -0300 -03 Type:remove}]} {Label:In Review Events:[{Timestamp:2024-02-22 10:08:55.525 -0300 -03 Type:add} {Timestamp:2024-02-22 10:29:10.786 -0300 -03 Type:remove}]
			}
		]
	}
*/
func TestStatusDuration(t *testing.T) {
	statusChange := []gitlab.Status{
		{
			Label: "In Progress",
			Events: []gitlab.StatusEvent{
				{Timestamp: time.Date(2024, time.February, 22, 14, 0, 0, 0, time.UTC), Type: gitlab.Add},
				{Timestamp: time.Date(2024, time.February, 27, 14, 0, 0, 0, time.UTC), Type: gitlab.Remove},
			},
		},
	}
	issue := gitlab.Issue{Id: 359, Title: "Teste Unitário totalizador (UTIL)", StatusChanges: statusChange}
	sds := newIssueProgress(issue)
	expected := StatusDuration{status: "In Progress", duration: time.Duration(259200000000000), isStillGoing: false}
	if sds != expected {
		t.Fatalf("Value %+v is not equals to %+v", sds[0], expected)
	}
}
