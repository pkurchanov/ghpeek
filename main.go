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
	"time"

	"github.com/fatih/color"
)

type Event struct {
	ID            string          `json:"id"`
	Type          string          `json:"type"`
	Actor         User            `json:"actor"`
	Repo          Repo            `json:"repo"`
	Public        bool            `json:"public"`
	CreatedAt     time.Time       `json:"created_at"`
	RawPayload    json.RawMessage `json:"payload"`
	ParsedPayload Formatter
}

type Formatter interface {
	Format(env Event) string
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

func parsePayload(env *Event) error {
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
	case "IssuesEvent":
		var p IssuesPayload
		if err := json.Unmarshal(env.RawPayload, &p); err != nil {
			return err
		}
		env.ParsedPayload = p
	case "CommitCommentEvent":
		var p CommitCommentPayload
		if err := json.Unmarshal(env.RawPayload, &p); err != nil {
			return err
		}
		env.ParsedPayload = p
	case "DeleteEvent":
		var p DeletePayload
		if err := json.Unmarshal(env.RawPayload, &p); err != nil {
			return err
		}
		env.ParsedPayload = p
	case "DiscussionEvent":
		var p DiscussionPayload
		if err := json.Unmarshal(env.RawPayload, &p); err != nil {
			return err
		}
		env.ParsedPayload = p
	case "ForkEvent":
		var p ForkPayload
		if err := json.Unmarshal(env.RawPayload, &p); err != nil {
			return err
		}
		env.ParsedPayload = p
	case "GollumEvent":
		var p GollumPayload
		if err := json.Unmarshal(env.RawPayload, &p); err != nil {
			return err
		}
		env.ParsedPayload = p
	case "MemberEvent":
		var p MemberPayload
		if err := json.Unmarshal(env.RawPayload, &p); err != nil {
			return err
		}
		env.ParsedPayload = p
	case "PublicEvent":
		var p PublicPayload
		if err := json.Unmarshal(env.RawPayload, &p); err != nil {
			return err
		}
		env.ParsedPayload = p
	case "PullRequestReviewEvent":
		var p PullRequestReviewPayload
		if err := json.Unmarshal(env.RawPayload, &p); err != nil {
			return err
		}
		env.ParsedPayload = p
	case "PullRequestReviewCommentEvent":
		var p PullRequestReviewCommentPayload
		if err := json.Unmarshal(env.RawPayload, &p); err != nil {
			return err
		}
		env.ParsedPayload = p
	case "ReleaseEvent":
		var p ReleasePayload
		if err := json.Unmarshal(env.RawPayload, &p); err != nil {
			return err
		}
		env.ParsedPayload = p
	default:
		env.ParsedPayload = nil
	}
	return nil
}

// `strings.Title` is deprecated.
// `cases` is a whole external module.
// This will have to do for my current use case.
func asciiLowerToTitle(s string) string {
	if s == "" {
		return ""
	}
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
func extractEventData(req *http.Request) ([]Event, error) {
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	switch resp.StatusCode {
	case 200:
		defer resp.Body.Close()
		data, err := io.ReadAll(resp.Body)
		if err != nil {
			return nil, fmt.Errorf("couldn't read event data: %v", err)
		}
		var envs []Event
		if err := json.Unmarshal(data, &envs); err != nil {
			return nil, fmt.Errorf("couldn't parse event data: %v", err)
		}
		return envs, nil
	case 403:
		return nil, errors.New("caching coming Soon™️.")
		// etag := resp.Header.Get("etag")
	case 404:
		// We can be reasonably sure that this is what 404 means (see userEventsEndpoint).
		return nil, errors.New("no user found by the given name.")
	case 503:
		return nil, errors.New("service unavailable. Try again in a few minutes.")
	default:
		return nil, fmt.Errorf("response came back with code %d.", resp.StatusCode)
	}
}

// Inspects a given event envelope and generates a type-appropriate report.
func makeEventReport(env Event) (string, error) {
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

// Modeling a contiguous block of events of exactly the same kind.
type EventGroup map[string]int

// Pretty-printer for EventGroup elements.
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

// The final pretty-printer.
// "Pretty" as in "pretty coupled with payload parsing".
func displayAll(envs []Event) error {
	var lastReport string
	eventGroup := make(EventGroup)
	lastDate := ""
	for _, env := range envs {
		newDate := env.CreatedAt.Format(time.DateOnly)
		if newDate != lastDate {
			// Squeeze what's left out of the previous day.
			display(lastReport, eventGroup)
			lastReport = ""

			fmt.Printf("\n  %s\n", color.HiMagentaString(newDate))
			lastDate = newDate
		}
		newReport, err := makeEventReport(env)
		if err != nil {
			return fmt.Errorf("couldn't parse event payload: %v", err)
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
	return nil
}

func main() {
	args := os.Args[1:]
	user := "torvalds"
	if len(args) != 1 {
		fmt.Print("Usage: ghpeek <username>\n\nShowing activity for torvalds\n")
	} else {
		user = args[0]
	}
	endpoint := userEventsEndpoint(user)
	request, err := makeRequest(endpoint)
	if err != nil {
		fmt.Println("Request construction error:", err)
		return
	}
	envelopes, err := extractEventData(request)
	if err != nil {
		fmt.Println("Data extraction error:", err)
		return
	}
	// Push latest events to the bottom.
	slices.SortFunc(
		envelopes,
		func(a, b Event) int {
			return a.CreatedAt.Compare(b.CreatedAt)
		})
	if err := displayAll(envelopes); err != nil {
		fmt.Println("Display error:", err)
	}
}
