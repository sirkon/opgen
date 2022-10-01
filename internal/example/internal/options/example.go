package options

import "log"

// ExampleName sets example name.
const ExampleName string = "example"

// ExampleUntypedWillNotWork this will be commented out.
// const ExampleUntypedWillNotWork = 12

// ExampleLogger sets a logger for an example.
func ExampleLogger(err error) {
	log.Println(err)
}

var ExampleSize int64
