package report

type StoryPointProgressConfig struct {
	DoneLabel string
}

func NewConfig(doneLabel string, progressLabels []string) StoryPointProgressConfig {
	return StoryPointProgressConfig{DoneLabel: doneLabel}
}

/*
Timestamp 							| Story Points 	| Remaining Story Points
2023-01-01 11:33:25			| 3	 						| 8
2023-01-01 11:33:25			| 5	 						| 3
*/
