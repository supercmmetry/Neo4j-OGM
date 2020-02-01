package lucyErr

import "errors"

var (
	EmptyQueryQueue error = errors.New("Query queue is empty.")
	InvalidFormatString error = errors.New("Invalid string format.")
	EndDomainChanged error = errors.New("End domain changed abruptly.")
)
