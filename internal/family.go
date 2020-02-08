package lucy

type FamilyType uint

const (
	Where FamilyType = iota
	Relation
	Creation
	Updation
	UpdationStr
	Deletion
	SetTarget
	Unknown
	And
	Or
	MiscNodeName
	Model
)
