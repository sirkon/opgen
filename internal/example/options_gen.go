// Code generated by opgen version (devel). DO NOT EDIT.

package example

import "github.com/sirkon/opgen/internal/example/internal/options"

// ExampleOptionsType for type Example
type ExampleOptionsType struct {
	opts []func(v *Example)
}

// ExampleOptions options builder constructor for Example
func ExampleOptions() ExampleOptionsType {
	res := ExampleOptionsType{}
	res.opts = make([]func(v *Example), 0, 2)
	res = res.Name(options.ExampleName)
	res = res.Logger(options.ExampleLogger)
	return res
}

// Example sets example name.
func (o ExampleOptionsType) Name(v string) ExampleOptionsType {
	o.opts = append(o.opts, func(vv *Example) {
		vv.setName(v)
	})
	return o
}

// Example sets a logger for an example.
func (o ExampleOptionsType) Logger(v func(err error)) ExampleOptionsType {
	o.opts = append(o.opts, func(vv *Example) {
		vv.setLogger(v)
	})
	return o
}

func (o ExampleOptionsType) Size(v int64) ExampleOptionsType {
	o.opts = append(o.opts, func(vv *Example) {
		vv.setSize(v)
	})
	return o
}
