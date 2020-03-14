package connect

import (
	"fmt"
	"testing"
)

func TestDo(t *testing.T) {
	var (
		keyPath     string
		keyPassword string
		user        string
		password    string
		host        string
		port        int
	)
	host = "127.0.0.1"
	port = 22
	user = "root"
	if true {
		fmt.Println("password auth test...")
		password = "yourpassword"
		_, err := Do(keyPath, keyPassword, user, password, host, port, 30)
		if err != nil {
			t.Errorf("user password login failed: %s", err)
		}
		password = ""
	}

	if true {
		fmt.Println("ssh key auth test...")
		keyPath = "~/.ssh/id_rsa"
		_, err := Do(keyPath, keyPassword, user, password, host, port, 30)
		if err != nil {
			t.Errorf("user login failed: %s", err)
		}
		keyPath = ""
	}

	/*
		if true {
			fmt.Println("ssh key with keypass auth test...")
			keyPath = "~/.ssh/id_rsa"
			keyPassword = "yourkeypass"
			_, err := Do(keyPath, keyPassword, user, password, host, port,30)
			if err != nil {
				t.Errorf("user login failed: %s", err)
			}
			keyPath = ""
			keyPassword = ""
		}
	*/

	/*
		//can test
		//1.use or not use ssh-agent #depends on if ssh-agent is up
		//2.ssh key without  keypass #depends on if have ~/.ssh/id_rsa file
		//3.password auth (will return err,becase password is not set)
		if true {
			fmt.Println("no args test ")
			_, err := Do(keyPath, keyPassword, user, password, host, port,30)
			if err != nil {
				t.Errorf("user login failed: %s", err)
			}
		}
	*/
}
