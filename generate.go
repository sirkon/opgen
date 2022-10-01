package main

import (
	"go/token"
	"go/types"
	"sort"
	"strings"

	"github.com/sirkon/errors"
	"github.com/sirkon/gogh"
	"github.com/sirkon/message"
	"github.com/sirkon/opgen/internal/app"
	"golang.org/x/tools/go/packages"
)

func generate(sourcePackage, dest string, types []string) error {
	sort.Strings(types)

	g, err := newGenerator(sourcePackage, dest, types)
	if err != nil {
		return errors.Wrap(err, "setup generator")
	}

	if err := g.render(); err != nil {
		return errors.Wrap(err, "run code generation")
	}

	return nil
}

type generator struct {
	source *packages.Package
	dest   string
	types  []string

	mapping map[string][]types.Object

	mod *gogh.Module[*gogh.Imports]
}

func newGenerator(source string, dest string, typs []string) (*generator, error) {
	res := &generator{
		dest:    dest,
		types:   typs,
		mapping: map[string][]types.Object{},
	}

	parsingMode := packages.NeedImports | packages.NeedTypes | packages.NeedName |
		packages.NeedDeps | packages.NeedSyntax | packages.NeedFiles | packages.NeedModule

	pkgs, err := packages.Load(
		&packages.Config{
			Mode:      parsingMode,
			Fset:      token.NewFileSet(),
			ParseFile: nil,
		},
		source,
	)
	if err != nil {
		return nil, errors.Wrapf(err, "load source package '%s'", source)
	}

	var pkg *packages.Package
	switch len(pkgs) {
	case 0:
		return nil, errors.Newf("no package '%s' found", source)
	case 1:
		pkg = pkgs[0]
	default:
		for _, p := range pkgs {
			if strings.HasSuffix(p.PkgPath, "/"+source) {
				pkg = p
				break
			}
		}

		if pkg == nil {
			return nil, errors.Newf("no package '%s' found", source)
		}
	}
	res.source = pkg

	m, err := gogh.New(
		gogh.FancyFmt,
		func(r *gogh.Imports) *gogh.Imports {
			return r
		},
	)
	if err != nil {
		return nil, errors.Wrap(err, "init code generator")
	}

	res.mod = m

	if err := res.mapOptions(); err != nil {
		return nil, errors.Wrapf(err, "map options items to types")
	}

	var failed bool
	for _, typ := range typs {
		opts := res.mapping[typ]
		if len(opts) == 0 {
			message.Errorf("no options found for %s", typ)
			failed = true
		}
	}
	if failed {
		return nil, failureError()
	}

	return res, nil
}

func (g *generator) render() error {
	pkg, err := g.mod.Current("")
	if err != nil {
		return errors.Wrap(err, "setup current package")
	}

	r := pkg.Go(g.dest, gogh.Autogen(app.Name))
	sort.Strings(g.types)
	for _, typName := range g.types {
		g.renderBuilder(r, typName)
	}

	if err := g.mod.Render(); err != nil {
		return errors.Wrap(err, "save source code")
	}

	return nil
}

func (g *generator) renderBuilder(r *gogh.GoRenderer[*gogh.Imports], typ string) {
	opts := g.mapping[typ]

	r.N()
	r.L(`// $0OptionsType for type $0`, typ)
	r.L(`type $0OptionsType struct{`, typ)
	r.L(`    opts []func(v *$0)`, typ)
	r.L(`}`)
	r.N()
	r.L(`// $0Options options builder constructor for $0`, typ)
	r.L(`func $0Options() $0OptionsType {`, typ)
	r.L(`    res := $0OptionsType{}`, typ)
	defaults := g.defaultOpts(opts)
	if len(defaults) > 0 {
		r.L(`    res.opts = make([]func(v *$0), 0, $1)`, typ, len(defaults))
		for _, opt := range defaults {
			g.renderDefaultApplication(r, opt, optName(opt, typ))
		}
	}
	r.L(`    return res`)
	r.L(`}`)
	r.N()

	for _, opt := range opts {
		g.renderSetMethod(r, typ, opt)
	}

	r.N()
	r.L(`func (o $0OptionsType) apply(vv *$0) {`, typ)
	r.L(`    for _, opt := range o.opts {`)
	r.L(`        opt(vv)`)
	r.L(`    }`)
	r.L(`}`)
}

func (g *generator) renderSetMethod(r *gogh.GoRenderer[*gogh.Imports], typ string, opt types.Object) {
	name := optName(opt, typ)

	cmts := g.extractComments(opt)

	r.N()
	for i, cmt := range cmts {
		if i == 0 {
			r.L(`// $0 $1`, optName(opt, typ), cmt)
		} else {
			r.L(`// $0`, cmt)
		}
	}
	r.L(`func (o $0OptionsType) $1(v $2) $0OptionsType {`, typ, name, r.Type(opt.Type()))
	r.L(`    o.opts = append(o.opts, func(vv *$0) {`, typ)
	r.L(`        vv.set$0(v)`, name)
	r.L(`    })`)
	r.L(`    return o`)
	r.L(`}`)
}

func (g *generator) renderDefaultApplication(r *gogh.GoRenderer[*gogh.Imports], opt types.Object, name string) {
	r.Imports().Add(opt.Pkg().Path()).Ref("optpkg")
	r.L(`    res = res.$0(${optpkg}.$1)`, name, opt.Name())
}

func optName(opt types.Object, typ string) string {
	return strings.TrimPrefix(opt.Name(), typ)
}
