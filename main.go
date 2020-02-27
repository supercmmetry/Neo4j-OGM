package main

import (
	"fmt"
	"github.com/neo4j/neo4j-go-driver/neo4j"
	lucy "github.com/supercmmetry/lucy/core"
	dialects "github.com/supercmmetry/lucy/dialects"
	lucytype "github.com/supercmmetry/lucy/types"
)

type DscDeveloper struct {
	Name     string `lucy:"name"`
	Age      int    `lucy:"age"`
	Position string `lucy:"position"`
}

func main() {
	fmt.Println("lucy - devel")

	driver, err := neo4j.NewDriver("bolt://localhost:7687", neo4j.BasicAuth("neo4j", "password", ""))

	if err != nil {
		panic(err)
	}

	lucifer := lucy.Lucy{}
	lucifer.AddRuntime(dialects.NewNeo4jRuntime(&driver))

	db := lucifer.DB()

	defer db.Close()

	tx := db.Begin()
	// Delete nodes
	err = tx.Model(DscDeveloper{}).Where("age > 0").Delete().Error

	// Create nodes in DB
	err = tx.Create(DscDeveloper{Name: "Vishaal Selvaraj", Age: 19, Position: "Core Member"}).Error
	err = tx.Create(DscDeveloper{Name: "Amogh Lele", Age: 20, Position: "Core Member"}).Error
	err = tx.Create(DscDeveloper{Name: "Angad Sharma", Age: 21, Position: "Community Lead"}).Error

	if err != nil {
		fmt.Println(err)
		tx.Rollback()
	}

	// Declare structs
	vishaal := &DscDeveloper{}
	amogh := &DscDeveloper{}
	angad := &DscDeveloper{}

	// Load data into structs
	err = tx.Where("name = ? and age >= ?", "Vishaal Selvaraj", 18).Find(vishaal).Error
	err = tx.Where("name = ? and age is not null", "Amogh Lele").Find(amogh).Error
	err = tx.Where("position = ?", "Community Lead").Find(angad).Error

	// Update values
	err = tx.Find(angad).Set("age = ?", 20).Error

	if err != nil {
		fmt.Println(err)
		tx.Rollback()
	}

	// Relate nodes
	err = tx.Relate(angad).To(amogh).By("HELPS", dialects.Neo4jBidirectional, lucytype.Exp{"IN": "DevOps"}).Error

	err = tx.Relate(vishaal).To(amogh).By("LEARNS_FROM", dialects.Neo4jUnidirectionalRight, lucytype.Exp{"ABOUT": "Android"}).Error
	err = tx.Relate(vishaal).To(angad).By("LEARNS_FROM").Error

	err = tx.Relate(amogh).To(vishaal).By("HELPS").Error
	err = tx.Relate(angad).To(vishaal).By("HELPS").Error

	if err != nil {
		fmt.Println(err)
		tx.Rollback()
	}

	tx.Commit()

	tx = db.Begin()

	// Get node collection
	coreMembers := &[]DscDeveloper{}

	err = tx.Where("position = ?", "Core Member").Find(coreMembers).Error

	if err != nil {
		fmt.Println(err)
		tx.Rollback()
	}

	fmt.Println("Some core members at DSC: ", coreMembers)

	// Get nodes using relation
	relPeeps := &[]DscDeveloper{}

	err = tx.Find(vishaal).Relation("LEARNS_FROM", dialects.Neo4jUnidirectionalRight, lucytype.Exp{"ABOUT": "Android"}).Find(relPeeps).Error

	if err != nil {
		fmt.Println(err)
		tx.Rollback()
	}

	fmt.Println("DscDevelopers in LEARNS_FROM.Y: ", relPeeps)

	tx.Commit()

	err = db.Error
	if err != nil {
		panic(err)
	}

}
