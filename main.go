package main

import (
	lucy "Neo4j-OGM/core"
	"Neo4j-OGM/neo4j"
	"fmt"
)

type Person struct {
	Name string `lucy:"name"`
	Age  uint   `lucy:"age"`
}

func main() {
	fmt.Println("lucy - devel")

	lucifer := &lucy.Lucy{}
	engine := (&neo4j.QueryEngine{}).NewQueryEngine()
	engine.AttachTo(lucifer)

	peep := Person{}

	err := lucifer.Where(lucy.Expr{"name": "Vishaal", "age": 19}).Find(&peep).Error
	if err != nil {
		panic(err)
	}


}
