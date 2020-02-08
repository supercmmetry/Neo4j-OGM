package main

import (
	"fmt"
	"github.com/neo4j/neo4j-go-driver/neo4j"
	lucy "github.com/supercmmetry/lucy/core"
	dialects "github.com/supercmmetry/lucy/dialects"
	"time"
)

type Person struct {
	Name string `lucy:"name"`
	Age  int    `lucy:"age"`
}

func main() {
	fmt.Println("lucy - devel")


	driver, err := neo4j.NewDriver("bolt://localhost:7687", neo4j.BasicAuth("neo4j", "password", ""))
	if err != nil {
		panic(err)
	}

	lucifer := lucy.Lucy{}
	lucifer.AddRuntime(dialects.NewNeo4jRuntime(driver))

	peep := Person{}
	db := lucifer.DB()

	t := time.Now()

	// err = db.Create(Person{Name: "Vishaal", Age: 20})

	if err != nil {
		panic(err)
	}

	err = db.Where("name = ?", "Vishaal").And("age >= ?", 18).Find(&peep)
	if err != nil {
		panic(err)
	}

	fmt.Println(time.Now().Sub(t))


	fmt.Println(peep)

}
