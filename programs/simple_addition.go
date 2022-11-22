package main

import (
	"fmt"

	tx "github.com/last/transactionalVariable"
)

func main() {
	// fmt.Println("starting program")
	x := tx.New("x", 2)
	y := tx.New("y", 3)

	// fmt.Println("Getting variable")
	fmt.Printf("variable %s has value %d\n", x.Name(), x.Get())
	fmt.Printf("variable %s has value %d\n", y.Name(), y.Get())

	// z := tx.New("z", x.Get()+y.Get())
	// fmt.Printf("variable %s has value %d\n", z.Name(), z.Get())

	// run multiple threads and increment variable x. Should add up.
	for i := 0; i < 10; i++ {
		go func() {
			x.Set(x.Get() + 1)
		}()
	}
}
