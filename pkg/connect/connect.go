package connect

import (
	"bufio"
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"os"
	osuser "os/user"
	"path"
	"sync"
	"time"

	"github.com/mitchellh/go-homedir"
	"golang.org/x/crypto/ssh"
	"golang.org/x/crypto/ssh/agent"
)

//Do func is used to connect to server
func Do(keyPath string, keyPassword string, user string, password string, node string, port int, sshTimeout int, pwdFile string) (*ssh.Client, error) {
	var (
		authMethod   []ssh.AuthMethod
		addr         string
		clientConfig *ssh.ClientConfig
		client       *ssh.Client
		config       ssh.Config
		err          error
	)
	if user == "" {
		currentUser, err := osuser.Current()
		if err != nil {
			panic(err)
		}
		user = currentUser.Name
	}
	authMethod = make([]ssh.AuthMethod, 0)
	clientConfig = &ssh.ClientConfig{
		User:            user,
		Auth:            authMethod,
		Timeout:         time.Duration(sshTimeout) * time.Second,
		Config:          config,
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
	}
	addr = fmt.Sprintf("%s:%d", node, port)
	auth := GetAuthMethod(keyPath, keyPassword, password)
	if auth == nil {
		//fmt.Println("try given passwords")
		passwords := GetPassword(pwdFile)
		client, err := TryPasswords(user, passwords, node, port, sshTimeout)
		if err != nil {
			return nil, err
		}
		return client, nil
	}
	clientConfig.Auth = append(clientConfig.Auth, auth)
	client, err = ssh.Dial("tcp", addr, clientConfig)
	if err != nil {
		//fmt.Println("private key auth failed,try given passwords")
		passwords := GetPassword(pwdFile)
		client, err := TryPasswords(user, passwords, node, port, sshTimeout)
		if err != nil {
			return nil, err
		}
		return client, nil
	}
	return client, nil
}

//PublicKeyAuth func return ssh.AuthMethod
func PublicKeyAuth(keyPath string) ssh.AuthMethod {
	if keyPath == "" {
		keyPath = "~/.ssh/id_rsa"
	}
	keyFile, err := homedir.Expand(keyPath)
	if err != nil {
		//fmt.Println("find key's home dir failed", err)
		return nil
	}
	key, err := ioutil.ReadFile(keyFile)
	if err != nil {
		//fmt.Println("read ssh key file failed", err)
		return nil
	}
	signer, err := ssh.ParsePrivateKey(key)
	if err != nil {
		//fmt.Println("get signer failed", err)
		return nil
	}
	return ssh.PublicKeys(signer)
}

//PublicKeyWithPasswordAuth func return ssh.AuthMethod
func PublicKeyWithPasswordAuth(keyPath string, keyPassword string) ssh.AuthMethod {
	if keyPath == "" {
		keyPath = "~/.ssh/id_rsa"
	}
	keyFile, err := homedir.Expand(keyPath)
	if err != nil {
		//fmt.Println("find key's home dir failed", err)
		return nil
	}
	key, err := ioutil.ReadFile(keyFile)
	if err != nil {
		//fmt.Println("read ssh key file failed", err)
		return nil
	}
	signer, err := ssh.ParsePrivateKeyWithPassphrase(key, []byte(keyPassword))
	if err != nil {
		//fmt.Println("get signer failed", err)
		return nil
	}
	return ssh.PublicKeys(signer)
}

//PublicKeyWithSSHAgentAuth func return ssh.AuthMethod
func PublicKeyWithSSHAgentAuth() ssh.AuthMethod {
	socket := os.Getenv("SSH_AUTH_SOCK")
	conn, err := net.Dial("unix", socket)
	if err != nil {
		//fmt.Println("Failed to open SSH_AUTH_SOCK:", err)
		return nil
	}
	agentClient := agent.NewClient(conn)
	return ssh.PublicKeysCallback(agentClient.Signers)
}

//GetAuthMethod func return ssh.AuthMethod.
//Only if password is assigned return ssh.Password AuthMethod,otherwise return key related AuthMethod or nil.
func GetAuthMethod(keyPath string, keyPassword string, password string) ssh.AuthMethod {
	//password is assigned
	if password != "" {
		return ssh.Password(password)
	}

	//password is not assigned; keyPath is assigned
	if keyPath != "" {
		if keyPassword != "" {
			return PublicKeyWithPasswordAuth(keyPath, keyPassword)
		}
		return PublicKeyAuth(keyPath)
	}

	//password is not assigned; keyPath is not assigned;keyPassword is assigned
	if keyPassword != "" {
		defaultKeyFile, err := homedir.Expand("~/.ssh/id_rsa")
		if err != nil {
			//fmt.Println("find default key's home dir failed: ", err)
			return nil
		}
		if _, err := os.Stat(defaultKeyFile); os.IsNotExist(err) {
			//fmt.Println("default key file is not exist: ", err)
			return nil
		}
		return PublicKeyWithPasswordAuth(defaultKeyFile, keyPassword)
	}

	//password is not assigned; keyPath is not assigned;keyPassword is not assigned
	if socket := os.Getenv("SSH_AUTH_SOCK"); socket != "" {
		return PublicKeyWithSSHAgentAuth()
	}
	defaultKeyFile, err := homedir.Expand("~/.ssh/id_rsa")
	if err != nil {
		//fmt.Println("find default key's home dir failed: ", err)
		return nil
	}
	if _, err := os.Stat(defaultKeyFile); os.IsNotExist(err) {
		//fmt.Println("default key file is not exist: ", err)
		return nil
	}
	return PublicKeyAuth(defaultKeyFile)
}

//GetPassword  parse password file and store passwords into slice
func GetPassword(pwdFile string) []string {
	passwords := []string{}
	if pwdFile == "" {
		home, err := homedir.Dir()
		if err != nil {
			fmt.Println("get homedir error:", err)
			return nil
		}
		pwdFile = path.Join(home, ".pwdfile")
	}
	file, err := os.Open(pwdFile)
	if err != nil {
		log.Printf("can not open passowrd file: %s, err: [%v]", pwdFile, err)
		return nil
	}
	defer file.Close()
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		passwords = append(passwords, scanner.Text())
	}
	if err := scanner.Err(); err != nil {
		log.Printf("Cannot scanner file: %s, err: [%v]", pwdFile, err)
		return nil
	}
	return passwords
}

//TryPasswords ssh to a machine using a set of possible passwords concurrently.
func TryPasswords(user string, passwords []string, node string, port int, sshTimeout int) (*ssh.Client, error) {
	timer := time.NewTimer(time.Duration(sshTimeout) * time.Second)
	defer timer.Stop()
	ch := make(chan *ssh.Client)
	count := 0
	var mutex sync.Mutex
	finish := make(chan bool)
	errTimeout := fmt.Errorf("Time out in %d seconds", sshTimeout)
	errAllPassWrong := fmt.Errorf("All passwords are wrong")
	for _, password := range passwords {
		go func(password string) {
			c, err := WithPass(user, password, node, port, sshTimeout)
			if err == nil {
				ch <- c
			} else {
				mutex.Lock()
				count = count + 1
				if count == len(passwords) {
					finish <- true
				}
				mutex.Unlock()
			}
		}(password)
	}
	select {
	case client := <-ch:
		return client, nil
	case <-finish:
		return nil, errAllPassWrong
	case <-timer.C:
		return nil, errTimeout
	}
}

//WithPass connect server with password
func WithPass(user string, password string, node string, port int, sshTimeout int) (*ssh.Client, error) {
	var (
		authMethod   []ssh.AuthMethod
		addr         string
		clientConfig *ssh.ClientConfig
		client       *ssh.Client
		config       ssh.Config
		err          error
	)
	authMethod = make([]ssh.AuthMethod, 0)
	authMethod = append(authMethod, ssh.Password(password))
	clientConfig = &ssh.ClientConfig{
		User:            user,
		Auth:            authMethod,
		Timeout:         time.Duration(sshTimeout) * time.Second,
		Config:          config,
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
	}
	addr = fmt.Sprintf("%s:%d", node, port)
	if client, err = ssh.Dial("tcp", addr, clientConfig); err != nil {
		return nil, err
	}
	return client, nil
}
