// Package rules is a stripped down version of github.com/traefik/traefik/v2/pkg/rules
package rules

import (
	"strings"

	"github.com/nilskohrs/pathauth/predicate"
)

const (
	and = "and"
	or  = "or"
)

type treeBuilder func() *tree

func andFunc(left, right treeBuilder) treeBuilder {
	return func() *tree {
		return &tree{
			matcher:   and,
			ruleLeft:  left(),
			ruleRight: right(),
		}
	}
}

func orFunc(left, right treeBuilder) treeBuilder {
	return func() *tree {
		return &tree{
			matcher:   or,
			ruleLeft:  left(),
			ruleRight: right(),
		}
	}
}

func invert(t *tree) *tree {
	switch t.matcher {
	case or:
		t.matcher = and
		t.ruleLeft = invert(t.ruleLeft)
		t.ruleRight = invert(t.ruleRight)
	case and:
		t.matcher = or
		t.ruleLeft = invert(t.ruleLeft)
		t.ruleRight = invert(t.ruleRight)
	default:
		t.not = !t.not
	}

	return t
}

func notFunc(elem treeBuilder) treeBuilder {
	return func() *tree {
		return invert(elem())
	}
}

func newParser() (predicate.Parser, error) {
	parserFuncs := make(map[string]interface{})

	for matcherName := range funcs {
		matcherName := matcherName
		fn := func(value ...string) treeBuilder {
			return func() *tree {
				return &tree{
					matcher: matcherName,
					value:   value,
				}
			}
		}
		parserFuncs[matcherName] = fn
		parserFuncs[strings.ToLower(matcherName)] = fn
		parserFuncs[strings.ToUpper(matcherName)] = fn
		parserFuncs[strings.Title(strings.ToLower(matcherName))] = fn
	}

	return predicate.NewParser(predicate.Def{
		Operators: predicate.Operators{
			AND: andFunc,
			OR:  orFunc,
			NOT: notFunc,
		},
		Functions: parserFuncs,
	})
}
