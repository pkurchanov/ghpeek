package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"
)

type EventEnvelope struct {
	ID        string          `json:"id"`
	Type      string          `json:"type"`
	Actor     Actor           `json:"actor"`
	Repo      Repo            `json:"repo"`
	Public    bool            `json:"public"`
	CreatedAt time.Time       `json:"created_at"`
	Payload   json.RawMessage `json:"payload"`
}

type Actor struct {
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
	RepositoryID int    `json:"repository_id"`
	PushID       int    `json:"push_id"`
	Ref          string `json:"ref"`
	Head         string `json:"head"`
	Before       string `json:"before"`
}

type CreatePayload struct {
	Ref          string `json:"ref"`
	RefType      string `json:"ref_type"`
	FullRef      string `json:"full_ref"`
	MasterBranch string `json:"master_branch"`
	Description  string `json:"description"`
	PusherType   string `json:"pusher_type"`
}

type WatchPayload struct {
	Action string `json:"action"`
}

type IssueCommentPayload struct {
	Action  string  `json:"action"`
	Issue   Issue   `json:"issue"`
	Comment Comment `json:"comment"`
}

type Issue struct {
	Number int `json:"number"`
}

type Comment struct {
	Body string `json:"body"`
}

func userEventsEndpoint(username string) string {
	return "https://api.github.com/users/" + username + "/events"
}

// Constructs a GET request for the given endpoint.
// Politely asks the server to spit JSON back.
func makeRequest(endpoint string) (*http.Request, error) {
	req, err := http.NewRequest("GET", endpoint, nil)
	req.Header.Set("Accept", "application/vnd.github+json")
	if err != nil {
		return nil, err
	}
	return req, nil
}

// Sends a given request and saves the response.
//
// 404 in current usage strictly means there's no such user
// (see userEventsEndpoint).
func getResponse(req *http.Request) (*http.Response, error) {
	resp, err := http.DefaultClient.Do(req)
	if resp.StatusCode == 404 {
		err = errors.New("no events found by the given username")
	}
	if err != nil {
		return nil, err
	}
	return resp, nil
}

// Assumes the argument to be JSON-encoded GitHub events.
// Then parses them except for the payloads.
func extractEventEnvelopes(resp *http.Response) []EventEnvelope {
	defer resp.Body.Close()
	data, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Println("Error reading event data:", err)
		return nil
	}
	var envs []EventEnvelope
	if err := json.Unmarshal(data, &envs); err != nil {
		fmt.Println("Error parsing event data:", err)
		return nil
	}
	return envs
}

// Now this is where the payloads are parsed on demand.
func makeEventReport(env EventEnvelope) (string, error) {
	switch env.Type {
	case "PushEvent":
		var payload PushPayload
		if err := json.Unmarshal(env.Payload, &payload); err != nil {
			return "", err
		}

		return fmt.Sprintf("Commit(s) pushed to %s", env.Repo.Name), nil
	case "CreateEvent":
		var payload CreatePayload
		if err := json.Unmarshal(env.Payload, &payload); err != nil {
			return "", err
		}

		return fmt.Sprintf("A %s created on %s", payload.RefType, env.Repo.Name), nil
	case "WatchEvent":
		var payload WatchPayload
		if err := json.Unmarshal(env.Payload, &payload); err != nil {
			return "", err
		}

		un := ""
		if payload.Action != "started" {
			// Not really implemented in the API...
			// But doesn't really hurt to have, right?
			un = "un"
		}
		return fmt.Sprintf("Repo %s"+"starred: %s", un, env.Repo.Name), nil
	case "IssueCommentEvent":
		var payload IssueCommentPayload
		if err := json.Unmarshal(env.Payload, &payload); err != nil {
			return "", err
		}

		return fmt.Sprintf("Comment %s on issue #%d in %s:\n\"%s\"",
			payload.Action,
			payload.Issue.Number,
			env.Repo.Name,
			payload.Comment.Body), nil
	default:
		return fmt.Sprintf("Event type not yet implemented: %s", env.Type), nil
	}
}

func main() {
	args := os.Args[1:]
	user := args[0]

	endpoint := userEventsEndpoint(user)
	request, err := makeRequest(endpoint)
	if err != nil {
		fmt.Println("Error forming request:", err)
		return
	}
	response, err := getResponse(request)
	if err != nil {
		fmt.Println("Error sending request:", err)
		return
	}
	envelopes := extractEventEnvelopes(response)

	for _, env := range envelopes {
		fmt.Printf("\n%s\n", env.CreatedAt.Format(time.RFC1123))
		eventReport, err := makeEventReport(env)
		if err != nil {
			fmt.Println("Error parsing event payload:", err)
		}
		fmt.Println(eventReport)
	}
}
