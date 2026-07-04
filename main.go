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

func userEventsEndpoint(username string) string {
	return "https://api.github.com/users/" + username + "/events"
}

// Construct a GET request expecting JSON back.
func makeRequest(endpoint string) (*http.Request, error) {
	req, err := http.NewRequest("GET", endpoint, nil)
	req.Header.Set("Accept", "application/vnd.github+json")
	if err != nil {
		return nil, err
	}
	return req, nil
}

// Send a given request and save the response.
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

// Assume the argument to be JSON-encoded GitHub events and parse them except for the payloads.
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
	fmt.Println("Events extracted!")
	return envs
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

	for idx, env := range envelopes {
		fmt.Printf("Event #%d:\n", idx+1)
		switch env.Type {
		case "PushEvent":
			var payload PushPayload
			if err := json.Unmarshal(env.Payload, &payload); err != nil {
				fmt.Println("Error parsing push event:", err)
				return
			}
			fmt.Println(payload)
		}
	}
}
