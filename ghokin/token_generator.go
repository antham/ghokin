package ghokin

import (
	"github.com/cucumber/gherkin-go"
)

type tokenGenerator struct {
	section *section
}

func (t *tokenGenerator) Build(tok *gherkin.Token) (bool, error) {
	if tok.IsEOF() {
		return true, nil
	}

	switch true {
	case t.section == nil:
		t.section = &section{kind: tok.Type, values: []*gherkin.Token{}}
	case tok.Type != t.section.kind:
		t.section.nex = &section{kind: tok.Type, values: []*gherkin.Token{}, prev: t.section}
		t.section = t.section.nex
	}

	t.section.values = append(t.section.values, tok)

	return true, nil
}

func (t *tokenGenerator) StartRule(r gherkin.RuleType) (bool, error) {
	return true, nil
}
func (t *tokenGenerator) EndRule(r gherkin.RuleType) (bool, error) {
	return true, nil
}
func (t *tokenGenerator) Reset() {
}

type section struct {
	kind   gherkin.TokenType
	values []*gherkin.Token
	prev   *section
	nex    *section
}

func (s *section) isExcluded(kind gherkin.TokenType, excluded []gherkin.TokenType) bool {
	for _, e := range excluded {
		if kind == e {
			return true
		}
	}

	return false
}

func (s *section) previous(excluded []gherkin.TokenType) *section {
	for sec := s.prev; sec != nil; sec = sec.prev {
		if !s.isExcluded(sec.kind, excluded) {
			return sec
		}
	}

	return nil
}

func (s *section) next(excluded []gherkin.TokenType) *section {
	for sec := s.nex; sec != nil; sec = sec.nex {
		if !s.isExcluded(sec.kind, excluded) {
			return sec
		}
	}

	return nil
}
