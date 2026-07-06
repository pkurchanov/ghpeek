package main

import (
	"fmt"
	"strings"
)

type PushPayload struct {
	// No fields needed so far.
}

func (p PushPayload) Format(env Event) string {
	return fmt.Sprintf("Pushed to %s", env.Repo.Name)
}

type CreatePayload struct {
	RefType string `json:"ref_type"`
}

func (p CreatePayload) Format(env Event) string {
	return fmt.Sprintf("Created a %s in %s", p.RefType, env.Repo.Name)
}

type WatchPayload struct {
	Action string `json:"action"`
}

func (p WatchPayload) Format(env Event) string {
	return fmt.Sprintf("Starred %s", env.Repo.Name)
}

type IssueCommentPayload struct {
	Action  string  `json:"action"`
	Issue   Issue   `json:"issue"`
	Comment Comment `json:"comment"`
}

func (p IssueCommentPayload) Format(env Event) string {
	issueType := "issue"
	// Empty struct here would mean that the issue is, in fact, an issue.
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

func (p PullRequestPayload) Format(env Event) string {
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

type IssuesPayload struct {
	Action   string `json:"action"`
	Issue    Issue  `json:"issue"`
	Assignee User   `json:"assignee"`
	Label    Label  `json:"label"`
}

func (p IssuesPayload) Format(env Event) string {
	action := p.Action
	switch action {
	case "assigned":
		return fmt.Sprintf("%s %s to issue #%d in %s",
			asciiLowerToTitle(action),
			p.Assignee.DisplayLogin,
			p.Issue.Number,
			env.Repo.Name,
		)
	case "unassigned":
		return fmt.Sprintf("%s %s from issue #%d in %s",
			asciiLowerToTitle(action),
			p.Assignee.DisplayLogin,
			p.Issue.Number,
			env.Repo.Name,
		)
	case "labeled":
		fallthrough
	case "unlabeled":
		return fmt.Sprintf("%s issue #%d as '%s' in %s",
			asciiLowerToTitle(action),
			p.Issue.Number,
			p.Label.Name,
			env.Repo.Name,
		)
	default:
		return fmt.Sprintf("%s issue #%d in %s",
			asciiLowerToTitle(action),
			p.Issue.Number,
			env.Repo.Name)
	}
}
