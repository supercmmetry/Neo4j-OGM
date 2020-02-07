package dialects

import (
	"fmt"
	"github.com/neo4j/neo4j-go-driver/neo4j"
	"reflect"

	lucy "github.com/supercmmetry/lucy/core"
	lucyErr "github.com/supercmmetry/lucy/errors"

	"regexp"
	"strings"
)

type Neo4jRuntime struct {
	driver  neo4j.Driver
	session neo4j.Session
}

var (
	cypherKeyCaptureRegex = regexp.MustCompile("([^\\s]*?)\\s*(?:(?:<>)|(?:<=)|(?:>=)|(?:IS NULL)|(?:IS NOT NULL)|=|>|<)")
	InQuoteRegex          = regexp.MustCompile("(?:(\"(?:.*?)\")|('(?:.*?)'))")
	CypherClauses         = []string{"CREATE", "UPDATE", "MATCH", "RETURN", "WITH", "UNWIND", "WHERE", "EXISTS", "ORDER", "BY",
		"SKIP", "LIMIT", "USING", "DELETE", "DETACH", "REMOVE", "FOR", "EACH", "MERGE", "ON", "CALL", "YIELD", "USE",
		"DROP", "START", "STOP", "SET"}
	HighSeverityClauses = []string{"DELETE", "DETACH", "REMOVE", "DROP", "SET", "UPDATE", "CALL", "CREATE"}
)

func (n *Neo4jRuntime) prefixNodeName(query string, nodeName string) string {
	matches := cypherKeyCaptureRegex.FindAllString(query, -1)
	for _, m := range matches {
		if !strings.Contains(m, ".") {
			query = strings.Replace(query, m, nodeName+"."+m, -1)
		}
	}
	return query
}

func (n *Neo4jRuntime) marshalToCypherExp(exp lucy.Exp) string {
	baseStr := ""

	for k, v := range exp {
		baseStr += fmt.Sprintf("%s: %s,", k, v)
	}
	return baseStr[:len(baseStr)-1]
}

func (n *Neo4jRuntime) CheckForInjection(expStr string) (uint, bool) {
	pcStr := InQuoteRegex.ReplaceAllString(strings.ToUpper(expStr), "")
	splStr := strings.Split(pcStr, " ")

	severity := lucyErr.NoSeverity

	for _, clause := range CypherClauses {
		for _, substr := range splStr {
			if substr == clause {
				severity = lucyErr.LowSeverity
				for _, hclause := range HighSeverityClauses {
					if hclause == clause {
						return lucyErr.HighSeverity, true
					}
				}
			}
		}
	}
	return uint(severity), severity != lucyErr.NoSeverity
}

func (n *Neo4jRuntime) Compile(cradle *lucy.QueryCradle) (string, error) {
	targetAction := ""
	className := ""
	nodeName := ""
	queryBody := ""

	for _, op := range *cradle.Ops.GetAll() {
		switch op {
		case lucy.SetTarget:
			targetAction = "MATCH"
			if reflect.TypeOf(cradle.Out).Kind() != reflect.Struct {
				className = reflect.TypeOf(cradle.Out).Elem().Name()
			} else {
				className = reflect.TypeOf(cradle.Out).Name()
			}

			if nodeName == "" {
				nodeName = "n"
			}

			genQuery := fmt.Sprintf("%s (%s: %s) %s RETURN {result: %s}", targetAction, nodeName, className, queryBody, nodeName)
			genQuery = n.prefixNodeName(genQuery, nodeName)
			return genQuery, nil

		case lucy.Creation:
			if reflect.TypeOf(cradle.Out).Kind() != reflect.Struct {
				className = reflect.TypeOf(cradle.Out).Elem().Name()
			} else {
				className = reflect.TypeOf(cradle.Out).Name()
			}

			if nodeName == "" {
				nodeName = "n"
			}

			exp, err := cradle.Exps.Get()
			if err != nil {
				return "", err
			}
			genQuery := fmt.Sprintf("CREATE (%s:%s {%s})", nodeName, className, n.marshalToCypherExp(exp.(lucy.Exp)))
			return genQuery, nil
		case lucy.WhereStr:
			queryBody = "WHERE"
			expression, err := cradle.Exps.Get()
			if err != nil {
				return "", err
			}

			queryBody = queryBody + " " + expression.(string)
		case lucy.AndStr:
			queryBody += " and"
			expression, err := cradle.Exps.Get()
			if err != nil {
				return "", err
			}
			queryBody = queryBody + " " + expression.(string)
		case lucy.OrStr:
			queryBody += " or"
			expression, err := cradle.Exps.Get()
			if err != nil {
				return "", err
			}
			queryBody = queryBody + " " + expression.(string)
		case lucy.MiscNodeName:
			expression, err := cradle.Exps.Get()
			if err != nil {
				return "", err
			}
			nodeName = expression.(string)
		}
	}

	return "", nil
}

func (n *Neo4jRuntime) Execute(query string, target interface{}) error {
	result, err := n.session.Run(query, map[string]interface{}{})
	if err != nil {
		return err
	}
	for result.Next() {
		record := result.Record().GetByIndex(0).(map[string]interface{})
		node := record["result"].(neo4j.Node)
		lucy.Unmarshal(node.Props(), target)
	}
	return nil
}

func NewNeo4jRuntime(driver neo4j.Driver) lucy.QueryRuntime {
	runtime := &Neo4jRuntime{}
	if session, err := driver.Session(neo4j.AccessModeWrite); err != nil {
		panic(err)
	} else {
		runtime.session = session
	}
	return runtime
}
