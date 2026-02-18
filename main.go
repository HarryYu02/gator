package main

import (
	"fmt"
	"os"

	"github.com/harryyu02/gator/internal/config"
)

type state struct {
	cfg *config.Config
}

type command struct {
	name string
	args []string
}

type commands struct {
	handlerMap map[string]func(*state, command) error
}

func printErrorAndExit(err error) {
	fmt.Printf("err: %v\n", err)
	os.Exit(1)
}

func handlerLogin(s *state, cmd command) error {
	if len(cmd.args) == 0 {
		return fmt.Errorf("command login expects a username\n")
	}
	user := cmd.args[0]
	err := s.cfg.SetUser(user)
	if err != nil {
		return err
	}
	fmt.Printf("User %s has been set\n", user)
	return nil
}

func (c *commands) run(s *state, cmd command) error {
	handler, ok := c.handlerMap[cmd.name]
	if !ok {
		return fmt.Errorf("command %s not found\n", cmd.name)
	}
	return handler(s, cmd)
}

func (c *commands) register(name string, f func(*state, command) error) error {
	c.handlerMap[name] = f
	return nil
}

func main() {
	cfg, err := config.Read()
	if err != nil {
		printErrorAndExit(err)
	}
	appState := state{
		cfg: &cfg,
	}

	handlerMap := make(map[string]func (*state, command) error)
	appCommands := commands{
		handlerMap: handlerMap,
	}
	appCommands.register("login", handlerLogin)

	args := os.Args
	if len(args) < 2 {
		printErrorAndExit(fmt.Errorf("command not provided"))
	}
	cmd := command{
		name: args[1],
		args: args[2:],
	}

	err = appCommands.run(&appState, cmd)
	if err != nil {
		printErrorAndExit(err)
	}
}
