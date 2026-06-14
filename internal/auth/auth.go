package auth

import (
	"context"
	"fmt"
	"os"

	"cloud.google.com/go/firestore"
	firebase "firebase.google.com/go/v4"
	"google.golang.org/api/option"
)

// GetFirestoreClient returns a Firestore client using service account or default credentials.
// Supports headless / CI use via GOOGLE_APPLICATION_CREDENTIALS.
func GetFirestoreClient(ctx context.Context) (*firestore.Client, error) {
	var app *firebase.App
	var err error

	// Default project ID for Ledger
	projectID := "kaiser-ledger"
	if envID := os.Getenv("LEDGER_PROJECT_ID"); envID != "" {
		projectID = envID
	}

	// Build firebase config with project ID
	cfg := &firebase.Config{
		ProjectID: projectID,
	}

	if credPath := os.Getenv("GOOGLE_APPLICATION_CREDENTIALS"); credPath != "" {
		opt := option.WithCredentialsFile(credPath)
		app, err = firebase.NewApp(ctx, cfg, opt)
	} else {
		// Falls back to gcloud auth or default credentials
		app, err = firebase.NewApp(ctx, cfg)
	}

	if err != nil {
		return nil, fmt.Errorf("failed to initialize Firebase: %w", err)
	}

	client, err := app.Firestore(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to create Firestore client: %w", err)
	}
	return client, nil
}