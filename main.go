package main

import (
	"fmt"
	"github.com/tmyers273/go-training/dice"
	"sync"
	"time"
)

func main() {
	var start time.Time

	fmt.Println("Welcome to go training! Starting things now...")

	// If we want to perform 100 dice rolls, the naive implementation would be to do a for loop
	start = time.Now()
	for i:=0; i<100;i++ {
		dice.Roll()
	}
	fmt.Printf("Took %s to do 100 dice rolls using a naive loop\n", time.Since(start))

	// But, if we were to assume that dice.Roll is calling some API endpoint with a response time
	// and if we are looking to make this as fast as possible, then the naive loop implementation
	// will be slow.

	// If the roll takes 100ms, doing 100 rolls in a loop will take 100ms * 100 = 10,000 ms = 10s

	// Enter: goroutines!
	// goroutines allow us to run things in parallel. The easiest to start with is probably wait
	// groups. Waitgroups basically give you a primitive to keep track of the number of running
	// things, or goroutines in this case.

	start = time.Now()
	wg := sync.WaitGroup{}
	wg.Add(100)
	for i:=0; i<100;i++ {
		go func() {
			dice.Roll()
			wg.Done()
		}()
	}
	wg.Wait()
	fmt.Printf("Took %s to do 100 dice rolls using a waitgroup\n", time.Since(start))

	// The above code will create 100 goroutines, each performing the dice.Roll once. This should
	// give us a total response time of ~100ms for ALL 100 dice rolls!

	// Things get a bit trickier if we care about the return values of the called methods. If we
	// wanted to take the sum of all 100 dice rolls, using a naive loop that would look like this

	sum := 0
	start = time.Now()
	for i:=0; i<100;i++ {
		sum += dice.Roll()
	}
	fmt.Printf("Took %s to sum 100 dice rolls using a naive loop. Sum is %d\n", time.Since(start), sum)

	// Channels are generally used to communicate and share data between goroutines. They are
	// perfect for our use case here! The 10s run time of the above code is far too long. Let's
	// see if we can do any better...

	sum = 0
	ch := make(chan int, 100)
	wg = sync.WaitGroup{}
	wg.Add(100)
	start = time.Now()
	go func() {
		for i:=0; i<100;i++ {
			go func() {
				ch <- dice.Roll()
				wg.Done()
			}()
		}
	}()

	wg.Wait()
	close(ch)

	for roll := range ch {
		sum += roll
	}

	fmt.Printf("Took %s to sum 100 dice rolls using a naive loop. Sum is %d\n", time.Since(start), sum)
}
