package lucy

type DomainType uint

const (
	Where        DomainType = iota
	Relation     DomainType = iota
	Creation     DomainType = iota
	Updation     DomainType = iota
	Deletion     DomainType = iota
	SetTarget    DomainType = iota
	Unknown      DomainType = iota
	And          DomainType = iota
	Or           DomainType = iota
	MiscNodeName DomainType = iota
	Model        DomainType = iota
	AndStr       DomainType = iota
	OrStr        DomainType = iota
	WhereStr     DomainType = iota
)
