package main

import (
	"fmt"
	"github.com/neo4j/neo4j-go-driver/neo4j"
	lucy "github.com/supercmmetry/lucy/core"
	dialects "github.com/supercmmetry/lucy/dialects"
	"time"
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
	lucifer.AddRuntime(dialects.NewNeo4jRuntime(driver))

	db := lucifer.DB()

	defer db.Close()

	t := time.Now()

	// Delete nodes
	err = db.Model(DscDeveloper{}).Where("age > 0").Delete().Error

	// Create nodes in DB
	err = db.Create(DscDeveloper{Name: "Vishaal Selvaraj", Age: 19, Position: "Core Member"}).Error
	err = db.Create(DscDeveloper{Name: "Amogh Lele", Age: 20, Position: "Core Member"}).Error
	err = db.Create(DscDeveloper{Name: "Angad Sharma", Age: 21, Position: "Community Lead"}).Error

	// Declare structs
	vishaal := &DscDeveloper{}
	amogh := &DscDeveloper{}
	angad := &DscDeveloper{}

	// Load data into structs
	err = db.Where("name = ? and age >= ?", "Vishaal Selvaraj", 18).Find(vishaal).Error
	err = db.Where("name = ? and age is not null", "Amogh Lele").Find(amogh).Error
	err = db.Where("position = ?", "Community Lead").Find(angad).Error

	// Update values
	err = db.Find(angad).Set("age = ?", 20).Error

	// Relate nodes
	err = db.Relate(angad).To(amogh).By("HELPS").Error
	err = db.Relate(amogh).To(angad).By("HELPS").Error

	err = db.Relate(vishaal).To(amogh).By("LEARNS_FROM").Error
	err = db.Relate(vishaal).To(angad).By("LEARNS_FROM").Error

	err = db.Relate(amogh).To(vishaal).By("HELPS").Error
	err = db.Relate(angad).To(vishaal).By("HELPS").Error

	// Get node collection
	coreMembers := &[]DscDeveloper{}

	err = db.Where("position = ?", "Core Member").Find(coreMembers).Error
	fmt.Println("Some core members at DSC: ", coreMembers)

	// Get nodes using relation
	someOfMyTeachers := &[]DscDeveloper{}

	err = db.Find(vishaal).Relation("LEARNS_FROM").Find(someOfMyTeachers).Error
	fmt.Println("Some teachers at DSC: ", someOfMyTeachers)

	if err != nil {
		panic(err)
	}

	fmt.Println(time.Now().Sub(t))

}
