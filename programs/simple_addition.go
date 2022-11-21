package main

import (
	"fmt"

	tx "github.com/last/transactionalVariable"
)

func main() {
	x := tx.New("x", 2)
	y := tx.New("y", 3)

	fmt.Printf("variable %s has value %d\n", x.Name(), x.Get())
	fmt.Printf("variable %s has value %d\n", y.Name(), y.Get())

	z := tx.New("z", x.Get()+y.Get())

	fmt.Printf("variable %s has value %d\n", z.Name(), z.Get())
}
