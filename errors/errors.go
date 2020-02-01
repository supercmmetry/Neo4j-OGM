package lucyErr

import "errors"

var (
	EmptyQueryQueue             = errors.New("lucy: query queue is empty")
	EndDomainChanged            = errors.New("lucy: end domain changed abruptly")
	QueryDependencyNotSatisfied = errors.New("lucy: query dependency was not satisfied")
	QueryChainLogicCorrupted    = errors.New("lucy: query chain logic was corrupted")
)
