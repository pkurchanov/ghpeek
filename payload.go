package main

import (
	"fmt"
	"strings"

	"github.com/fatih/color"
)

type PushPayload struct {
	Ref string `json:"ref"`
}

func (p PushPayload) Format(env Event) string {
	refName := p.Ref
	if parts := strings.Split(refName, "/"); len(parts) >= 3 {
		refName = parts[len(parts)-1]
	}
	return fmt.Sprintf("%s to %s in %s", color.BlueString("Pushed"), refName, env.Repo.Name)
}

type CreatePayload struct {
	RefType string `json:"ref_type"`
}

func (p CreatePayload) Format(env Event) string {
	return fmt.Sprintf("%s a %s in %s", color.GreenString("Created"), p.RefType, env.Repo.Name)
}

type WatchPayload struct {
	Action string `json:"action"`
}

func (p WatchPayload) Format(env Event) string {
	return fmt.Sprintf("%s %s", color.YellowString("Starred"), env.Repo.Name)
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
		color.CyanString(asciiLowerToTitle(p.Action)), issueType, p.Issue.Number, env.Repo.Name, quote(p.Comment.Body))
}

func quote(s string) string {
	s = strings.TrimSpace(s)
	style := color.New(color.Faint, color.Italic).SprintFunc()
	prefix := color.New(color.FgHiBlack).Sprint("  │ ")

	lines := strings.Split(s, "\n")
	for i, line := range lines {
		lines[i] = prefix + style(line)
	}
	return strings.Join(lines, "\n")
}

type Issue struct {
	Number      int `json:"number"`
	PullRequest PR  `json:"pull_request"`
}

type PR struct {
	Number int    `json:"number"`
	URL    string `json:"html_url"`
}

type Comment struct {
	Body string `json:"body"`
}

type PullRequestPayload struct {
	Action   string `json:"action"`
	Number   int    `json:"number"`
	Assignee User   `json:"assignee"`
	Label    Label  `json:"label"`
}

func (p PullRequestPayload) Format(env Event) string {
	coloredAction := color.YellowString(asciiLowerToTitle(p.Action))
	switch p.Action {
	case "assigned":
		return fmt.Sprintf("%s %s to PR #%d in %s",
			coloredAction,
			p.Assignee.DisplayLogin,
			p.Number,
			env.Repo.Name,
		)
	case "unassigned":
		return fmt.Sprintf("%s %s from PR #%d in %s",
			coloredAction,
			p.Assignee.DisplayLogin,
			p.Number,
			env.Repo.Name,
		)
	case "labeled":
		fallthrough
	case "unlabeled":
		return fmt.Sprintf("%s PR #%d as '%s' in %s",
			coloredAction,
			p.Number,
			p.Label.Name,
			env.Repo.Name,
		)
	default:
		return fmt.Sprintf("%s PR #%d in %s",
			coloredAction,
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
	coloredAction := color.YellowString(asciiLowerToTitle(p.Action))
	switch p.Action {
	case "assigned":
		return fmt.Sprintf("%s %s to issue #%d in %s",
			coloredAction,
			p.Assignee.DisplayLogin,
			p.Issue.Number,
			env.Repo.Name,
		)
	case "unassigned":
		return fmt.Sprintf("%s %s from issue #%d in %s",
			coloredAction,
			p.Assignee.DisplayLogin,
			p.Issue.Number,
			env.Repo.Name,
		)
	case "labeled":
		fallthrough
	case "unlabeled":
		return fmt.Sprintf("%s issue #%d as '%s' in %s",
			coloredAction,
			p.Issue.Number,
			p.Label.Name,
			env.Repo.Name,
		)
	default:
		return fmt.Sprintf("%s issue #%d in %s",
			coloredAction,
			p.Issue.Number,
			env.Repo.Name)
	}
}

type CommitCommentPayload struct {
	Action  string  `json:"action"`
	Comment Comment `json:"comment"`
}

func (p CommitCommentPayload) Format(env Event) string {
	return fmt.Sprintf("%s on a commit in %s:\n%s",
		color.CyanString("Commented"), env.Repo.Name, quote(p.Comment.Body))
}

type DeletePayload struct {
	Ref     string `json:"ref"`
	RefType string `json:"ref_type"`
}

func (p DeletePayload) Format(env Event) string {
	return fmt.Sprintf("%s %s %s in %s", color.RedString("Deleted"), p.RefType, p.Ref, env.Repo.Name)
}

type DiscussionPayload struct {
	Action     string     `json:"action"`
	Discussion Discussion `json:"discussion"`
}

type Discussion struct {
	Number int    `json:"number"`
	Title  string `json:"title"`
}

func (p DiscussionPayload) Format(env Event) string {
	return fmt.Sprintf("%s discussion #%d: %s in %s",
		color.GreenString(asciiLowerToTitle(p.Action)), p.Discussion.Number, p.Discussion.Title, env.Repo.Name)
}

type ForkPayload struct {
	Action string     `json:"action"`
	Forkee Repository `json:"forkee"`
}

type Repository struct {
	FullName string `json:"full_name"`
}

func (p ForkPayload) Format(env Event) string {
	return fmt.Sprintf("%s %s to %s", color.MagentaString("Forked"), env.Repo.Name, p.Forkee.FullName)
}

type GollumPayload struct {
	Pages []WikiPage `json:"pages"`
}

type WikiPage struct {
	PageName string `json:"page_name"`
	Action   string `json:"action"`
}

func (p GollumPayload) Format(env Event) string {
	if len(p.Pages) == 0 {
		return fmt.Sprintf("%s wiki in %s", color.BlueString("Updated"), env.Repo.Name)
	}
	page := p.Pages[0]
	return fmt.Sprintf("%s wiki page '%s' in %s", color.BlueString(asciiLowerToTitle(page.Action)), page.PageName, env.Repo.Name)
}

type MemberPayload struct {
	Action string `json:"action"`
	Member User   `json:"member"`
}

func (p MemberPayload) Format(env Event) string {
	return fmt.Sprintf("%s %s as a collaborator to %s",
		color.MagentaString(asciiLowerToTitle(p.Action)), p.Member.DisplayLogin, env.Repo.Name)
}

type PublicPayload struct {
	// Haven't needed anything here yet.
}

func (p PublicPayload) Format(env Event) string {
	return fmt.Sprintf("%s %s public", color.CyanString("Made"), env.Repo.Name)
}

type PullRequestReviewPayload struct {
	Action      string `json:"action"`
	PullRequest PR     `json:"pull_request"`
	Review      Review `json:"review"`
}

type Review struct {
	Body  string `json:"body"`
	State string `json:"state"`
}

func (p PullRequestReviewPayload) Format(env Event) string {
	return fmt.Sprintf("%s a review on PR #%d in %s:\n%s",
		color.CyanString(asciiLowerToTitle(p.Action)), p.PullRequest.Number, env.Repo.Name, quote(p.Review.Body))
}

type PullRequestReviewCommentPayload struct {
	Action      string  `json:"action"`
	PullRequest PR      `json:"pull_request"`
	Comment     Comment `json:"comment"`
}

func (p PullRequestReviewCommentPayload) Format(env Event) string {
	return fmt.Sprintf("%s a review comment on PR #%d in %s:\n%s",
		color.CyanString(asciiLowerToTitle(p.Action)), p.PullRequest.Number, env.Repo.Name, quote(p.Comment.Body))
}

type ReleasePayload struct {
	Action  string  `json:"action"`
	Release Release `json:"release"`
}

type Release struct {
	TagName string `json:"tag_name"`
	Name    string `json:"name"`
}

func (p ReleasePayload) Format(env Event) string {
	name := p.Release.Name
	if name == "" {
		name = p.Release.TagName
	}
	return fmt.Sprintf("%s release %s in %s",
		color.GreenString(asciiLowerToTitle(p.Action)), name, env.Repo.Name)
}
