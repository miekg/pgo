package main

import (
	"bytes"
	"context"
	"fmt"
	"io/ioutil"
	"os"
	"os/user"
	"strings"

	"golang.org/x/crypto/ssh"
)

func querySSH(ctx context.Context, machine, command string, args []string) ([]byte, error) {
	var (
		key []byte
		err error
	)
	ident := ctx.Value("i").(string)
	switch ident {
	default:
		key, err = ioutil.ReadFile(ident)
		if err != nil {
			return nil, err
		}
	case "":
		key, err = IDFromEnv()
		if err != nil {
			return nil, fmt.Errorf("identity not given, -i flag; %v", err)
		}
	}
	port := ctx.Value("p").(string)
	if port == "" {
		port = "2222"
	}

	// Create the Signer for this private key.
	signer, err := ssh.ParsePrivateKey(key)
	if err != nil {
		return nil, err
	}

	user, err := user.Current()
	if err != nil {
		return nil, err
	}

	config := &ssh.ClientConfig{
		User:            user.Username,
		Auth:            []ssh.AuthMethod{ssh.PublicKeys(signer)},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
	}

	client, err := ssh.Dial("tcp", machine+":"+port, config)
	if err != nil {
		return nil, err
	}
	defer client.Close()
	ss, err := client.NewSession()
	if err != nil {
		return nil, err
	}
	defer ss.Close()

	stdoutBuf := &bytes.Buffer{}
	ss.Stdout = stdoutBuf

	cmdline := command + " " + strings.Join(args, " ")
	err = ss.Run(cmdline)
	return stdoutBuf.Bytes(), err
}

func IDFromEnv() ([]byte, error) {
	key := os.Getenv("PGOCTL_ID")
	if key == "" {
		return nil, fmt.Errorf("no enviroment variable found with name: %q", "PGOCTL_ID")
	}
	return []byte(key), nil
}
