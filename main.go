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

	lucifer := lucy.Lucy{}
	lucifer.AddRuntime(neo4j.NewNeo4jRuntime())

	peep := Person{}

	tx := lucifer.DB().Begin()
	err := tx.Where(lucy.Expr{"name":"Vishaal"}).And(lucy.Expr{"age": 19}).Find(&peep).Error

	if err != nil {
		panic(err)
	}

}
