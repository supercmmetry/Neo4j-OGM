package neo4j

import (
	lucy "Neo4j-OGM/core"
	"github.com/neo4j/neo4j-go-driver/neo4j"
)

type Neo4jRuntime struct {
	driver  neo4j.Driver
	session neo4j.Session
	result  neo4j.Result
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





