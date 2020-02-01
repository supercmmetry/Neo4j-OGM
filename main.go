package main

import (
	"fmt"
	lucy "lucy/core"
	"lucy/neo4j"
	"time"
)

type Person struct {
	Name string `lucy:"name"`
	Age  uint   `lucy:"age"`
}

func main() {
	fmt.Println("lucy - devel")

	t := time.Now()
	

	lucifer := lucy.Lucy{}
	lucifer.AddRuntime(neo4j.NewNeo4jRuntime())

	peep := Person{}

	tx := lucifer.DB().Begin()
	err := tx.Where(lucy.Exp{"name":"Vishaal"}).And("age=%d", 19).Find(&peep).Error


	fmt.Println(time.Now().Sub(t))

	if err != nil {
		panic(err)
	}

}
