package main

import (
	"context"
	"fmt"
	"strings"

	flag "github.com/spf13/pflag"
	"go.science.ru.nl/log"
)

type ExecContext struct {
	Identity string
	Port     string
}

var routes = map[string]struct{}{
	"up":   {},
	"down": {},
	"ps":   {},
	"pull": {},
	"logs": {},

	"ping": {},
}

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

	var out []byte
	_, ok := routes[command]
	if !ok {
		log.Fatalf("Command %q doesn't match any route", command)
	}

	out, err = querySSH(ctx, machine, name+"//"+command, flag.Args()[1:])
	if len(out) > 0 {
		fmt.Printf(string(out))
	}

	if err != nil {
		log.Error(err)
	}

}

// parseCommand parses: machine:dhz//ps in name (dhz) and command (status)
func parseCommand(s string) (machine, name, command string, error error) {
	items := strings.SplitN(s, ":", 2)
	if len(items) != 2 {
		return "", "", "", fmt.Errorf("expected machine:name//command, got %s", s)
	}
	machine = items[0]
	rest := items[1]
	items = strings.Split(rest, "//")
	if len(items) != 2 {
		return "", "", "", fmt.Errorf("expected name//command, got %s", rest)
	}
	name = items[0]
	command = items[1]
	return machine, name, command, nil
}
