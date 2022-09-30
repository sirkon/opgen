# opgen

Option builder generator.

## Installation.

```shell
go get github.com/sirkon/main@latest
```

## Rationale.

There are so-called functional options described nicely in this [article](https://golang.cafe/blog/golang-functional-options-pattern.html).

But imagine a situation where we have a custom buffered io package where both reader and writer need a way to set up
a buffer size and there will be also options that are reader and writer specific. Something like:


```go
r, err := cio.NewBufferedReader(
	r, 
	cio.WithBufferSize(size), 
	cio.WithStartPosition(pos), 
	cio.WithMaxReadPosition(readend), // This is reader specific options. 
)

...

w, err := cio.NewBufferedWriter(w, cio.WithBufferSize(size))
```

This code will not compile because options for reader and writer will have different type in order to prevent
`cio.WithMaxReadPosition` usage for writer.

There are ways to overcome this using generics: [article](https://golang.design/research/generic-option/)

But in the end the usage will be not exactly nice – shared options to have mandatory type parameter:

```go
r, err := cio.NewBufferedReader(r, cio.WithBufferSize[*BufferedReader](size))
```
.

This utility provides a different way to functional options: we don't set options directly, we fulfill them with a 
generated builder:

```go
r, err := cio.NewBufferedReader(r, cio.BufferedReaderOptions().BufferSize(size))
```

where builders for reader and writer will have different types and therefore no type clashes.

## How to use.

Option builder generation requires a package with Go code where you set up options for each type.

Imagine we have `github.com/company/cio` repository with `BufferedReader` type in it. Then we create a package
`github.com/company/cio/internal/options` with a set of files in it.

```go
// github.com/company/cio/internal/options/buffered_reader_options.go
package options

import (
	"log"
)

// BufferedReaderBufferSize sets a buffer size, a default value is 4096.
const BufferedReaderBufferSize int = 4096

// BufferedReaderLogger sets an error logger for a buffered reader.
func BufferedReaderLogger(err error) {
	log.Println(err)
}
```

Now, generate a builder:

```shell
opgen -s internal/options -d options_gen.go BufferedReader
```

It will look for `BufferedReaderXXX` named constants (will provide a default value), variables (no default value) and
functions (will provide a default value – the function itself) where a type of each option will match a type of constant/variable/function.

The builder will have methods in case of the file above:

```go
func (b BufferedReaderOptionsType) BufferSize(v int) BufferedReaderOptionsType { … }
func (b BufferedReaderOptionsType) Logger(v func(error)) BufferedReaderOptionsType { … }
```

And the usage is supposed to be like this:

```go
v, err := cio.NewBufferedReader(
    r, 
    cio.BufferedReaderOptions().
        BufferSize(8192).
        Logger(func (err error) {
            fmt.Println(err)
        }), 
)
```
.
