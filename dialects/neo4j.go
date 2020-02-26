package lucy

import (
	"fmt"
	"github.com/neo4j/neo4j-go-driver/neo4j"
	"reflect"

	e "github.com/supercmmetry/lucy/internal"
	t "github.com/supercmmetry/lucy/types"
	"regexp"
	"strings"
)

type RelationType uint

const (
	Neo4jUnidirectionalLeft RelationType = iota
	Neo4jUnidirectionalRight
	Neo4jBidirectional
)

type Neo4jRuntime struct {
	driver      *neo4j.Driver
	session     *neo4j.Session
	transaction *neo4j.Transaction
}

var (
	cypherKeyCaptureRegex = regexp.MustCompile("([^\\s]*?)\\s*(?:(?:<>)|(?:=~)|(?:<=)|(?:>=)|(?:(?i)IS NULL(?-i))" +
		"|(?:(?i)IS NOT NULL(?-i))|(?:(?i)STARTS WITH(?-i))" +
		"|(?:(?i)ENDS WITH(?-i))|(?:(?i)CONTAINS(?-i))" +
		"|\\+|-|=|>|<)")

	InQuoteRegex  = regexp.MustCompile("(?:(\"(?:.*?)\")|('(?:.*?)'))")
	CypherClauses = []string{"CREATE", "UPDATE", "MATCH", "RETURN", "WITH", "UNWIND", "WHERE", "EXISTS", "ORDER", "BY",
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

func (n *Neo4jRuntime) marshalToCypherExp(exp t.Exp) string {
	if len(exp) == 0 {
		return ""
	}
	baseStr := ""

	for k, v := range exp {
		baseStr += fmt.Sprintf("%s:%s ,", k, v)
	}
	return baseStr[:len(baseStr)-1]
}

func (n *Neo4jRuntime) marshalToCypherBody(exp t.Exp) string {
	baseStr := ""

	for k, v := range exp {
		baseStr += fmt.Sprintf("%s = %s , ", k, v)
	}
	return baseStr[:len(baseStr)-4]
}

func (n *Neo4jRuntime) parseRelationToCypher(relName string, relType RelationType, data t.Exp) string {
	switch relType {
	case Neo4jUnidirectionalLeft:
		return fmt.Sprintf("CREATE (n)<-[:%s {%s}]-(m)", relName, n.marshalToCypherExp(data))
	case Neo4jUnidirectionalRight:
		return fmt.Sprintf("CREATE (n)-[:%s {%s}]->(m)", relName, n.marshalToCypherExp(data))
	case Neo4jBidirectional:
		expStr := n.marshalToCypherExp(data)
		return fmt.Sprintf("CREATE (n)-[:%s {%s}]->(m), (n)<-[:%s {%s}]-(m)", relName, expStr, relName, expStr)
	default:
		return ""
	}
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
	kvp := make(map[string]string)
	kvp["Context.Find"] = "Self"
	queryBody := ""

	for _, op := range *cradle.Ops.GetAll() {
		switch op {
		case e.Model:
			exp, err := cradle.Exps.Get()
			if err != nil {
				return "", err
			}
			kvp["className"] = exp.(string)
			break
		case e.SetTarget:
			kvp["className"] = e.GetTypeName(cradle.Out)

			if _, ok := kvp["nodeName"]; !ok {
				kvp["nodeName"] = "n"
			}

			if kvp["Context.Find"] == "Self" {
				exp := e.Marshal(cradle.Out)
				e.SanitizeExp(exp)
				queryBody = n.marshalToCypherExp(exp)
				return fmt.Sprintf("MATCH (%s: %s {%s}) RETURN {result: %s}", kvp["nodeName"],
					kvp["className"], queryBody, kvp["nodeName"]), nil
			} else if kvp["Context.Find"] == "Where" {
				genQuery := fmt.Sprintf("MATCH (%s: %s) %s RETURN {result: %s}", kvp["nodeName"],
					kvp["className"], queryBody, kvp["nodeName"])
				genQuery = n.prefixNodeName(genQuery, kvp["nodeName"])
				return genQuery, nil
			} else if kvp["Context.Find"] == "Relation" {
				genQuery := fmt.Sprintf("MATCH (n: %s {%s})-[:%s]->(m: %s) RETURN {result: m}", kvp["classNameX"],
					kvp["cypherA"], kvp["relName"], kvp["className"])
				return genQuery, nil
			}

		case e.Creation:
			if reflect.TypeOf(cradle.Out).Kind() == reflect.Ptr {
				if reflect.TypeOf(cradle.Out).Elem().Kind() == reflect.Struct {
					kvp["className"] = reflect.TypeOf(cradle.Out).Elem().Name()
				} else if reflect.TypeOf(cradle.Out).Elem().Kind() == reflect.Slice &&
					reflect.TypeOf(cradle.Out).Elem().Elem().Kind() == reflect.Struct {
					kvp["className"] = reflect.TypeOf(cradle.Out).Elem().Elem().Name()
				}
			} else if reflect.TypeOf(cradle.Out).Kind() == reflect.Struct {
				kvp["className"] = reflect.TypeOf(cradle.Out).Name()
			}

			if _, ok := kvp["nodeName"]; !ok {
				kvp["nodeName"] = "n"
			}

			exp, err := cradle.Exps.Get()
			if err != nil {
				return "", err
			}
			genQuery := fmt.Sprintf("CREATE (%s:%s {%s})", kvp["nodeName"], kvp["className"], n.marshalToCypherExp(exp.(t.Exp)))
			return genQuery, nil
		case e.Where:
			queryBody = "WHERE"
			expression, err := cradle.Exps.Get()
			if err != nil {
				return "", err
			}

			queryBody = queryBody + " " + expression.(string)
			kvp["Context.Find"] = "Where"

			break
		case e.And:
			queryBody += " and"
			expression, err := cradle.Exps.Get()
			if err != nil {
				return "", err
			}
			queryBody = queryBody + " " + expression.(string)
			break
		case e.Or:
			queryBody += " or"
			expression, err := cradle.Exps.Get()
			if err != nil {
				return "", err
			}
			queryBody = queryBody + " " + expression.(string)
			break
		case e.Updation:

			if reflect.TypeOf(cradle.Out).Kind() == reflect.Ptr {
				if reflect.TypeOf(cradle.Out).Elem().Kind() == reflect.Struct {
					kvp["className"] = reflect.TypeOf(cradle.Out).Elem().Name()
				} else if reflect.TypeOf(cradle.Out).Elem().Kind() == reflect.Slice &&
					reflect.TypeOf(cradle.Out).Elem().Elem().Kind() == reflect.Struct {
					kvp["className"] = reflect.TypeOf(cradle.Out).Elem().Elem().Name()
				}
			} else if reflect.TypeOf(cradle.Out).Kind() == reflect.Struct {
				kvp["className"] = reflect.TypeOf(cradle.Out).Name()
			}

			if _, ok := kvp["nodeName"]; !ok {
				kvp["nodeName"] = "n"
			}

			exp, err := cradle.Exps.Get()

			if err != nil {
				return "", err
			}

			genQuery := ""
			if queryBody != "" {
				queryBody = n.prefixNodeName(queryBody, kvp["nodeName"])
				genQuery = fmt.Sprintf("MATCH (%s: %s) %s SET %s = {%s} RETURN {result: %s}", kvp["nodeName"], kvp["className"],
					queryBody, kvp["nodeName"],
					n.prefixNodeName(n.marshalToCypherExp(exp.(t.Exp)), kvp["nodeName"]), kvp["nodeName"])
			} else {
				// We haven't encountered a where clause yet. So fetch search params from cradle.out
				marsh := e.Marshal(cradle.Out)
				e.SanitizeExp(marsh)
				cypherA := n.marshalToCypherExp(marsh)
				genQuery = fmt.Sprintf("MATCH (%s: %s {%s}) SET %s = {%s} RETURN {result: %s}", kvp["nodeName"], kvp["className"],
					cypherA,
					kvp["nodeName"], n.prefixNodeName(n.marshalToCypherExp(exp.(t.Exp)), kvp["nodeName"]), kvp["nodeName"])
			}

			return genQuery, nil
		case e.UpdationStr:

			if _, ok := kvp["nodeName"]; !ok {
				kvp["nodeName"] = "n"
			}

			exp, err := cradle.Exps.Get()

			if err != nil {
				return "", err
			}

			genQuery := ""

			// If queryBody is non-empty this means that we have encountered a where clause.
			if queryBody != "" {
				queryBody = n.prefixNodeName(queryBody, kvp["nodeName"])
				genQuery = fmt.Sprintf("MATCH (%s: %s) %s SET %s RETURN {result: %s}", kvp["nodeName"], kvp["className"],
					queryBody,
					n.prefixNodeName(exp.(string), kvp["nodeName"]), kvp["nodeName"])
			} else {
				// We haven't encountered a where clause yet. So fetch search params from cradle.out
				marsh := e.Marshal(cradle.Out)
				e.SanitizeExp(marsh)
				cypherA := n.marshalToCypherExp(marsh)

				if reflect.TypeOf(cradle.Out).Kind() == reflect.Ptr &&
					reflect.TypeOf(cradle.Out).Elem().Kind() == reflect.Struct {

					kvp["className"] = reflect.TypeOf(cradle.Out).Elem().Name()
				} else if reflect.TypeOf(cradle.Out).Kind() == reflect.Struct {
					kvp["className"] = reflect.TypeOf(cradle.Out).Name()
				}

				genQuery = fmt.Sprintf("MATCH (%s: %s {%s}) SET %s RETURN {result: %s}", kvp["nodeName"],
					kvp["className"], cypherA,
					n.prefixNodeName(exp.(string), kvp["nodeName"]), kvp["nodeName"])
			}

			return genQuery, nil
		case e.Deletion:
			genQuery := ""
			if _, ok := kvp["nodeName"]; !ok {
				kvp["nodeName"] = "n"
			}
			if queryBody != "" {
				genQuery = fmt.Sprintf("MATCH (%s: %s) %s DETACH DELETE %s", kvp["nodeName"], kvp["className"],
					queryBody, kvp["nodeName"])
				genQuery = n.prefixNodeName(genQuery, kvp["nodeName"])
			} else {
				// We haven't encountered a where clause yet. So fetch search params from cradle.out
				marsh := e.Marshal(cradle.Out)
				e.SanitizeExp(marsh)
				cypherA := n.marshalToCypherExp(marsh)

				if reflect.TypeOf(cradle.Out).Kind() == reflect.Ptr &&
					reflect.TypeOf(cradle.Out).Elem().Kind() == reflect.Struct {

					kvp["className"] = reflect.TypeOf(cradle.Out).Elem().Name()
				} else if reflect.TypeOf(cradle.Out).Kind() == reflect.Struct {
					kvp["className"] = reflect.TypeOf(cradle.Out).Name()
				}

				genQuery = fmt.Sprintf("MATCH (%s: %s {%s}) DETACH DELETE %s", kvp["nodeName"], kvp["className"],
					cypherA, kvp["nodeName"])
			}

			return genQuery, nil
		case e.RelationX:
			exp, err := cradle.Exps.Get()
			if err != nil {
				return "", err
			}

			if reflect.TypeOf(exp).Kind() == reflect.Ptr {
				if reflect.TypeOf(exp).Elem().Kind() == reflect.Struct {
					kvp["classNameX"] = reflect.TypeOf(exp).Elem().Name()
				} else if reflect.TypeOf(exp).Elem().Kind() == reflect.Slice &&
					reflect.TypeOf(exp).Elem().Elem().Kind() == reflect.Struct {
					return "", e.Error(e.ExpectedStructNotSlice)
				}
			} else if reflect.TypeOf(exp).Kind() == reflect.Struct {
				kvp["classNameX"] = reflect.TypeOf(exp).Name()
			}

			lucyExp := e.Marshal(exp)
			e.SanitizeExp(lucyExp)

			kvp["matchExpX"] = n.marshalToCypherExp(lucyExp)
		case e.RelationY:
			exp, err := cradle.Exps.Get()
			if err != nil {
				return "", err
			}

			if reflect.TypeOf(exp).Kind() == reflect.Ptr {
				if reflect.TypeOf(exp).Elem().Kind() == reflect.Struct {
					kvp["classNameY"] = reflect.TypeOf(exp).Elem().Name()
				} else if reflect.TypeOf(exp).Elem().Kind() == reflect.Slice &&
					reflect.TypeOf(exp).Elem().Elem().Kind() == reflect.Struct {
					return "", e.Error(e.ExpectedStructNotSlice)
				}
			} else if reflect.TypeOf(exp).Kind() == reflect.Struct {
				kvp["classNameY"] = reflect.TypeOf(exp).Name()
			}

			lucyExp := e.Marshal(exp)
			e.SanitizeExp(lucyExp)

			kvp["matchExpY"] = n.marshalToCypherExp(lucyExp)
		case e.By:
			exp, err := cradle.Exps.Get()
			if err != nil {
				return "", err
			}

			genQuery := ""
			pExp := exp.(t.Exp)
			relName := pExp["relation"].(string)
			I := pExp["params"].([]interface{})

			if len(I) == 0 {
				genQuery = fmt.Sprintf("MATCH (n: %s {%s}), (m: %s {%s}) CREATE (n)-[:%s]->(m)", kvp["classNameX"],
					kvp["matchExpX"], kvp["classNameY"], kvp["matchExpY"], relName)
			} else if len(I) == 1 {
				// RelationType passed in parameter.
				relType, ok := I[0].(RelationType)
				if !ok {
					return "", e.Error(e.UnrecognizedExpression)
				}
				genQuery = fmt.Sprintf("MATCH (n: %s {%s}), (m: %s {%s}) %s", kvp["classNameX"],
					kvp["matchExpX"], kvp["classNameY"], kvp["matchExpY"], n.parseRelationToCypher(relName, relType, t.Exp{}))
			} else if len(I) == 2 {
				// RelationType and Data passed in parameters.
				relType, ok := I[0].(RelationType)
				if !ok {
					return "", e.Error(e.UnrecognizedExpression)
				}

				exp, ok := I[1].(t.Exp)
				if !ok {
					return "", e.Error(e.UnrecognizedExpression)
				}
				e.SanitizeExp(exp)

				genQuery = fmt.Sprintf("MATCH (n: %s {%s}), (m: %s {%s}) %s", kvp["classNameX"],
					kvp["matchExpX"], kvp["classNameY"], kvp["matchExpY"], n.parseRelationToCypher(relName, relType, exp))
			}

			return genQuery, nil
		case e.MTRelation:
			kvp["Context.Find"] = "Relation"

			exp, err := cradle.Exps.Get()
			if err != nil {
				return "", err
			}

			pExp := exp.(t.Exp)

			kvp["classNameX"] = e.GetTypeName(pExp["out"])

			outExp := e.Marshal(pExp["out"])
			e.SanitizeExp(outExp)
			kvp["cypherA"] = n.marshalToCypherExp(outExp)
			kvp["relName"] = pExp["relation"].(string)
			break
		}
	}

	return "", nil
}

func (n *Neo4jRuntime) Execute(query string, cradle *e.QueryCradle, target interface{}) error {
	var result neo4j.Result
	var err error
	if n.transaction != nil {
		result, err = (*n.transaction).Run(query, map[string]interface{}{})
	} else {
		result, err = (*n.session).Run(query, map[string]interface{}{})
	}

	if err != nil {
		return err
	}

	if target == nil {
		if !cradle.AllowEmptyResult && !result.Next() {
			return e.Error(e.NoRecordsFound)
		}
		return nil
	}

	targetType := reflect.TypeOf(target)

	if targetType.Kind() == reflect.Ptr && targetType.Elem().Kind() == reflect.Slice &&
		targetType.Elem().Elem().Kind() == reflect.Struct {

		records := make([]map[string]interface{}, 0)
		for result.Next() {
			records = append(records, result.Record().GetByIndex(0).(map[string]interface{}))
		}

		if len(records) == 0 {
			if !cradle.AllowEmptyResult {
				return e.Error(e.NoRecordsFound)
			}
		}

		reflectSlice := reflect.MakeSlice(targetType.Elem(), len(records), len(records))

		for i := 0; i < len(records); i++ {
			temp := reflect.New(targetType.Elem().Elem())
			node := records[i]["result"].(neo4j.Node)
			e.Unmarshal(node.Props(), temp.Interface())
			reflectSlice.Index(i).Set(reflect.ValueOf(temp.Interface()).Elem())
		}

		reflect.ValueOf(target).Elem().Set(reflectSlice)

	} else if targetType.Kind() == reflect.Ptr && targetType.Elem().Kind() == reflect.Struct {
		// Stores the first record in the target.
		if result.Next() {
			record := result.Record().GetByIndex(0).(map[string]interface{})
			node := record["result"].(neo4j.Node)
			e.Unmarshal(node.Props(), target)
		} else {
			if !cradle.AllowEmptyResult {
				return e.Error(e.NoRecordsFound)
			}
		}
	}

	return nil
}

func (n *Neo4jRuntime) Close() error {
	err := (*n.session).Close()
	return err
}

func (n *Neo4jRuntime) Commit() error {
	if n.transaction != nil {
		return (*n.transaction).Commit()
	} else {
		return e.Error(e.InvalidOperation, "cannot commit without transaction")
	}
}

func (n *Neo4jRuntime) Rollback() error {
	if n.transaction != nil {
		return (*n.transaction).Rollback()
	} else {
		return e.Error(e.InvalidOperation, "cannot rollback without transaction")
	}
}

func (n *Neo4jRuntime) BeginTransaction() error {
	t, err := (*n.session).BeginTransaction()
	if err != nil {
		return err
	}

	n.transaction = &t
	return nil
}

func (n *Neo4jRuntime) CloseTransaction() error {
	if n.transaction != nil {
		err := (*n.transaction).Close()
		if err != nil {
			return err
		}
	}

	n.transaction = nil
	return nil
}

func (n *Neo4jRuntime) Clone() e.QueryRuntime {
	return &Neo4jRuntime{
		session:     n.session,
		driver:      n.driver,
		transaction: nil,
	}
}

func NewNeo4jRuntime(driver *neo4j.Driver) e.QueryRuntime {
	runtime := &Neo4jRuntime{}
	runtime.driver = driver
	if session, err := (*driver).Session(neo4j.AccessModeWrite); err != nil {
		panic(err)
	} else {
		runtime.session = &session
	}
	return runtime
}
