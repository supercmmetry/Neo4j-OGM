package lucy

type FamilyType uint

const (
	Where        FamilyType = iota
	Relation     FamilyType = iota
	Creation     FamilyType = iota
	Updation     FamilyType = iota
	Deletion     FamilyType = iota
	SetTarget    FamilyType = iota
	Unknown      FamilyType = iota
	And          FamilyType = iota
	Or           FamilyType = iota
	MiscNodeName FamilyType = iota
	Model        FamilyType = iota
	AndStr       FamilyType = iota
	OrStr        FamilyType = iota
	WhereStr     FamilyType = iota
)
