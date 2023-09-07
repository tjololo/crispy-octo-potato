/*
Copyright © 2023 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"encoding/json"
	"fmt"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/tjololo/crispy-octo-potato/processors"
	"os"
	"regexp"
)

// searchIssuesCmd represents the searchIssues command
var searchIssuesCmd = &cobra.Command{
	Use:   "search-issues [file/to/search.json]",
	Args:  cobra.ExactArgs(1),
	Short: "Find issues matching regexes",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Searching for matches....")
		regexes, err := cmd.Flags().GetStringToString("regex")
		if err != nil {
			panic(err)
		}
		rf := viper.GetStringMapString("regexes")
		matchIdPath := viper.GetString("identifierPath")
		matchPaths := viper.GetStringSlice("matchPaths")
		regexes = addNonConflictinMapEntries(regexes, rf)
		file := args[0]
		bytes, err := os.ReadFile(file)
		if err != nil {
			fmt.Printf("Unable to read file %s\n", file)
		}
		var data []map[string]interface{}
		err = json.Unmarshal(bytes, &data)
		if err != nil {
			fmt.Printf("Unable to deserilize file %s\n", file)
		}
		for k, v := range regexes {
			ch := make(chan processors.MatchResult)
			go processors.FindObjectMatchingQuery(processors.MatchQuery{
				Name:                k,
				Regex:               regexp.MustCompile(v),
				MatchIdentifierPath: matchIdPath,
				MatchPaths:          matchPaths,
				Data:                data,
			}, ch)
			for {
				res, ok := <-ch
				if !ok {
					break
				}
				fmt.Printf("Element '%s' matched regex '%s' in object '%s'\n", res.MatchingElement, res.Name, res.Identifier)
			}
		}
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
	searchIssuesCmd.Flags().StringP("file-to-search", "f", "", "Json file with list of elements")
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
