package lucy

type DomainType uint

const (
	Where     DomainType = 0
	Relation  DomainType = 1
	Creation  DomainType = 2
	Updation  DomainType = 3
	Deletion  DomainType = 4
	SetTarget DomainType = 6
	Unknown   DomainType = 7
)
