package main

import (
	"fmt"
	"io"
	"net/http"
	"os"
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
    - What you get back is an array of objects like this:
        {
            "id": "22249084964",
            "type": "PushEvent",
            "actor": {
                "id": 583231,
                "login": "octocat",
                "display_login": "octocat",
                "gravatar_id": "",
                "url": "https://api.github.com/users/octocat",
                "avatar_url": "https://avatars.githubusercontent.com/u/583231?v=4"
            },
            "repo": {
                "id": 1296269,
                "name": "octocat/Hello-World",
                "url": "https://api.github.com/repos/octocat/Hello-World"
            },
            "payload": {
                "repository_id": 1296269,
                "push_id": 10115855396,
                "ref": "refs/heads/master",
                "head": "7a8f3ac80e2ad2f6842cb86f576d4bfe2c03e300",
                "before": "883efe034920928c47fe18598c01249d1a9fdabd"
            },
            "public": true,
            "created_at": "2022-06-09T12:47:28Z"
        }
*/

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
	fmt.Println("Response status:", resp.Status)
	fmt.Printf("Response body:\n%s\n", body)
}
