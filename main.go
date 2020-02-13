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
	peeps := make([]Person, 0)

	db := lucifer.DB()

	defer db.Close()

	t := time.Now()
	// err = db.Create(Person{Name: "Vishaal", Age: 20}).Error




	err = db.Model(peep).Where("name = ?", "Vishaal").And("age >= ?", 18).
		Set("age = ?", 18).Error

	//err = db.Where("name = ?", "Vishaal").And("age >= ?", 18).Find(&peep).Error
	//err = db.Where("name = ?", "Vishaal").And("age >= ?", 18).Find(&peeps).Error
	//err = db.Where("name = ?", "Vishaal").And("age >= ?", 18).Find(&peep).
	//	Set("age = ?", peep.Age + 2).Error
	//
	//err = db.Model(peep).Where("name = ?", "Vishaal").Delete().Error
	//


	if err != nil {
		panic(err)
	}

	fmt.Println("Collection: ", peeps)
	fmt.Println("First Record: ", peep)



	fmt.Println(time.Now().Sub(t))

}
