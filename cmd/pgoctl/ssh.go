package main

import (
	"bytes"
	"context"
	"fmt"
	"io/ioutil"
	"os/user"
	"strings"

	"golang.org/x/crypto/ssh"
)

func querySSH(ctx context.Context, machine, command string, args []string) ([]byte, error) {
	ident := ctx.Value("i").(string)
	if ident == "" {
		return nil, fmt.Errorf("identity file not given, -i flag")
	}
	port := ctx.Value("p").(string)
	if port == "" {
		port = "2222"
	}
	println(ident)
	key, err := ioutil.ReadFile(ident)
	if err != nil {
		return nil, err
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
