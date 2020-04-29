package main

import (
	"context"
	"encoding/json"
	"os"

	"github.com/google/go-github/github"
	"golang.org/x/oauth2"
)

func main() {
	github_token := os.Getenv("GITHUB_TOKEN")
	if github_token == "" {
		os.Exit(1)
	}

	var github_event_path map[string]interface{}
	json.Unmarshal([]byte(os.Getenv("GITHUB_EVENT_PATH")), &github_event_path)
	action := github_event_path["action"].(string)
	pull_request := github_event_path["pull_request"].(map[string]interface{})
	merged := pull_request["merged"].(string)

	if action != "closed" || merged != "true" {
		os.Exit(0)
	}

	// delete the branch
	head := pull_request["head"].(map[string]interface{})
	ref := head["ref"].(string)
	repo := head["repo"].(map[string]interface{})
	login := repo["owner"].(map[string]interface{})["login"].(string)
	name := repo["name"].(string)

	ctx := context.Background()
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: github_token},
	)
	tc := oauth2.NewClient(ctx, ts)
	client := github.NewClient(tc)

	// what is the default branch for the repository
	repository, _, err := client.Repositories.Get(ctx, login, name)
	if err != nil {
		os.Exit(1)
	}
	if ref == *repository.DefaultBranch {
		os.Exit(0)
	}

	branch, _, err := client.Repositories.GetBranch(ctx, login, name, ref)
	if err != nil {
		os.Exit(1)
	}
	if *branch.Protected {
		os.Exit(0)
	}

	options := new(github.PullRequestListOptions)
	options.Base = ref
	pr, _, err := client.PullRequests.List(ctx, login, name, options)
	if err != nil {
		os.Exit(1)
	}
	if len(pr) > 0 {
		os.Exit(0)
	}

	_, err = client.Git.DeleteRef(ctx, login, name, ref)
	if err != nil {
		os.Exit(1)
	}
}
