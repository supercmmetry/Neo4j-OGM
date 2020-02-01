package lucyErr

import "errors"

var (
	EmptyQueue                  = errors.New("lucy: queue is empty")
	QueryDependencyNotSatisfied = errors.New("lucy: query dependency was not satisfied")
	QueryChainLogicCorrupted    = errors.New("lucy: query chain logic was corrupted")
	ExpressionExpected          = errors.New("lucy: Expression expected in parameter")
	ExpressionNotRecognized     = errors.New("lucy: Expression not recognized")
	QueryInjectionDetected      = errors.New("lucy: query injection detected")
)
