package github

import (
	"context"
)

type UserDetails struct {
	Name  string
	Email string
}

func FetchUserDetails(ctx context.Context, getClient GetClientFn) (UserDetails, error) {
	client, err := getClient(ctx)
	if err != nil {
		return UserDetails{}, err
	}

	user, _, err := client.Users.Get(ctx, "")
	if err != nil {
		return UserDetails{}, err
	}

	userName := user.GetName()
	if userName == "" {
		userName = user.GetLogin()
	}

	userEmail := user.GetEmail()
	if userEmail == "" {
		emails, _, err := client.Users.ListEmails(ctx, nil)
		if err != nil {
			return UserDetails{}, err
		}
		for _, email := range emails {
			if email.GetPrimary() {
				userEmail = email.GetEmail()
				break
			}
		}
	}

	return UserDetails{
		Name:  userName,
		Email: userEmail,
	}, nil
} 