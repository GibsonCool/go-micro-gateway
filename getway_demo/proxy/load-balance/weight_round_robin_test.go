package load_balance

import (
	"fmt"
	"testing"
)

func TestWeightRoundRobinBalance(t *testing.T) {
	rrb := &WeightRoundRobinBalance{}

	rrb.Add("127.0.0.1:2001", "2")
	rrb.Add("127.0.0.1:2002", "3")
	rrb.Add("127.0.0.1:2003", "4")

	for i := 0; i < 19; i++ {
		fmt.Println(rrb.Next())
	}
}
