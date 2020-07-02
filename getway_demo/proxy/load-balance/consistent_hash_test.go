package load_balance

import (
	"fmt"
	"testing"
)

func TestConsistentHashBalance(t *testing.T) {
	chb := NewConsistentHashBalance(nil, 10)
	chb.Add("127.0.0.1:2001",
		"127.0.0.1:2002",
		"127.0.0.1:2003",
		"127.0.0.1:2004",
		"127.0.0.1:2005")

	//url hash
	fmt.Println(chb.Get("http://127.0.0.1:2002/base/getinfo"))
	fmt.Println(chb.Get("http://127.0.0.1:2002/base/error"))
	fmt.Println(chb.Get("http://127.0.0.1:2002/base/getinfo"))
	fmt.Println(chb.Get("http://127.0.0.1:2002/base/changepwd"))

	//ip hash
	fmt.Println(chb.Get("127.0.0.1"))
	fmt.Println(chb.Get("192.168.0.1"))
	fmt.Println(chb.Get("127.0.0.1"))

}
