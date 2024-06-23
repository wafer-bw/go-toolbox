package always_test

import (
	"errors"
	"fmt"

	"github.com/wafer-bw/go-toolbox/always"
)

func ExampleMust_panics() {
	always.Must(func() (string, error) { return "", errors.New("oh no") }())
}

func ExampleMust_notPanics() {
	v := always.Must(func() (string, error) { return "hello", nil }())
	fmt.Println(v)
	// Output:
	// hello
}

func ExampleMustDo_panics() {
	always.MustDo(func() error { return errors.New("oh no") }())
}

func ExampleMustDo_notPanics() {
	always.MustDo(func() error { return nil }())
}

func ExampleAccept() {
	v := always.Accept(func() (string, bool) { return "hello", false }())
	fmt.Println(v)
	// Output:
	// hello
}
