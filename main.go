package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"
)

/*
What I do know so far:
    - Listing public events yields one of four responses:
        - 200 (OK);
        - 304 (Not modified);
        - 403 (Forbidden);
        - 503 (Service unavailable).
    - The timeline will include up to 300 events.
    - Events older than 30 days will never be included.
    - For queries by user, the default params fetch a single page of 30 results.
        - 100 is the upper limit for per_page
*/

type Event struct {
	ID        string    `json:"id"`
	Type      string    `json:"type"`
	Actor     Actor     `json:"actor"`
	Repo      Repo      `json:"repo"`
	Payload   Payload   `json:"payload"`
	Public    bool      `json:"public"`
	CreatedAt time.Time `json:"created_at"`
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

type Payload struct {
	RepositoryID int    `json:"repository_id"`
	PushID       int64  `json:"push_id"`
	Ref          string `json:"ref"`
	Head         string `json:"head"`
	Before       string `json:"before"`
}

func userEventsEndpoint(username string) string {
	return "https://api.github.com/users/" + username + "/events"
}

func main() {
	args := os.Args[1:]
	user := args[0]

	client := http.DefaultClient
	endpoint := userEventsEndpoint(user)
	fmt.Println("Endpoint:", endpoint)

	req, err := http.NewRequest("GET", endpoint, nil)
	if err != nil {
		fmt.Println("Error forming request:", err)
		return
	}
	req.Header.Set("Accept", "application/vnd.github+json")

	resp, err := client.Do(req)
	if err != nil {
		fmt.Println("Error sending request", err)
		return
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Println("Error reading from body:", err)
		return
	}
	fmt.Println("Response status:", resp.Status)

	var events []Event
	err = json.Unmarshal(body, &events)
	if err != nil {
		fmt.Println("Error parsing JSON:", err)
		return
	}
	fmt.Printf("Events extracted!\n\n%v\n\n", events)
}
