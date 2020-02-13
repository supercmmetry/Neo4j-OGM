package lucy

const (
	EmptyQueue = iota
	UnsatisfiedDependency
	CorruptedQueryChain
	ExpressionExpected
	UnrecognizedExpression
	QueryInjection
	NoRecordsFound
)

const (
	NoSeverity = iota
	LowSeverity
	HighSeverity
)

var (
	lucyErrors        = (&LucyErrors{}).Init()
	injectionSeverity = (&InjectionSeverity{}).Init()
)

type LucyErrors struct {
	Code     uint
	Data     string
	errorMap map[uint]string
}

func (e *LucyErrors) Init() *LucyErrors {
	e.Data = ""
	e.errorMap = map[uint]string{
		EmptyQueue:             "lucy: queue is empty",
		UnsatisfiedDependency:  "lucy: query dependency was not satisfied",
		CorruptedQueryChain:    "lucy: query chain logic was corrupted",
		ExpressionExpected:     "lucy: expression expected in parameter",
		UnrecognizedExpression: "lucy: expression not recognized",
		QueryInjection:         "lucy: query injection detected",
		NoRecordsFound:         "lucy: No records found",
	}
	return e
}

func (e *LucyErrors) Error() string {
	errStr := e.errorMap[e.Code]
	if len(e.Data) > 0 {
		return errStr + " [" + e.Data + "]"
	}
	return errStr
}

func Error(code uint, data ...string) error {
	joinData := ""

	for _, sub := range data {
		joinData += sub
	}
	lucyErrors.Code = code
	lucyErrors.Data = joinData
	return lucyErrors
}

type InjectionSeverity struct {
	Code   uint
	sevMap map[uint]string
}

func (s *InjectionSeverity) Init() *InjectionSeverity {
	s.sevMap = map[uint]string{
		HighSeverity: "severity: high",
		LowSeverity:  "severity: low",
		NoSeverity:   "",
	}
	return s
}

func (s *InjectionSeverity) Severity() string {
	return s.sevMap[s.Code]
}

func Severity(code uint) string {
	injectionSeverity.Code = code
	return injectionSeverity.Severity()
}
