package lucy

type FamilyType uint

const (
	Where FamilyType = iota
	RelationX
	RelationY
	Creation
	Updation
	UpdationStr
	Deletion
	SetTarget
	Unknown
	And
	Or
	By
	Model
	MTRelation
)
