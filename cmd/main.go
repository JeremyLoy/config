package main

import (
	"fmt"

	"github.com/JeremyLoy/config"
)

type QConfig struct {
	R string
	S bool
}

type MyInt int

type MyConfig struct {
	A MyInt
	B int8
	C int16
	D int32
	E int64
	F string
	G bool
	H float32
	I float64
	J uint
	K uint8
	L uint16
	M uint32
	N uint64
	O struct {
		P int
		Q QConfig
	}
	R *int
	T *struct {
		U int
		V *int
	}
	// TODO slice
}

func main() {
	var c MyConfig

	config.From("defaults").FromEnv().To(&c)

	fmt.Printf("%+v\n", c)
	fmt.Println(*c.T)
	fmt.Println(*c.T.V)

}
