package funcopts

import (
	"fmt"
)

type Option func(*T) error

func WithNumber(v int) Option {
	return func(t *T) error {
		t.number = v
		return nil
	}
}

func WithWord(w string) Option {
	return func(t *T) error {
		t.word = w
		return nil
	}
}

type T struct {
	number int
	word   string
}

func (t T) GetNumber() int {
	return t.number
}

func (t T) GetWord() string {
	return t.word
}

func NewT(opts ...Option) (*T, error) {
	t := &T{}
	
	if err := Process(t, opts...); err != nil {
		return nil, err
	}

	return t, nil
}

func ExampleProcess() {
	t, err := NewT(WithNumber(42), WithWord("hello"))
	if err != nil {
		panic(err) // Handle error appropriately.
	}

	// Output:
	// 42
	// hello
	fmt.Println(t.GetNumber())
	fmt.Println(t.GetWord())
}
