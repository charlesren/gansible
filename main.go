package main

import (
	"gansible/src/autologin"
	"log"
	"os"

	"golang.org/x/crypto/ssh"
)

func main() {
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
	/*
		cmd := "date"
		out, err := session.CombinedOutput(cmd)
		if err != nil {
			log.Fatal("Remote Exec Field:", err)
		}
		fmt.Println("Remote Exec Output:\n", string(out))
	*/
	session.Stdout = os.Stdout
	session.Stderr = os.Stderr
	session.Stdin = os.Stdin
	modes := ssh.TerminalModes{
		ssh.ECHO:          0,     // disable echoing
		ssh.TTY_OP_ISPEED: 14400, // input speed = 14.4kbaud
		ssh.TTY_OP_OSPEED: 14400, // output speed = 14.4kbaud
	}
	err = session.RequestPty("xterm-256color", 40, 80, modes)
	if err != nil {
		log.Fatal(err)
	}
	err = session.Shell()
	err = session.Wait()
}
