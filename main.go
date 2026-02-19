package main

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"time"

	"github.com/google/uuid"
	"github.com/harryyu02/gator/internal/config"
	"github.com/harryyu02/gator/internal/database"

	_ "github.com/lib/pq"
)

type state struct {
	db  *database.Queries
	cfg *config.Config
}

type command struct {
	name string
	args []string
}

type commands struct {
	handlerMap map[string]func(*state, command) error
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

func printErrorAndExit(err error) {
	if err != nil {
		fmt.Printf("err: %v\n", err)
		os.Exit(1)
	}
}

func handleLogin(s *state, cmd command) error {
	if len(cmd.args) == 0 {
		return fmt.Errorf("command login expects a username\n")
	}
	username := cmd.args[0]

	ctx := context.Background()
	_, err := s.db.GetUser(ctx, username)
	if err != nil {
		return err
	}

	err = s.cfg.SetUser(username)
	if err != nil {
		return err
	}
	fmt.Printf("User %s has been set\n", username)
	return nil
}

func handleRegister(s *state, cmd command) error {
	if len(cmd.args) == 0 {
		return fmt.Errorf("command login expects a username\n")
	}
	username := cmd.args[0]

	ctx := context.Background()
	user, err := s.db.CreateUser(ctx, database.CreateUserParams{
		ID: uuid.New(),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		Name: username,
	})
	if err != nil {
		return err
	}

	handleLogin(s, command{
		name: "login",
		args: []string{username},
	})

	fmt.Println("User created!")
	fmt.Printf("user: %v\n", user)

	return nil
}

func handleReset(s *state, cmd command) error {
	ctx := context.Background()
	return s.db.DeleteAllUsers(ctx)
}

func handleUsers(s *state, cmd command) error {
	ctx := context.Background()
	users, err := s.db.GetUsers(ctx)
	if err != nil {
		return err
	}

	for _, user := range users {
		if s.cfg.CurrentUserName == user.Name {
			fmt.Printf("* %s (current)\n", user.Name)
		} else {
			fmt.Printf("* %s\n", user.Name)
		}
	}
	return nil
}

func main() {
	cfg, err := config.Read()
	if err != nil {
		printErrorAndExit(err)
	}
	db, err := sql.Open("postgres", cfg.DbUrl)
	if err != nil {
		printErrorAndExit(err)
	}
	dbQueries := database.New(db)
	appState := state{
		db:  dbQueries,
		cfg: &cfg,
	}

	handlerMap := make(map[string]func(*state, command) error)
	appCommands := commands{
		handlerMap: handlerMap,
	}
	appCommands.register("login", handleLogin)
	appCommands.register("register", handleRegister)
	appCommands.register("reset", handleReset)
	appCommands.register("users", handleUsers)

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
