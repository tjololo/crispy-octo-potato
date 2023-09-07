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
var searchFilesCmd = &cobra.Command{
	Use:   "search-file [file/to/search.json]",
	Args:  cobra.ExactArgs(1),
	Short: "Find fields matching regexes in json file",
	Run: func(cmd *cobra.Command, args []string) {
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
		regexes, err := cmd.Flags().GetStringToString("regex")
		if err != nil {
			panic(err)
		}
		searchData(regexes, data)
		fmt.Println("Search completed")
	},
}

func init() {
	rootCmd.AddCommand(searchFilesCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// searchIssuesCmd.PersistentFlags().String("foo", "", "A help for foo")
	searchFilesCmd.Flags().StringToStringP("regex", "r", make(map[string]string), "Provide regex in addition to those in configfile on the command line example: --regex digits=\\d+")
	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// searchIssuesCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}

func searchData(cmdRegexes map[string]string, data []map[string]interface{}) {
	fmt.Println("Searching for matches....")
	rf := viper.GetStringMapString("regexes")
	matchIdPath := viper.GetString("identifierPath")
	matchPaths := viper.GetStringSlice("matchPaths")
	cmdRegexes = addNonConflictinMapEntries(cmdRegexes, rf)
	for k, v := range cmdRegexes {
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
}
