package main

import (
	"fmt"
	"gansible/src/autologin"
	"log"
)

func main() {
	fmt.Println("hello")
	chiperList := []string{}
	client, err := autologin.Connect("rencc", "zzb11zzb", "127.0.0.1", "", 22, chiperList)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(client)
}
