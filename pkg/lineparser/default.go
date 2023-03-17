package lineparser

type DefaultLineParser struct {
}

func NewDefaultLineParser() *DefaultLineParser {
	return &DefaultLineParser{}
}

func (*DefaultLineParser) Parse(line string) ([]byte, error) {
	return []byte(line), nil
}
