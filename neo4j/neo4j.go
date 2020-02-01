package neo4j

import (
	"github.com/neo4j/neo4j-go-driver/neo4j"
	lucy "lucy/core"
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
	Neo4jInjectionRegex = regexp.MustCompile("\\s(?:SET)|" +
		"(?:CREATE)|" +
		"(?:UPDATE)|" +
		"(?:MATCH)|" +
		"(?:RETURN)|" +
		"(?:WITH)|" +
		"(?:UNWIND)|" +
		"(?:WHERE)|" +
		"(?:EXISTS)|" +
		"(?:ORDER BY)|" +
		"(?:SKIP)|" +
		"(?:LIMIT)|" +
		"(?:USING)|" +
		"(?:DELETE)|" +
		"(?:DETACH)|" +
		"(?:REMOVE)|" +
		"(?:FOR EACH)|" +
		"(?:MERGE)|" +
		"(?:ON CREATE)|" +
		"(?:ON MATCH)|" +
		"(?:CALL)|" +
		"(?:YIELD)|" +
		"(?:USE)|" +
		"(?:DROP)|" +
		"(?:START)|" +
		"(?:STOP)")
)

func (n *Neo4jRuntime) CheckForInjection(expStr string) bool {
	pcStr := InQuoteRegex.ReplaceAllString(strings.ToUpper(expStr), "")
	if Neo4jInjectionRegex.MatchString(pcStr) {
		return true
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
