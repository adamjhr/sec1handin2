package util

import (
	"fmt"
	"math/rand"
)

func StringifyPort(port int) string {
	return fmt.Sprintf(":%d", port)
}

func CreateShares(r int, data int, amount int) []int {
	var shares []int
	var totalShares int

	for i := 0; i < amount; i++ {
		share := rand.Intn(r-1) + 1
		shares = append(shares, share)
		totalShares += share
	}

	shares = append(shares, (data-totalShares)%r)
	return shares
}
