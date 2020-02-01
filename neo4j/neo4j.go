package neo4j

import (
	lucy "Neo4j-OGM/core"
	"github.com/neo4j/neo4j-go-driver/neo4j"
)

type Neo4J struct {
	driver  neo4j.Driver
	session neo4j.Session
	result  neo4j.Result
	lucy    *lucy.Lucy
}

func (n *Neo4J) MapToLucy() {
	n.lucy = &lucy.Lucy{Mapping:n}
}

func (n *Neo4J) Find(param interface{}) error {
	// todo: Implement using Neo4J.
	return nil
}
