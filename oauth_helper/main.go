package main

import (
	"context"
	"fmt"
	"net/url"
	"os"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
)

func main() {
	token := os.Getenv("GMAIL_API_TOKEN")

	if len(token) == 0 {
		archiverState := "archiver_state"
		ctx := context.Background()
		clientId := os.Getenv("GMAIL_API_ID")

		if len(clientId) == 0 {
			panic("GMAIL_API_ID doesn't exists, check environmental variables")
		}

		clientSecret := os.Getenv("GMAIL_API_SECRET")

		if len(clientSecret) == 0 {
			panic("GMAIL_API_SECRET doesn't exists, check environmental variables")
		}

		config := &oauth2.Config{
			ClientID:     clientId,
			ClientSecret: clientSecret,
			Scopes:       []string{"https://mail.google.com/"},
			Endpoint:     google.Endpoint,
			RedirectURL:  "http://localhost",
		}

		authURL := config.AuthCodeURL(archiverState, oauth2.AccessTypeOffline)

		var oauthCode string

		fmt.Printf("Your link for oauth: %s\n", authURL)
		fmt.Println("Enter auth code: ")
		_, err := fmt.Scanln(&oauthCode)

		if err != nil {
			fmt.Println("error at scan")
			panic(err)
		}

		unescapedCode, err := url.QueryUnescape(oauthCode)

		if err != nil {
			fmt.Println("error at unescape")
			panic(err)
		}

		token, err := config.Exchange(ctx, unescapedCode)

		if err != nil {
			fmt.Println("error at exchange")
			panic(err)
		}

		err = os.WriteFile("/app/vol/token.txt", []byte(token.AccessToken), 0644)

		if err != nil {
			panic(err)
		}
	} else {
		fmt.Println("Token already set")
	}
}
