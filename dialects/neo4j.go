package dialects

import (
	"github.com/neo4j/neo4j-go-driver/neo4j"
	lucy "github.com/supercmmetry/lucy/core"
	"regexp"
	"strings"
)

type Neo4jRuntime struct {
	driver  neo4j.Driver
	session neo4j.Session
	result  neo4j.Result
}

var (
	InQuoteRegex = regexp.MustCompile("(?:(\"(?:.*?)\")|('(?:.*?)'))")
	CypherClauses = []string{"CREATE", "UPDATE", "MATCH", "RETURN", "WITH", "UNWIND", "WHERE", "EXISTS", "ORDER", "BY",
		"SKIP", "LIMIT", "USING", "DELETE", "DETACH", "REMOVE", "FOR", "EACH", "MERGE", "ON", "CALL", "YIELD", "USE",
	"DROP", "START", "STOP", "SET"}
)

func (n *Neo4jRuntime) CheckForInjection(expStr string) bool {
	pcStr := InQuoteRegex.ReplaceAllString(strings.ToUpper(expStr), "")
	splStr := strings.Split(pcStr, " ")

	for _, clause := range CypherClauses {
		for _, substr := range splStr {
			if substr == clause {
				return true
			}
		}
	}
	return false
}

func (n *Neo4jRuntime) Compile(cradle *lucy.QueryCradle) (string, error) {
	panic("implement me")
}

func (n *Neo4jRuntime) Execute(query string, target interface{}) error {
	panic("implement me")
}

func NewNeo4jRuntime() lucy.QueryRuntime {
	return &Neo4jRuntime{}
}
