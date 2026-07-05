package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"slices"
	"strconv"
	"strings"
	"time"
)

type EventEnvelope struct {
	ID            string          `json:"id"`
	Type          string          `json:"type"`
	Actor         User            `json:"actor"`
	Repo          Repo            `json:"repo"`
	Public        bool            `json:"public"`
	CreatedAt     time.Time       `json:"created_at"`
	RawPayload    json.RawMessage `json:"payload"`
	ParsedPayload EventFormatter
}

type EventFormatter interface {
	Format(env EventEnvelope) string
}

type User struct {
	ID           int    `json:"id"`
	Login        string `json:"login"`
	DisplayLogin string `json:"display_login"`
	GravatarID   string `json:"gravatar_id"`
	URL          string `json:"url"`
	AvatarURL    string `json:"avatar_url"`
}

type Repo struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
	URL  string `json:"url"`
}

type PushPayload struct {
	// No fields needed so far.
}

func (p PushPayload) Format(env EventEnvelope) string {
	return fmt.Sprintf("Pushed to %s", env.Repo.Name)
}

type CreatePayload struct {
	RefType string `json:"ref_type"`
}

func (p CreatePayload) Format(env EventEnvelope) string {
	return fmt.Sprintf("Created a %s in %s", p.RefType, env.Repo.Name)
}

type WatchPayload struct {
	Action string `json:"action"`
}

func (p WatchPayload) Format(env EventEnvelope) string {
	return fmt.Sprintf("Starred %s", env.Repo.Name)
}

type IssueCommentPayload struct {
	Action  string  `json:"action"`
	Issue   Issue   `json:"issue"`
	Comment Comment `json:"comment"`
}

func (p IssueCommentPayload) Format(env EventEnvelope) string {
	issueType := "issue"
	// Empty struct here would mean that there was nothing
	// under the pull_request key, confirming that
	// the issue in question is, in fact, an issue.
	if p.Issue.PullRequest != (PR{}) {
		issueType = "PR"
	}
	return fmt.Sprintf("%s a comment on %s #%d in %s:\n%s",
		asciiLowerToTitle(p.Action), issueType, p.Issue.Number, env.Repo.Name, quote(p.Comment.Body))
}

func quote(s string) string {
	lines := strings.Split(s, "\n")
	for i, line := range lines {
		lines[i] = "  > " + line
	}
	return strings.Join(lines, "\n")
}

type Issue struct {
	Number      int `json:"number"`
	PullRequest PR  `json:"pull_request"`
}

type PR struct {
	URL string `json:"html_url"`
}

type Comment struct {
	Body string `json:"body"`
}

type PullRequestPayload struct {
	Action   string `json:"action"`
	Number   int    `json:"number"`
	Assignee User   `json:"assignee,omitempty"`
	Label    Label  `json:"label,omitempty"`
}

func (p PullRequestPayload) Format(env EventEnvelope) string {
	action := p.Action
	switch action {
	case "assigned":
		return fmt.Sprintf("%s %s to PR #%d in %s",
			asciiLowerToTitle(action),
			p.Assignee.DisplayLogin,
			p.Number,
			env.Repo.Name,
		)
	case "unassigned":
		return fmt.Sprintf("%s %s from PR #%d in %s",
			asciiLowerToTitle(action),
			p.Assignee.DisplayLogin,
			p.Number,
			env.Repo.Name,
		)
	case "labeled":
		fallthrough
	case "unlabeled":
		return fmt.Sprintf("%s PR #%d as '%s' in %s",
			asciiLowerToTitle(action),
			p.Number,
			p.Label.Name,
			env.Repo.Name,
		)
	default:
		return fmt.Sprintf("%s PR #%d in %s",
			asciiLowerToTitle(action),
			p.Number,
			env.Repo.Name)
	}
}

type Label struct {
	Name string `json:"name"`
}

func parsePayload(env *EventEnvelope) error {
	switch env.Type {
	case "PushEvent":
		var p PushPayload
		if err := json.Unmarshal(env.RawPayload, &p); err != nil {
			return err
		}
		env.ParsedPayload = p
	case "CreateEvent":
		var p CreatePayload
		if err := json.Unmarshal(env.RawPayload, &p); err != nil {
			return err
		}
		env.ParsedPayload = p
	case "WatchEvent":
		var p WatchPayload
		if err := json.Unmarshal(env.RawPayload, &p); err != nil {
			return err
		}
		env.ParsedPayload = p
	case "IssueCommentEvent":
		var p IssueCommentPayload
		if err := json.Unmarshal(env.RawPayload, &p); err != nil {
			return err
		}
		env.ParsedPayload = p
	case "PullRequestEvent":
		var p PullRequestPayload
		if err := json.Unmarshal(env.RawPayload, &p); err != nil {
			return err
		}
		env.ParsedPayload = p
	default:
		env.ParsedPayload = nil
	}
	return nil
}

// Modeling a contiguous block of events of exactly the same kind.
type EventGroup map[string]int

// `strings.Title` is deprecated.
// `cases` is a whole external module.
// This will have to do for my current use case.
func asciiLowerToTitle(s string) string {
	return string(s[0]+'A'-'a') + s[1:]
}

func userEventsEndpoint(username string) string {
	return "https://api.github.com/users/" + username + "/events"
}

func makeRequest(endpoint string) (*http.Request, error) {
	req, err := http.NewRequest("GET", endpoint, nil)
	req.Header.Set("Accept", "application/vnd.github+json")
	if err != nil {
		return nil, err
	}
	return req, nil
}

// Gets us as far as medium-rare envelopes.
func extractEventData(req *http.Request) ([]EventEnvelope, error) {
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode == 404 {
		// We can be reasonably sure that this is what 404 means (see userEventsEndpoint).
		return nil, errors.New("no user found by the given name.")
	}

	// Caching coming Soon™️
	// etag := resp.Header.Get("etag")

	defer resp.Body.Close()
	data, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Println("Error reading event data:", err)
		return nil, err
	}
	var envs []EventEnvelope
	if err := json.Unmarshal(data, &envs); err != nil {
		fmt.Println("Error parsing event data:", err)
		return nil, err
	}
	return envs, nil
}

// Inspects a given event envelope and generates a type-appropriate report.
//
// In the future it might be nice to have an intermediate consolidation step.
func makeEventReport(env EventEnvelope) (string, error) {
	if err := parsePayload(&env); err != nil {
		return "", err
	}
	var report string
	if env.ParsedPayload == nil {
		report = fmt.Sprintf("Event type not yet implemented: %s", env.Type)
		return report, nil
	}
	report = env.ParsedPayload.Format(env)
	return report, nil
}

func display(rep string, eg EventGroup) {
	if rep != "" {
		count := eg[rep]
		countLabel := ""
		if count > 1 {
			countLabel = "x" + strconv.Itoa(count)
		}
		fmt.Println("-", rep, countLabel)
	}
}

func displayAll(envs []EventEnvelope) {
	var lastReport string
	eventGroup := make(EventGroup)
	lastDate := ""
	for _, env := range envs {
		newDate := env.CreatedAt.Format(time.DateOnly)
		if newDate != lastDate {
			// Stop groups from migrating across day boundaries.
			display(lastReport, eventGroup)
			fmt.Printf("\n  %s\n", newDate)
			lastDate = newDate
		}
		newReport, err := makeEventReport(env)
		if err != nil {
			fmt.Println("Error parsing event payload:", err)
		}
		if lastReport != newReport {
			display(lastReport, eventGroup)
			lastReport = newReport
			eventGroup[newReport] = 1
		} else {
			_, ok := eventGroup[newReport]
			if ok {
				eventGroup[newReport] += 1
			}
		}
	}
	// Keep the last group from evaporating.
	display(lastReport, eventGroup)
}

func main() {
	args := os.Args[1:]
	if len(args) != 1 {
		fmt.Println("Usage: ghpeek <username>")
		return
	} else {
		user := args[0]

		endpoint := userEventsEndpoint(user)
		request, err := makeRequest(endpoint)
		if err != nil {
			fmt.Println("Error forming request:", err)
			return
		}
		envelopes, err := extractEventData(request)
		if err != nil {
			fmt.Println("Error extracting data:", err)
			return
		}
		// Latest events at the bottom.
		slices.SortFunc(envelopes, func(a, b EventEnvelope) int { return a.CreatedAt.Compare(b.CreatedAt) })

		displayAll(envelopes)
	}
}
