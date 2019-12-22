package main

import (
	"fmt"
	"gansible/src/autologin"
	"log"
)

func main() {
	fmt.Println("hello")
	client, err := autologin.Connect("root", "zzb11zzb", "127.0.0.1",  22)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(client)
}
