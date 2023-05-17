package main

import (
	"context"
	"fmt"
	"os"
	"strings"

	flag "github.com/spf13/pflag"
	"go.science.ru.nl/log"
)

type ExecContext struct {
	Identity string
	Port     string
}

// <machine>:dhz//ps
func main() {
	exec := ExecContext{}
	flag.StringVarP(&exec.Identity, "identity", "i", "", "identify file")
	flag.StringVarP(&exec.Port, "port", "p", "2222", "remote ssh port to use")

	flag.Parse()

	machine, name, command, err := parseCommand(flag.Arg(0))
	if err != nil {
		log.Fatal(err)
	}

	ctx := context.TODO()
	ctx = context.WithValue(ctx, "i", exec.Identity)
	ctx = context.WithValue(ctx, "p", exec.Port)

	switch command {
	case "ps":
		err = cmdPs(ctx, machine, name, command, os.Args[1:])
	}
	if err != nil {
		log.Error(err)
	}

}

func cmdPs(ctx context.Context, machine, name, command string, args []string) error {
	out, err := querySSH(ctx, machine, name+"//"+command, args...)
	if err != nil {
		return err
	}
	fmt.Print(out)
	return nil
}

// parseCommand parses: machine:dhz//ps in name (dhz) and command (status)
func parseCommand(s string) (machine, name, command string, error error) {
	items := strings.SplitN(s, ":", 2)
	if len(items) != 2 {
		return "", "", "", fmt.Errorf("expected machine:name//command, got %s", s)
	}
	machine = items[0]
	items = strings.Split(items[1], "//")
	if len(items) != 2 {
		return "", "", "", fmt.Errorf("expected name//command, got %s", items[1])
	}
	name = items[0]
	command = items[1]
	return machine, name, command, nil
}
