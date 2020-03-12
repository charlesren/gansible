package autologin

import (
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"os"
	"time"

	homedir "github.com/mitchellh/go-homedir"
	"golang.org/x/crypto/ssh"
	"golang.org/x/crypto/ssh/agent"
)

// Connect func
func Connect(user string, password string, host string, port int) (*ssh.Client, error) {
	var (
		auth         []ssh.AuthMethod
		addr         string
		clientConfig *ssh.ClientConfig
		client       *ssh.Client
		config       ssh.Config
		err          error
	)
	auth = make([]ssh.AuthMethod, 0)
	auth = append(auth, ssh.Password(password))
	clientConfig = &ssh.ClientConfig{
		User:            user,
		Auth:            auth,
		Timeout:         30 * time.Second,
		Config:          config,
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
	}
	addr = fmt.Sprintf("%s:%d", host, port)
	if client, err = ssh.Dial("tcp", addr, clientConfig); err != nil {
		return nil, err
	}
	return client, nil
}
func publicKeyAuth(keyPath string) ssh.AuthMethod {
	keyFile, err := homedir.Expand(keyPath)
	if err != nil {
		log.Fatal("find key's home dir failed", err)
	}
	key, err := ioutil.ReadFile(keyFile)
	if err != nil {
		log.Fatal("read ssh key file failed", err)
	}
	signer, err := ssh.ParsePrivateKey(key)
	if err != nil {
		log.Fatal("get signer failed", err)
	}
	return ssh.PublicKeys(signer)
}

func publicKeyWithPasswordAuth(keyPath string, keyPassword string) ssh.AuthMethod {
	keyFile, err := homedir.Expand(keyPath)
	if err != nil {
		log.Fatal("find key's home dir failed", err)
	}
	key, err := ioutil.ReadFile(keyFile)
	if err != nil {
		log.Fatal("read ssh key file failed", err)
	}
	signer, err := ssh.ParsePrivateKeyWithPassphrase(key, []byte(keyPassword))
	if err != nil {
		log.Fatal("get signer failed", err)
	}
	return ssh.PublicKeys(signer)
}

func publicKeyWithSSHAgentAuth() ssh.AuthMethod {
	socket := os.Getenv("SSH_AUTH_SOCK")
	conn, err := net.Dial("unix", socket)
	if err != nil {
		log.Fatalf("Failed to open SSH_AUTH_SOCK: %v", err)
	}
	agentClient := agent.NewClient(conn)
	return ssh.PublicKeysCallback(agentClient.Signers)
}

func getPublicKeyAuthMethod(keyPath string, keyPassword string, password string) ssh.AuthMethod {
	if password != "" {
		return ssh.Password(password)
	} else if keyPath != "" {
		if keyPassword != "" {
			return publicKeyWithPasswordAuth(keyPath, keyPassword)
		}
		return publicKeyAuth(keyPath)
	} else if socket := os.Getenv("SSH_AUTH_SOCK"); socket != "" {
		return publicKeyWithSSHAgentAuth()
	} else {
		return publicKeyAuth(keyPath)
	}
}
