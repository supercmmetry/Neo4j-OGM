package main

import (
	"fmt"
	lucy "github.com/supercmmetry/lucy/core"
	"github.com/supercmmetry/lucy/dialects"
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


	lucifer.AddRuntime(dialects.NewNeo4jRuntime())

	peep := Person{}

	db := lucifer.DB()

	err := db.Where(lucy.Exp{"name": "Vishaal"}).And("age >= ? SET", 19).Find(&peep).Error

	fmt.Println(time.Now().Sub(t))

	if err != nil {
		panic(err)
	}

}
