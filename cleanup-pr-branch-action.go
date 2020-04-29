package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"

	"github.com/google/go-github/github"
	"golang.org/x/oauth2"
)

func main() {
	github_token := os.Getenv("GITHUB_TOKEN")
	if github_token == "" {
		fmt.Println("GITHUB_TOKEN not set")
		os.Exit(1)
	}

	github_event_json, err := ioutil.ReadFile(os.Getenv("GITHUB_EVENT_PATH"))
	if err != nil {
		fmt.Printf("error reading event: %s", err.Error())
	}
	var event map[string]interface{}
	json.Unmarshal(github_event_json, &event)
	fmt.Println(event)
	action := event["action"].(string)
	pull_request := event["pull_request"].(map[string]interface{})
	merged := pull_request["merged"].(string)

	if action != "closed" || merged != "true" {
		fmt.Println("pull request not closed and merged")
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
		fmt.Printf("error looking up repository %s/%s: %s", login, name, err.Error())
		os.Exit(1)
	}
	if ref == *repository.DefaultBranch {
		fmt.Printf("not deleting default branch %s", ref)
		os.Exit(0)
	}

	branch, _, err := client.Repositories.GetBranch(ctx, login, name, ref)
	if err != nil {
		fmt.Printf("error getting branch %s from repository %s/%s: %s", ref, login, name, err.Error())
		os.Exit(1)
	}
	if *branch.Protected {
		fmt.Printf("branch %s is protected", ref)
		os.Exit(0)
	}

	options := new(github.PullRequestListOptions)
	options.Base = ref
	pr, _, err := client.PullRequests.List(ctx, login, name, options)
	if err != nil {
		fmt.Printf("error listing pull requests with base %s from repository %s/%s: %s", ref, login, name, err.Error())
		os.Exit(1)
	}
	if len(pr) > 0 {
		fmt.Printf("branch %s from repository %s/%s is the base branch of pr %d", ref, login, name, pr[0].Number)
		os.Exit(0)
	}

	_, err = client.Git.DeleteRef(ctx, login, name, ref)
	if err != nil {
		fmt.Printf("error deleting branch %s from repository %s/%s: %s", ref, login, name, err.Error())
		os.Exit(1)
	}
}
