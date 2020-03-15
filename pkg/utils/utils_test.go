package utils

import (
	"fmt"
	"testing"
)

func TestParseIPStr(t *testing.T) {
	var ipStr []string
	ipStr = []string{"10.0.0.1-3", "10.0.0.1-10.0.0.3", "10.0.0.1;10.0.0.2;10.0.0.3"}
	for _, str := range ipStr {
		ip, err := ParseIPStr(str)
		if err != nil {
			fmt.Println(err)
		}
		fmt.Println(ip)
		if ip[0] != "10.0.0.1" {
			t.Errorf("error")
		}
		if len(ip) != 3 {
			t.Errorf("error")
		}
		//	if ip[2] != "10.0.0.3" {
		//	t.Errorf("error")
		//}
	}
}

func TestRemoveDuplicateString(t *testing.T) {
	var ip []string
	ip = []string{"10.0.0.1", "10.0.0.2", "10.0.0.3", "10.0.0.1"}
	ip = RemoveDuplicateString(ip)
	fmt.Println(ip)
	if ip[0] != "10.0.0.1" {
		t.Errorf("error")
	}
	if len(ip) != 3 {
		t.Errorf("error")
	}
}
func TestParseIP(t *testing.T) {
	ipFile := ""
	ipStr := "10.0.0.1-3;10.0.0.1;10.0.0.2-3"
	ip, _ := ParseIP(ipFile, ipStr)
	fmt.Println(ip)
	if ip[0] != "10.0.0.1" {
		t.Errorf("error")
	}
	if len(ip) != 3 {
		t.Errorf("error")
	}
	//	if ip[2] != "10.0.0.3" {
	//	t.Errorf("error")
	//}
}
