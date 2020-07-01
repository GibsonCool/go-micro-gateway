package load_balance

import (
	"fmt"
	"testing"
)

func TestRandomBalance(t *testing.T) {
	rb := &RandomBalance{}
	rb.Add("127.0.0.1:2001",
		"127.0.0.1:2002",
		"127.0.0.1:2003",
		"127.0.0.1:2004",
		"127.0.0.1:2005",
	)
	rb.Add("192.168.1.1:3030")

	for i := 0; i < 20; i++ {
		fmt.Println(rb.Next())
	}
}
