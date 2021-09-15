package main

import (
	"fmt"
	"github.com/gammazero/workerpool"
	"github.com/tmyers273/go-training/dice"
	"sync"
	"time"
)

const NumberOfRolls = 100

func main() {
	fmt.Println("Welcome to go training! Be sure to check out the individual methods for comments and explanations.")
	fmt.Println("Starting things now...")

	rollUsingLoop()
	rollUsingWaitGroups()

	sumRollsUsingLoop()
	sumRollsUsingBufferedChannel()
	sumRollsUsingUnbufferedChannel()

	sumRollsUsingConcurrencyLimit()
}

func rollUsingLoop() {
	// If we want to perform 100 dice rolls, the naive implementation would be to do a for loop
	start := time.Now()
	for i:=0; i<NumberOfRolls;i++ {
		dice.Roll()
	}
	fmt.Printf("Took %s to do %d dice rolls using a naive loop\n", time.Since(start), NumberOfRolls)

	// But, if we were to assume that dice.Roll is calling some API endpoint with a response time
	// and if we are looking to make this as fast as possible, then the naive loop implementation
	// will be slow.

	// If the roll takes 100ms, doing 100 rolls in a loop will take 100ms * 100 = 10,000 ms = 10s
	// If we were to do 1,000 rolls, then our time increases linearly to 100s
}

func rollUsingWaitGroups() {
	// Enter: goroutines!
	// goroutines allow us to run things in parallel. The easiest to start with is probably wait
	// groups. Waitgroups basically give you a primitive to keep track of the number of running
	// things, or goroutines in this case.

	start := time.Now()
	wg := sync.WaitGroup{}
	wg.Add(NumberOfRolls)
	for i:=0; i<NumberOfRolls;i++ {
		go func() {
			dice.Roll()
			wg.Done()
		}()
	}
	wg.Wait()
	fmt.Printf("Took %s to do %d dice rolls using a waitgroup\n", time.Since(start), NumberOfRolls)

	// The above code will create 100 goroutines, each performing the dice.Roll once. This should
	// give us a total response time of ~100ms for ALL 100 dice rolls!
}

func sumRollsUsingLoop() {
	// Things get a bit trickier if we care about the return values of the called methods. If we
	// wanted to take the sum of all 100 dice rolls, using a naive loop that would look like this

	sum := 0
	start := time.Now()
	for i:=0; i<NumberOfRolls;i++ {
		sum += dice.Roll()
	}
	fmt.Printf("Took %s to sum %d dice rolls using a naive loop. Sum is %d\n", time.Since(start), NumberOfRolls, sum)
}

func sumRollsUsingBufferedChannel() {
	// Channels are generally used to communicate and share data between goroutines. They are
	// perfect for our use case here! The 10s run time of the above code is far too long. Let's
	// see if we can do any better...

	sum := 0
	ch := make(chan int, NumberOfRolls) // <-- using a buffered channel
	wg := sync.WaitGroup{}
	wg.Add(NumberOfRolls)
	start := time.Now()
	for i:=0; i<NumberOfRolls;i++ {
		go func() {
			ch <- dice.Roll()
			wg.Done()
		}()
	}

	// Since the channel is buffered and has a max length of 100, we can push all of our results onto the channel,
	// _then_ start reading the values from the channel _after_ they have all been populated.

	wg.Wait()
	close(ch)

	for roll := range ch {
		sum += roll
	}

	fmt.Printf("Took %s to sum %d dice rolls using a buffered channel. Sum is %d\n", time.Since(start), NumberOfRolls, sum)

	// Buffered channels can be great, but they come at a cost. Whenever you created a buffered channel you are
	// allocating memory equal to your buffer length * the size of the value in the channel. For 100 ints this
	// isn't a bit deal, but if you had 10 million BIG structs, you could be chewing up a LOT of memory unnecessarily.
}

func sumRollsUsingUnbufferedChannel() {
	sum := 0
	ch := make(chan int) // <-- using an unbuffered channel
	start := time.Now()
	go func() { // <-- notice the extra goroutine here (1)
		for i:=0; i<NumberOfRolls;i++ {
			go func() {
				ch <- dice.Roll()
			}()
		}
	}()

	// Since the channel is unbuffered, it has an effective max length of 1. So we
	// need to be reading from the channel at the same time we are writing to it.
	//
	// The extra goroutine (1) gives us a way to be writing to
	// the channel at the same time we are reading from it (2)

	// Read from the channel (2)
	for i:=0;i<NumberOfRolls;i++ {
		sum += <- ch
	}
	close(ch)

	fmt.Printf("Took %s to sum %d dice rolls using an unbuffered channel. Sum is %d\n", time.Since(start), NumberOfRolls, sum)

	// Unbuffered channels allow us a much more minimal memory footprint, at the cost of a slightly higher complexity.
}

func sumRollsUsingConcurrencyLimit() {
	// Concurrency is great and can _really_ help the performance of different applications. But, it can also be
	// easily to overload the target system. In this particular case, a dice roll is a very easy, lightweight
	// computation. But if you were calling an API endpoint that had some heavy calculations or a throttling
	// quotas in some way, you need a way of controlling how many concurrent things you are calling.

	// Let's assume we are working with a dice rolling service that only allows you to do 10 concurrent
	// dice rolls at once. How would you go about implementing a system that allows you to do no more than
	// 10 concurrent dice rolls at once?

	// Question: if you are doing 100 rolls, and it takes 100ms for each roll, with a maximum level of
	// concurrency of 10, how long would you expect this to take?

	// Answer: if you said 1s, you would be right! Each worker, on average, should process 10 tasks.
	// 100 rolls (or tasks) / the number of workers (10) = 10 rolls (or tasks) per worker, on average
	// If each roll (or task) takes 100ms, then each worker will take approximately 1s to complete.
	// Since the 10 workers are all working in parallel, we should get the result of all 100 rolls in ~1s.

	// This concept is generally referred to as "worker pools". While you can certainly implement these on your
	// own, I greatly prefer to reach for a package for these. https://github.com/gammazero/workerpool is my
	// favorite of the bunch.

	// Create a new worker pool limited to 10 concurrent tasks
	start := time.Now()
	concurrencyLimit := 10
	wp := workerpool.New(concurrencyLimit)

	ch := make(chan int)
	for i:=0;i<NumberOfRolls;i++ {
		// The worker pool Submit method does _not_ block, so we don't need any goroutines here
		wp.Submit(func() {
			// We'll perform a dice roll and push the results onto an unbuffered channel
			ch <- dice.Roll()
		})
	}

	// This is telling the worker pool to wait until all the current tasks in the queue are completed.
	// Afterwards, we close the channel, allowing us to perform a range on it and have the range exit.
	go func() {
		wp.StopWait()
		close(ch)
	}()

	sum := 0
	for roll := range ch {
		sum += roll
	}

	fmt.Printf("Took %s to sum %d dice rolls using an unbuffered channel and a concurrency limit of %d. Sum is %d\n", time.Since(start), NumberOfRolls, concurrencyLimit, sum)
}