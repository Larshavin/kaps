package utils

import (
	"errors"
	"fmt"
	"net"
	"os"
	"time"

	"golang.org/x/crypto/ssh"
)

const (
	CertPassword      = 1
	CertPublicKeyFile = 2
	DefaultTimeout    = 3 // Second
)

type SSH struct {
	IP   string
	User string
	Cert string //password or key file path
	Port int
	// session *ssh.Session
	client *ssh.Client
}

func (S *SSH) readPublicKeyFile(file string) ssh.AuthMethod {
	buffer, err := os.ReadFile(file)
	if err != nil {
		return nil
	}

	key, err := ssh.ParsePrivateKey(buffer)
	if err != nil {
		return nil
	}

	return ssh.PublicKeys(key)
}

// Connect the SSH Server
func (S *SSH) Connect(mode int) error {
	var sshConfig *ssh.ClientConfig
	var auth []ssh.AuthMethod
	if mode == CertPassword {
		auth = []ssh.AuthMethod{
			ssh.Password(S.Cert),
		}
	} else if mode == CertPublicKeyFile {
		auth = []ssh.AuthMethod{
			S.readPublicKeyFile(S.Cert),
		}
	} else {
		// log.Println("does not support mode: ", mode)
		return errors.New("does not support mode")
	}

	sshConfig = &ssh.ClientConfig{
		User: S.User,
		Auth: auth,
		HostKeyCallback: func(hostname string, remote net.Addr, key ssh.PublicKey) error {
			return nil
		},
		Timeout: time.Second * DefaultTimeout,
	}

	client, err := ssh.Dial("tcp", fmt.Sprintf("%s:%d", S.IP, S.Port), sshConfig)
	if err != nil {
		fmt.Println(err)
		return err
	}

	S.client = client
	return nil
}

func (S *SSH) Close() {
	S.client.Close()
}

// RunCmd to SSH Server
func (S *SSH) RunCmd(cmd string) (string, error) {
	session, err := S.client.NewSession()
	if err != nil {
		return "", err
	}
	defer session.Close()

	out, err := session.CombinedOutput(cmd)
	if err != nil {
		return "", err
	}

	return string(out), nil
}

// RunCmd for specific reason : get kubeadm join command from k8s control-plane node
func GetKubeadmJoinOutput(client SSH, mode int) (string, error) {

	err := client.Connect(mode) // If you are using a key file, use 'CertPublicKeyFile' instead. // [1 = CertPublicKeyFile, 2 = CertPassword]
	if err != nil {
		return "", err
	}

	_, err = client.RunCmd("cloud-init status --wait") // blocking effect until cloud-init is complete
	if err != nil {
		return "", err
	}

	dataPlaneJoinCommand, err := client.RunCmd(`kubeadm token create --print-join-command`)
	if err != nil {
		return "", err
	}

	defer client.Close()
	return dataPlaneJoinCommand, nil
}

// RunCmd for general purpose. Input cmd is slice of string.
func GetSSHOutputs(client SSH, mode int, cmd []string) ([]string, error) {
	err := client.Connect(mode)
	if err != nil {
		return nil, err
	}

	outputs := []string{}

	for _, c := range cmd {
		out, err := client.RunCmd(c)
		if err != nil {
			return nil, err
		}
		fmt.Println(out)
		outputs = append(outputs, out)
	}

	defer client.Close()
	return outputs, nil
}
