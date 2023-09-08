/*
Copyright © 2023 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/cli/oauth/api"
	"github.com/cli/oauth/device"
	"github.com/spf13/cobra"
	"io"
	"net/http"
	"os"
)

var clientID = "set-at-build-time"

// searchIssuesCmd represents the searchIssues command
var searchIssuesCmd = &cobra.Command{
	Use:   "search-issues [org/repo]",
	Args:  cobra.ExactArgs(1),
	Short: "Find issues and pr matching regexes in repo, comments can be included in search",
	Run: func(cmd *cobra.Command, args []string) {
		regexes, err := cmd.Flags().GetStringToString("regex")
		if err != nil {
			panic(err)
		}
		private, err := cmd.Flags().GetBool("private-repo")
		if err != nil {
			fmt.Printf("Failed to load flags, %v", err)
		}
		repo := args[0]
		scopes := []string{}
		if private {
			scopes = []string{"repo"}
		}
		auth := fetchGithubToken(cmd.Context(), scopes)
		state, err := cmd.Flags().GetString("state")
		if err != nil {
			fmt.Printf("Failed to read flags, %v", err)
		}
		incComments, err := cmd.Flags().GetBool("comments")
		if err != nil {
			fmt.Printf("Failed to read flags, %v", err)
		}
		data := fetchGithubIssues(repo, state, auth.Token, 0)
		if incComments {
			data = append(data, fetchGithubComments(repo, auth.Token, 0)...)
		}
		fmt.Printf("Found %d issues in %s\n", len(data), repo)
		searchData(regexes, data)
		fmt.Println("Search completed")
	},
}

func init() {
	rootCmd.AddCommand(searchIssuesCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// searchIssuesCmd.PersistentFlags().String("foo", "", "A help for foo")
	searchIssuesCmd.Flags().StringToStringP("regex", "r", make(map[string]string), "Provide regex in addition to those in configfile on the command line example: --regex digits=\\d+")
	searchIssuesCmd.Flags().StringP("state", "s", "open", "Issue state to search for")
	searchIssuesCmd.Flags().BoolP("comments", "c", false, "Include comments")
	searchIssuesCmd.Flags().BoolP("private-repo", "p", false, "Adds repo scope to token to query private repo")
	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// searchIssuesCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}

func addNonConflictinMapEntries(base, additions map[string]string) map[string]string {
	for k, v := range additions {
		if _, ok := base[k]; !ok {
			base[k] = v
		}
	}
	return base
}

func fetchGithubToken(ctx context.Context, scopes []string) *api.AccessToken {
	httpClient := http.DefaultClient
	code, err := device.RequestCode(httpClient, "https://github.com/login/device/code", clientID, scopes)
	if err != nil {
		fmt.Printf("Unable to retrive device code %v", err)
		os.Exit(1)
	}
	fmt.Printf("Please open %s and provide device code to authorize crispy-octo-potato to retrive an oAuth token\nDevice code: %s\n%s", code.VerificationURI, code.UserCode, code.VerificationURIComplete)
	waitOpt := device.WaitOptions{
		ClientID:   clientID,
		DeviceCode: code,
	}
	accessToken, err := device.Wait(ctx, httpClient, "https://github.com/login/oauth/access_token", waitOpt)
	if err != nil {
		fmt.Printf("Unable to retrive auth token %v", err)
		os.Exit(1)
	}
	return accessToken

}

func fetchGithubIssues(repo, state, token string, page int) []map[string]interface{} {
	url := fmt.Sprintf("https://api.github.com/repos/%s/issues?state=%s&per_page=%d&page=%d", repo, state, 100, page)
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		fmt.Printf("Failed to build request, %v", err)
		os.Exit(1)
	}
	req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", token))
	req.Header.Add("Accept", "application/vnd.github+json")
	req.Header.Add("X-GitHub-Api-Version", "2022-11-28")
	client := http.DefaultClient
	resp, err := client.Do(req)
	if err != nil {
		fmt.Printf("Request to fetch issues failed, %v\n", err)
		os.Exit(1)
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Printf("Failed to read respons body, %v\n", err)
	}
	var data []map[string]interface{}
	err = json.Unmarshal(body, &data)
	if err != nil {
		fmt.Printf("Unable to deserilize response %v\n", err)
		os.Exit(1)
	}
	if len(data) == 100 {
		nextPageNumber := page + 1
		data = append(data, fetchGithubIssues(repo, state, token, nextPageNumber)...)
	}
	return data
}
