package util

import (
	"fmt"
	"time"

	"golang.org/x/crypto/ssh"
)

func CheckSSHConection(ipAddr, port string, username string, timeout time.Duration, keyPair RSAKeyPair) error {
	signer, err := ssh.ParsePrivateKey([]byte(keyPair.PrivateKey))

	if err != nil {
		return fmt.Errorf("failed to check ssh connection: %w", err)
	}

	config := &ssh.ClientConfig{
		User: username,
		Auth: []ssh.AuthMethod{
			ssh.PublicKeys(signer),
		},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
		Timeout:         timeout,
	}
	addr := fmt.Sprintf("%s:%s", ipAddr, port)
	_, err = ssh.Dial("tcp", addr, config)

	if err != nil {
		return fmt.Errorf("failed to dial: %w", err)
	}
	return nil
}
