package address

import "fmt"

type ParserType string

type Parser interface {
	ParseRawAddress(raw string) (string, error)
}

type manager struct {
	parsers map[ParserType]Parser
}

func (m *manager) register(name ParserType, sig Parser) {
	m.parsers[name] = sig
}

func (m *manager) get(name ParserType) (Parser, error) {
	p, ok := m.parsers[name]

	if !ok {
		return nil, fmt.Errorf("not supported chain type: %v", name)
	}

	return p, nil
}

var defaultManager = manager{
	parsers: make(map[ParserType]Parser),
}

func GetParser(name ParserType) (Parser, error) {
	return defaultManager.get(name)
}
