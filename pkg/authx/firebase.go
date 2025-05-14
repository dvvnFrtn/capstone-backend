package authx

import (
	"context"
	"sync"

	firebase "firebase.google.com/go/v4"
	"firebase.google.com/go/v4/auth"
	"google.golang.org/api/option"
)

type FirebaseClients struct {
	App  *firebase.App
	Auth *auth.Client
}

var (
	instance *FirebaseClients
	once     sync.Once
)

func InitFirebase(ctx context.Context, credentialsPath string) (*FirebaseClients, error) {
	var err error

	once.Do(func() {
		opt := option.WithCredentialsFile(credentialsPath)

		app, appErr := firebase.NewApp(ctx, nil, opt)
		if appErr != nil {
			err = appErr
			return
		}

		authClient, authErr := app.Auth(ctx)
		if authErr != nil {
			err = authErr
			return
		}

		instance = &FirebaseClients{
			App:  app,
			Auth: authClient,
		}
	})

	return instance, err
}
