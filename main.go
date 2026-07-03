package main

import (
	"fmt"
	"os"
)

/*
What I do know so far:
- Listing public events returns either 200, 304, 403 or 503
- Events older than 30 days will strictly not be included
- The default query params give you a single page of 15 results
*/

func userEventsEndpoint(username string) string {
	return fmt.Sprintf("https://api.github.com/users/%s/events", username)
}

func main() {
	args := os.Args[1:]
	user := args[0]

	fmt.Println(userEventsEndpoint(user))
}
