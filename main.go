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

func makeGetRequest(endpoint string) *http.Request {
	req, err := http.NewRequest("GET", endpoint, nil)
	if err != nil {
		fmt.Println("Error forming request:", err)
		return nil
	}
	req.Header.Set("Accept", "application/vnd.github+json")
	return req
}

func getResponse(req *http.Request) (*http.Response, error) {
	clt := http.DefaultClient
	resp, err := clt.Do(req)
	if resp.StatusCode == 404 {
		err = errors.New("no events found by the given username")
	}
	if err != nil {
		return nil, err
	}
	return resp, nil
}

func getRawEvents(resp *http.Response) []byte {
	defer resp.Body.Close()
	data, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Println("Error reading event data:", err)
		return nil
	}
	return data
}

func parseEvents(data []byte) []EventEnvelope {
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
	fmt.Println("Endpoint:", endpoint)

	request := makeGetRequest(endpoint)
	response, err := getResponse(request)
	if err != nil {
		fmt.Println("Error sending request:", err)
		return
	}

	eventData := getRawEvents(response)
	envelopes := parseEvents(eventData)
	// Still yet to parse the payloads properly... or at all.
	// But at least I'm no longer losing data that doesn't fit push events.

	for idx, event := range envelopes {
		fmt.Print("#", idx+1, "\n", event, "\n")
	}
}
