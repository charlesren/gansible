package main

import (
	"fmt"
	"gansible/src/autologin"
	"golang.org/x/crypto/ssh"
	"log"
)

func main() {
	fmt.Println("hello")
	passwords := []string{"abc", "passw0rd"}
	var client *ssh.Client
	var err error
	for _, password := range passwords {
		if client, err = autologin.Connect("root", password, "127.0.0.1", 22); err == nil {
			break
		}
	}
	defer client.Close()
	session, err := client.NewSession()
	if err != nil {
		log.Fatal(err)
	}
	defer session.Close()
	cmd := "date >> /tmp/1"
	session.Run(cmd)
	fmt.Println("end")
}
