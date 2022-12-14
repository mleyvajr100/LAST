package main

import (
	"fmt"
	"sync"

	tx "github.com/last/transactionalVariable"
)

// func inc(x *tx.TxVar) {
// 	fmt.Println("incrementing x by 1")
// 	s := tx.CreateSession()
// 	x.Set(x.Get(s)+1, s)
// 	tx.CommitSession(s)
// }

func main() {
	// fmt.Println("starting program")
	// session := tx.CreateSession()

	s := tx.CreateSession()

	fmt.Println("Got session")

	x := tx.New("x", 1, s)
	fmt.Println("MADE IT")
	newVal := x.Get(s) + 1
	fmt.Println("MADE IT")
	x.Set(newVal, s)
	fmt.Println("MADE IT")
	tx.CommitSession(s)
	fmt.Println("MADE IT")
	// y := tx.New("y", 2, s)

	// // fmt.Println("Getting variable")
	// fmt.Printf("variable %s has value %d\n", x.Name(), x.Get(s))
	// fmt.Printf("variable %s has value %d\n", y.Name(), y.Get(s))

	// z := tx.New("z", x.Get(s)+y.Get(s), s)
	// fmt.Printf("variable %s has value %d\n", z.Name(), z.Get(s))

	// tx.CommitSession(s)

	count := 0

	var wg sync.WaitGroup
	// run multiple threads and increment variable x. Should add up.
	for i := 0; i < 2; i++ {
		wg.Add(1)
		// time.Sleep(1 * time.Second)
		go func() {
			defer wg.Done()
			s := tx.CreateSession()
			x.Set(x.Get(s)+1, s)
			tx.CommitSession(s)
			count += 1
		}()
	}

	wg.Wait()

	fmt.Println(count)

	// // run multiple threads and increment variable x. Should add up.
	// for i := 0; i < 10; i++ {
	// 	go atomic([x], func() {
	// 		x.Set(x.Get() + 1)
	// 	}())
	// }
}
