package main

import (
	"go/ast"
	"go/types"
	"strings"
	"unicode"

	"github.com/sirkon/errors"
)

func (g *generator) mapOptions() {
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

	for _, name := range names {
		item := g.source.Types.Scope().Lookup(name)
		switch item.(type) {
		case *types.Const, *types.Var, *types.Func:
		default:
			continue
		}

		if owner := g.owner(item.Name()); owner != "" {
			g.mapping[owner] = append(g.mapping[owner], item)
		}
	}
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
	for i, comment := range group.List {
		cmt := strings.TrimLeft(comment.Text, "/ ")
		if i == 0 {
			cmt = strings.TrimPrefix(cmt, opt.Name())
			cmt = strings.TrimLeft(cmt, " ")
		}
		res = append(res, cmt)
	}

	return res
}
