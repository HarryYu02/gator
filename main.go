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
	"github.com/harryyu02/gator/internal/rss"

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

func middlewareLoggedIn(handler func(s *state, cmd command, user database.User) error) func(*state, command) error {
	return func(s *state, cmd command) error {
		user, err := s.db.GetUser(context.Background(), s.cfg.CurrentUserName)
		if err != nil {
			return err
		}
		return handler(s, cmd, user)
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
		return fmt.Errorf("command register expects a username\n")
	}
	username := cmd.args[0]

	ctx := context.Background()
	user, err := s.db.CreateUser(ctx, database.CreateUserParams{
		ID:        uuid.New(),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		Name:      username,
	})
	if err != nil {
		return err
	}

	fmt.Println("User created!")
	handleLogin(s, command{
		name: "login",
		args: []string{user.Name},
	})

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

func handleAgg(s *state, cmd command) error {
	ctx := context.Background()
	feed, err := rss.FetchFeed(ctx, "https://www.wagslane.dev/index.xml")
	if err != nil {
		return err
	}
	fmt.Printf("feed: %v\n", *feed)

	return nil
}

func handleAddFeed(s *state, cmd command, user database.User) error {
	if len(cmd.args) < 2 {
		return fmt.Errorf("command addfeed expects a name and a url\n")
	}
	name := cmd.args[0]
	url := cmd.args[1]

	ctx := context.Background()
	_, err := s.db.CreateFeed(ctx, database.CreateFeedParams{
		ID: uuid.New(),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		Name: name,
		Url: url,
		UserID: user.ID,
	})
	if err != nil {
		return err
	}

	handleFollow(s, command{
		name: "follow",
		args: []string{url},
	}, user)
	return nil
}

func handleFeeds(s *state, cmd command) error {
	ctx := context.Background()
	feeds, err := s.db.GetFeedsWithUsers(ctx)
	if err != nil {
		return err
	}

	for _, feed := range feeds {
		fmt.Println("--------------------")
		fmt.Printf("Feed name: %s\n", feed.Name)
		fmt.Printf("Feed url: %s\n", feed.Url)
		fmt.Printf("Feed creator: %s\n", feed.Name_2)
		fmt.Println("--------------------")
	}

	return nil
}

func handleFollow(s *state, cmd command, user database.User) error {
	if len(cmd.args) < 1 {
		return fmt.Errorf("command follow expects a url\n")
	}
	url := cmd.args[0]

	ctx := context.Background()
	feedToFollow, err := s.db.GetFeedByUrl(ctx, url)
	if err != nil {
		return err
	}
	feedFollow, err := s.db.CreateFeedFollow(ctx, database.CreateFeedFollowParams{
		ID: uuid.New(),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		UserID: user.ID,
		FeedID: feedToFollow.ID,
	})
	if err != nil {
		return err
	}
	fmt.Println("--------------------")
	fmt.Printf("Feed name: %s\n", feedFollow.FeedName)
	fmt.Printf("Feed creator: %s\n", feedFollow.UserName)
	fmt.Println("--------------------")
	return nil
}

func handleFollowing(s *state, cmd command, user database.User) error {
	ctx := context.Background()
	feeds, err := s.db.GetFeedFollowsForUser(ctx, user.ID)
	if err != nil {
		return err
	}
	for _, feed := range feeds {
		fmt.Println("--------------------")
		fmt.Printf("Feed name: %s\n", feed.FeedName)
		fmt.Printf("Feed creator: %s\n", feed.UserName)
		fmt.Println("--------------------")
	}
	return nil
}

func handleUnfollow(s *state, cmd command, user database.User) error {
	if len(cmd.args) < 1 {
		return fmt.Errorf("command unfollow expects a url\n")
	}
	url := cmd.args[0]

	ctx := context.Background()
	feedToUnfollow, err := s.db.GetFeedByUrl(ctx, url)
	if err != nil {
		return err
	}
	err = s.db.DeleteFeedFollow(ctx, database.DeleteFeedFollowParams{
		UserID: user.ID,
		FeedID: feedToUnfollow.ID,
	})
	if err != nil {
		return err
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
	appCommands.register("agg", handleAgg)
	appCommands.register("addfeed", middlewareLoggedIn(handleAddFeed))
	appCommands.register("feeds", handleFeeds)
	appCommands.register("follow", middlewareLoggedIn(handleFollow))
	appCommands.register("following", middlewareLoggedIn(handleFollowing))
	appCommands.register("unfollow", middlewareLoggedIn(handleUnfollow))

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
