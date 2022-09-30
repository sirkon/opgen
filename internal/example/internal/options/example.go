package options

import "log"

// ExampleName sets example name.
const ExampleName string = "example"

// ExampleLogger sets a logger for an example.
func ExampleLogger(err error) {
	log.Println(err)
}

var ExampleSize int64
