package dice

import (
	"math/rand"
	"time"
)

func Roll() int {
	time.Sleep(100 * time.Millisecond)
	return rand.Intn(5)+1
}
