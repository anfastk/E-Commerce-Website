package helper

import (
	"fmt"
	"math/rand"
	"strings"
	"time"
)

func GenerateOrderID() string {
	rand.Seed(time.Now().UnixNano())
	section1 := rand.Intn(900) + 100
	section2 := rand.Intn(9000000) + 1000000
	section3 := time.Now().UnixNano() % 10000000

	return strings.ToUpper(fmt.Sprintf("#%d-%d-%07d", section1, section2, section3))
}
 