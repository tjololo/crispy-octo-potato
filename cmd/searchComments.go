/*
Copyright © 2023 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"

	"github.com/spf13/cobra"
)

// searchCommentsCmd represents the searchComments command
var searchCommentsCmd = &cobra.Command{
	Use:   "search-comments [org/repo]",
	Short: "Find comments matching regexes in repo",
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
		data := fetchGithubComments(repo, auth.Token, 0)
		fmt.Printf("Found %d issues in %s\n", len(data), repo)
		searchData(regexes, data)
		fmt.Println("Search completed")
	},
}

func init() {
	rootCmd.AddCommand(searchCommentsCmd)
	searchCommentsCmd.Flags().StringToStringP("regex", "r", make(map[string]string), "Provide regex in addition to those in configfile on the command line example: --regex digits=\\d+")
	searchCommentsCmd.Flags().BoolP("private-repo", "p", false, "Adds repo scope to token to query private repo")
}

func fetchGithubComments(repo, token string, page int) []map[string]interface{} {
	url := fmt.Sprintf("https://api.github.com/repos/%s/issues/comments?per_page=%d&page=%d", repo, 100, page)
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
		fmt.Printf("Request to fetch comments failed, %v\n", err)
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
		data = append(data, fetchGithubComments(repo, token, nextPageNumber)...)
	}
	return data
}
