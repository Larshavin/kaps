package utils

import (
	"errors"
	"fmt"
	"net"
	"os"
	"strings"
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
	Config *ssh.ClientConfig
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

	S.Config = sshConfig

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

// Change Client using ssh client as a proxy ( ssh client -> ssh server1 -> ssh server2 )
// Server 1 & Server 2 share same ssh key and user info ( ex: centos )
// If two servers have different ssh key and user info, you should use different ssh config in NewClientConn.
func (S *SSH) ChangeClient(client *ssh.Client, ip string) error {

	clientConn, err := client.Dial("tcp", fmt.Sprintf("%s:%d", ip, S.Port))
	if err != nil {
		fmt.Println(err)
		return err
	}

	ncc, chans, reqs, err := ssh.NewClientConn(clientConn, ip, S.Config)
	if err != nil {
		fmt.Println(err)
		return err
	}
	sClient := ssh.NewClient(ncc, chans, reqs)
	S.client = sClient
	return nil
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

	str := strings.Replace(dataPlaneJoinCommand, "\n", "", -1)

	defer client.Close()
	return str, nil
}

// RunCmd by proxied ssh client
func InjectDataplaneJoin(client SSH, mode int, cmd string, destIP string) (string, error) {
	err := client.Connect(mode) // If you are using a key file, use 'CertPublicKeyFile' instead. // [1 = CertPublicKeyFile, 2 = CertPassword]
	if err != nil {
		return "", err
	}

	err = client.ChangeClient(client.client, destIP)
	if err != nil {
		return "", err
	}

	cmd1 := `sudo tee /etc/sysctl.d/k8s.conf <<EOF
	net.bridge.bridge-nf-call-ip6tables = 1
	net.bridge.bridge-nf-call-iptables = 1
	EOF`
	cmd2 := `sudo modprobe br_netfilter`
	cmd3 := `sudo sysctl net.ipv4.ip_forward=1`
	_, err = client.RunCmd(cmd1)
	if err != nil {
		return "", err
	}
	_, err = client.RunCmd(cmd2)
	if err != nil {
		return "", err
	}
	_, err = client.RunCmd(cmd3)
	if err != nil {
		return "", err
	}

	output, err := client.RunCmd(cmd) // Cmd is `kubeadm join ...` for data-plane node
	if err != nil {
		return "", err
	}

	// str := strings.Replace(output, "\n", "", -1)

	defer client.Close()
	return output, nil
}
