package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"

	"github.com/izmno/ansible-sftpgo/internal/ansible"
	"github.com/izmno/ansible-sftpgo/internal/sftpgo"
)

type UserInput struct {
	sftpgo.ClientConfig
	User  *sftpgo.User  `json:"userdata"`
	State ansible.State `json:"state"`
}

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, os.Kill)
	defer cancel()

	input := ansible.Read[UserInput]()

	c, err := sftpgo.NewClient(input.BaseURL, input.Username, input.Password)
	if err != nil {
		ansible.Fail(err)
	}

	var (
		u     = input.User
		state = input.State
	)

	u.Fix()

	existingSdkUser, err := c.GetUser(ctx, u.Username)
	if err != nil {
		ansible.Fail(fmt.Errorf("Failed to get user from SFTPGo server: %s", err.Error()))
	}

	existing := sftpgo.NewUserFromSdkUser(existingSdkUser)

	if state == ansible.StateAbsent && existing != nil {
		ansible.Change("User deleted", c.DeleteUser(ctx, u.Username))
	} else if state == ansible.StateAbsent {
		ansible.NoChange("User does not exist")
	} else if existing == nil {
		ansible.Change("User created", c.CreateUser(ctx, u.AsSdkUser()))
	} else if existing.NeedsUpdate(u) {
		ansible.Change("User updated", c.UpdateUser(ctx, u.AsSdkUser()))
	} else {
		ansible.Return(ansible.Unchanged("User is up to date"))
	}
}
