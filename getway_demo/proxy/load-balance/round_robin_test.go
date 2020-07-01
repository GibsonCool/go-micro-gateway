package load_balance

import (
	"fmt"
	"testing"
)

func TestRoundRobinBalance(t *testing.T) {
	rrb := &RoundRobinBalance{}

	rrb.Add(
		"127.0.0.1:2001",
		"127.0.0.1:2002",
		"127.0.0.1:2003",
		"127.0.0.1:2004",
		"127.0.0.1:2005",
		"1.1.1.1:9090",
	)
	rrb.Add("23424.2342.3242.545")

	for i := 0; i < 10; i++ {
		fmt.Println(rrb.Next())
	}
}
