package main

import (
	"go/ast"
	"go/types"
	"strings"
	"unicode"

	"github.com/sirkon/errors"
	"github.com/sirkon/message"
)

func (g *generator) mapOptions() error {
	var names []string
	for _, file := range g.source.Syntax {
		ast.Inspect(file, func(node ast.Node) bool {
			switch v := node.(type) {
			case *ast.GenDecl:
				for _, spec := range v.Specs {
					vv, ok := spec.(*ast.ValueSpec)
					if !ok {
						continue
					}

					for _, name := range vv.Names {
						names = append(names, name.Name)
					}
				}
			case *ast.FuncDecl:
				names = append(names, v.Name.Name)
				return false
			}

			return true
		})
	}

	var failed bool
	for _, name := range names {
		item := g.source.Types.Scope().Lookup(name)
		switch item.(type) {
		case *types.Const, *types.Var, *types.Func:
		default:
			continue
		}

		if owner := g.owner(item.Name()); owner != "" {
			if v, ok := item.(*types.Const); ok {
				// We should prohibit the case of untyped constants.
				switch v.Type().(*types.Basic).Kind() {
				case types.UntypedNil, types.UntypedBool, types.UntypedRune, types.UntypedString,
					types.UntypedInt, types.UntypedFloat, types.UntypedComplex:

					message.Errorf(
						"%s untyped constants are not supported, please specify constant type explicitly",
						g.source.Fset.Position(item.Pos()),
					)
					failed = true
				}
			}

			g.mapping[owner] = append(g.mapping[owner], item)
		}
	}

	if failed {
		return failureError()
	}

	return nil
}

// Look for the type that "owns" this name, i.e. that is a tight prefix
// for the given name. Returns empty string if it doesn't belong to any
// of requested types.
// Tight prefix means it is the prefix and the next letter after the prefix
// part is either capital one or underscore.
func (g *generator) owner(name string) string {
	for _, typ := range g.types {
		if !strings.HasPrefix(name, typ) {
			continue
		}

		rest := []rune(strings.TrimPrefix(typ, name))
		if len(rest) == 0 {
			continue
		}

		if !unicode.IsUpper(rest[0]) && rest[0] != '_' {
			continue
		}

		return typ
	}

	return ""
}

func (g *generator) defaultOpts(opts []types.Object) []types.Object {
	var res []types.Object
	for _, opt := range opts {
		switch opt.(type) {
		case *types.Const, *types.Func:
			res = append(res, opt)
		}
	}

	return res
}

func (g *generator) extractComments(opt types.Object) []string {
	rawfile := g.source.Fset.File(opt.Pos())

	var file *ast.File
	for _, syntax := range g.source.Syntax {
		srawfile := g.source.Fset.File(syntax.Pos())
		if srawfile == rawfile {
			file = syntax
			break
		}
	}

	if file == nil {
		panic(errors.New("no syntax match found for " + rawfile.Name()))
	}

	var group *ast.CommentGroup
	ast.Inspect(file, func(node ast.Node) bool {
		if node == nil {
			return false
		}

		if g.source.Fset.Position(node.Pos()).Line == g.source.Fset.Position(opt.Pos()).Line {
			switch v := node.(type) {
			case *ast.GenDecl:
				group = v.Doc
			case *ast.FuncDecl:
				group = v.Doc
			}

			return false
		}

		return true
	})

	if group == nil {
		return nil
	}

	var res []string
	var offs int
	for i, comment := range group.List {
		var cmt string
		if i == 0 {
			cmt = strings.TrimLeft(comment.Text, "/ ")
			offs = len(comment.Text) - len(cmt)
			if i == 0 {
				cmt = strings.TrimPrefix(cmt, opt.Name())
				cmt = strings.TrimLeft(cmt, " ")
			}
		} else {
			cmt = comment.Text[offs:]
		}
		res = append(res, cmt)
	}

	return res
}
