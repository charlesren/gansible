package utils

import (
	"fmt"
	"testing"
)

func TestParseIPStr(t *testing.T) {
	var ipStr []string
	ipStr = []string{"10.0.0.1-3", "10.0.0.1-10.0.0.3"}
	for _, str := range ipStr {
		ip, err := ParseIPStr(str)
		if err != nil {
			fmt.Println(err)
		}
		if ip[0] != "10.0.0.1" {
			t.Errorf("error")
		}
		if ip[2] != "10.0.0.3" {
			t.Errorf("error")
		}
	}
}
