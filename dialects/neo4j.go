package lucy

import (
	"fmt"
	"github.com/neo4j/neo4j-go-driver/neo4j"
	"reflect"

	e "github.com/supercmmetry/lucy/internal"

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

func (n *Neo4jRuntime) marshalToCypherExp(exp e.Exp) string {
	baseStr := ""

	for k, v := range exp {
		baseStr += fmt.Sprintf("%s:%s ,", k, v)
	}
	return baseStr[:len(baseStr)-1]
}

func (n *Neo4jRuntime) marshalToCypherBody(exp e.Exp) string {
	baseStr := ""

	for k, v := range exp {
		baseStr += fmt.Sprintf("%s = %s , ", k, v)
	}
	return baseStr[:len(baseStr)-4]
}

func (n *Neo4jRuntime) CheckForInjection(expStr string) (uint, bool) {
	pcStr := InQuoteRegex.ReplaceAllString(strings.ToUpper(expStr), "")
	splStr := strings.Split(pcStr, " ")

	severity := e.NoSeverity

	for _, clause := range CypherClauses {
		for _, substr := range splStr {
			if substr == clause {
				severity = e.LowSeverity
				for _, hclause := range HighSeverityClauses {
					if hclause == clause {
						return e.HighSeverity, true
					}
				}
			}
		}
	}
	return uint(severity), severity != e.NoSeverity
}

func (n *Neo4jRuntime) Compile(cradle *e.QueryCradle) (string, error) {
	targetAction := ""
	className := ""
	nodeName := ""
	queryBody := ""

	for _, op := range *cradle.Ops.GetAll() {
		switch op {
		case e.Model:
			exp, err := cradle.Exps.Get()
			if err != nil {
				return "", err
			}
			className = exp.(string)
		case e.SetTarget:
			targetAction = "MATCH"
			if reflect.TypeOf(cradle.Out).Kind() != reflect.Struct {
				className = reflect.TypeOf(cradle.Out).Elem().Name()
			} else {
				className = reflect.TypeOf(cradle.Out).Name()
			}

			if nodeName == "" {
				nodeName = "n"
			}

			if queryBody == "" {
				exp, err := cradle.Exps.Get()
				if err != nil {
					return "", err
				}
				queryBody = n.marshalToCypherExp(exp.(e.Exp))
				return fmt.Sprintf("MATCH (%s: %s {%s}) RETURN {result: %s}", nodeName, className, queryBody, nodeName), nil
			}

			genQuery := fmt.Sprintf("%s (%s: %s) %s RETURN {result: %s}", targetAction, nodeName, className, queryBody, nodeName)
			genQuery = n.prefixNodeName(genQuery, nodeName)
			return genQuery, nil
		case e.Creation:
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
			genQuery := fmt.Sprintf("CREATE (%s:%s {%s})", nodeName, className, n.marshalToCypherExp(exp.(e.Exp)))
			return genQuery, nil
		case e.Where:
			queryBody = "WHERE"
			expression, err := cradle.Exps.Get()
			if err != nil {
				return "", err
			}

			queryBody = queryBody + " " + expression.(string)
		case e.And:
			queryBody += " and"
			expression, err := cradle.Exps.Get()
			if err != nil {
				return "", err
			}
			queryBody = queryBody + " " + expression.(string)
		case e.Or:
			queryBody += " or"
			expression, err := cradle.Exps.Get()
			if err != nil {
				return "", err
			}
			queryBody = queryBody + " " + expression.(string)
		case e.MiscNodeName:
			expression, err := cradle.Exps.Get()
			if err != nil {
				return "", err
			}
			nodeName = expression.(string)
		case e.Updation:

			if reflect.TypeOf(cradle.Out).Kind() != reflect.Struct {
				className = reflect.TypeOf(cradle.Out).Elem().Name()
			} else if reflect.TypeOf(cradle.Out).Kind() == reflect.Ptr {
				className = reflect.TypeOf(cradle.Out).Name()
			}

			if nodeName == "" {
				nodeName = "n"
			}

			exp, err := cradle.Exps.Get()
			if err != nil {
				return "", err
			}

			queryBody = n.prefixNodeName(queryBody, nodeName)
			genQuery := fmt.Sprintf("MATCH (%s: %s) %s SET %s = {%s}", nodeName, className, queryBody, nodeName,
				n.prefixNodeName(n.marshalToCypherExp(exp.(e.Exp)), nodeName))

			return genQuery, nil
		case e.UpdationStr:

			if nodeName == "" {
				nodeName = "n"
			}

			exp, err := cradle.Exps.Get()
			if err != nil {
				return "", err
			}

			queryBody = n.prefixNodeName(queryBody, nodeName)
			genQuery := fmt.Sprintf("MATCH (%s: %s) %s SET %s", nodeName, className, queryBody,
				n.prefixNodeName(exp.(string), nodeName))

			return genQuery, nil
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
		e.Unmarshal(node.Props(), target)
	}
	return nil
}

func NewNeo4jRuntime(driver neo4j.Driver) e.QueryRuntime {
	runtime := &Neo4jRuntime{}
	if session, err := driver.Session(neo4j.AccessModeWrite); err != nil {
		panic(err)
	} else {
		runtime.session = session
	}
	return runtime
}
