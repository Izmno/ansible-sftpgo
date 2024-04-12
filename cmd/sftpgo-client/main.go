package main

import (
	"context"
	"encoding/json"
	"log"
	"os"

	"github.com/izmno/ansible-sftpgo/internal/sftpgo"
	"github.com/sirupsen/logrus"
	"github.com/urfave/cli/v2"
)

func main() {
	app := &cli.App{
		Name:  "SFTPGo client",
		Usage: "make an explosive entrance",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:     "base-url",
				Aliases:  []string{"url"},
				Usage:    "Base URL of the SFTPGo server",
				Required: true,
			},
			&cli.StringFlag{
				Name:     "username",
				Aliases:  []string{"u"},
				Usage:    "Admin username of the SFTPGo server",
				Required: true,
			},
			&cli.StringFlag{
				Name:     "password",
				Aliases:  []string{"p"},
				Usage:    "Admin password of the SFTPGo server",
				Required: true,
			},
		},
		Commands: []*cli.Command{
			{
				Name:    "user",
				Aliases: []string{"u"},
				Usage:   "Options for managing users",
				Subcommands: []*cli.Command{
					{
						Name:      "get",
						Usage:     "Get a user by username",
						Action:    GetUser,
						Args:      true,
						ArgsUsage: "USERNAME",
					},
					{
						Name:   "create",
						Usage:  "Create a new user from JSON data on stdin",
						Action: CreateUser,
					},
					{
						Name:   "update",
						Usage:  "Update an existing user from JSON data on stdin",
						Action: UpdateUser,
					},
					{
						Name:      "delete",
						Usage:     "Delete a user by username",
						Action:    DeleteUser,
						Args:      true,
						ArgsUsage: "USERNAME",
					},
				},
			},
		},
	}

	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}

	return
}

func GetUser(c *cli.Context) error {
	u, err := getUser(c, c.Args().First())
	if err != nil {
		return err
	}

	return json.NewEncoder(os.Stdout).Encode(u)
}

func CreateUser(c *cli.Context) error {
	u, err := readUser(c.Context)
	if err != nil {
		return err
	}

	client, err := createClient(c)
	if err != nil {
		return err
	}

	return client.CreateUser(c.Context, u.AsSdkUser())
}

func UpdateUser(c *cli.Context) error {
	u, err := readUser(c.Context)
	if err != nil {
		return err
	}

	client, err := createClient(c)
	if err != nil {
		return err
	}

	return client.CreateUser(c.Context, u.AsSdkUser())
}

func DeleteUser(c *cli.Context) error {
	return deleteUser(c, c.Args().First())
}

func readUser(ctx context.Context) (*sftpgo.User, error) {
	user := &sftpgo.User{}
	if err := json.NewDecoder(os.Stdin).Decode(user); err != nil {
		logrus.WithContext(ctx).WithError(err).Error("Failed to read user from stdin")
		return nil, err
	}

	user.Fix()

	return user, nil
}

func getUser(c *cli.Context, username string) (*sftpgo.User, error) {
	client, err := createClient(c)
	if err != nil {
		return nil, err
	}

	sdkUser, err := client.GetUser(c.Context, username)
	if err != nil {
		logrus.WithContext(c.Context).WithError(err).Error("Failed to get user from SFTPGo server")

		return nil, err
	}

	return sftpgo.NewUserFromSdkUser(sdkUser), nil
}

func deleteUser(c *cli.Context, username string) error {
	client, err := createClient(c)
	if err != nil {
		return err
	}

	return client.DeleteUser(c.Context, username)
}

func createClient(c *cli.Context) (*sftpgo.Client, error) {
	client, err := sftpgo.NewClient(c.String("base-url"), c.String("username"), c.String("password"))
	if err != nil {
		logrus.WithContext(c.Context).WithError(err).Error("Failed to create SFTPGo client")

		return nil, err
	}

	return client, nil
}
