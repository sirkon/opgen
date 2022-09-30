package main

import (
	"go/ast"
	"go/parser"
	"strings"

	"github.com/sirkon/errors"
)

type cliDefinition struct {
	Version              bool       `short:"v" help:"Show version and exit."`
	OptionsSourcePackage string     `short:"s" help:"Source package to look for options in." required:"true"`
	Dest                 goFileName `short:"d" help:"File name to save generated code in." required:"true"`
	Types                []typeName `arg:"" help:"Type names to generate options builder for." required:"true"`
}

type typeName string

// UnmarshalText to implement encoding.TextUnmarshaler.
func (t *typeName) UnmarshalText(text []byte) error {
	expr, err := parser.ParseExpr(string(text))
	if err != nil {
		return errors.Wrapf(err, "parse option %v", text)
	}

	switch v := expr.(type) {
	case *ast.Ident:
		*t = typeName(v.Name)
	default:
		return errors.Newf("invalid type name %s", string(text))
	}

	return nil
}

type goFileName string

// UnmarshalText to implement encoding.TextUnmarshaler.
func (t *goFileName) UnmarshalText(text []byte) error {
	if !strings.HasSuffix(string(text), ".go") {
		return errors.Newf("invalid go file name '%s'", string(text))
	}

	*t = goFileName(text)
	return nil
}
