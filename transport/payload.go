package transport

type Source int

const (
	Stdin Source = iota
	Stdout
	Stderr
)
