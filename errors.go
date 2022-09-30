package main

import "github.com/sirkon/errors"

func failureError() error {
	return errors.New("failed to proceed")
}
