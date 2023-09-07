package processors

import (
	"fmt"
	"regexp"
	"strings"
	"sync"
)

func FindObjectMatchingQuery(query MatchQuery, ch chan MatchResult) {
	wg := new(sync.WaitGroup)
	wg.Add(len(query.Data))
	for _, v := range query.Data {
		go matchObject(query.Name, query.MatchIdentifierPath, query.Regex, query.MatchPaths, v, ch, wg)
	}
	wg.Wait()
	close(ch)
}

func matchObject(name string, matchIdPath string, regex *regexp.Regexp, matchPaths []string, data map[string]interface{}, ch chan MatchResult, wg *sync.WaitGroup) {
	defer wg.Done()
	for _, v := range matchPaths {
		matchPathInObject(name, matchIdPath, regex, v, data, ch)
	}
}

func matchPathInObject(name string, matchIdPath string, regex *regexp.Regexp, matchPath string, data map[string]interface{}, ch chan MatchResult) {
	val := getDataOnPath(strings.Split(matchPath, "."), data)
	if regex.Match(val) {
		ch <- MatchResult{
			Name:            name,
			Identifier:      string(getDataOnPath(strings.Split(matchIdPath, "."), data)),
			MatchingElement: matchPath,
			MatchContext:    string(regex.Find(val)),
		}
	}
}

func getDataOnPath(pathArr []string, data map[string]interface{}) []byte {
	k := pathArr[0]
	if len(pathArr) == 1 {
		return []byte(fmt.Sprint(data[k]))
	}
	return getDataOnPath(pathArr[1:], data[k].(map[string]interface{}))
}

type MatchQuery struct {
	Name                string
	Regex               *regexp.Regexp
	MatchIdentifierPath string
	MatchPaths          []string
	Data                []map[string]interface{}
}

type MatchResult struct {
	Name            string
	Identifier      string
	MatchingElement string
	MatchContext    string
}
