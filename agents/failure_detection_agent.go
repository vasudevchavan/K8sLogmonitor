package agents

import (
	"regexp"
)

type FailureDetectionAgent struct {
	patterns []*regexp.Regexp
}

func NewFailureDetectionAgent() *FailureDetectionAgent {
	patternStrings := []string{
		"panic:", "error:", "failed to .*", "connection refused",
		"pull image", "startup error", "waiting to start", "imagepullbackoff",
	}
	patterns := make([]*regexp.Regexp, len(patternStrings))
	for i, p := range patternStrings {
		patterns[i] = regexp.MustCompile("(?i)" + p)
	}
	return &FailureDetectionAgent{patterns: patterns}
}

func (a *FailureDetectionAgent) DetectFailures(logs string) []string {
	var failures []string
	for _, re := range a.patterns {
		matches := re.FindAllString(logs, -1)
		failures = append(failures, matches...)
	}
	return failures
}
