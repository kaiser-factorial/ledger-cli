package client

import (
	"context"
	"time"

	"cloud.google.com/go/firestore"
	"github.com/kaiser/ledger-cli/internal/auth"
	"google.golang.org/api/iterator"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// Project represents a Ledger project
type Project struct {
	ID              string       `json:"id"`
	Name            string       `json:"name"`
	LastTouched     time.Time    `json:"lastTouched"`
	LastTouchReason string      `json:"lastTouchReason"`
	NextAction      string       `json:"nextAction"`
	StatusNote      string       `json:"statusNote"`
	TouchHistory    []TouchEntry `json:"touchHistory"`
}

// TouchEntry represents a single touch event in history
type TouchEntry struct {
	Timestamp time.Time `json:"timestamp"`
	Reason    string    `json:"reason"`
}

// Client wraps a Firestore client.
type Client struct {
	*firestore.Client
	projectID string
}

// New creates a new authenticated Firestore client.
func New(ctx context.Context) (*Client, error) {
	fsClient, err := auth.GetFirestoreClient(ctx)
	if err != nil {
		return nil, err
	}
	return &Client{Client: fsClient, projectID: "kaiser-ledger"}, nil
}

// Close closes the Firestore client
func (c *Client) Close() error {
	return c.Client.Close()
}

// isNotFound checks if an error represents "not found"
func isNotFound(err error) bool {
	if st, ok := status.FromError(err); ok {
		return st.Code() == codes.NotFound
	}
	return false
}

// GetAllProjects fetches all projects from Firestore
func (c *Client) GetAllProjects(ctx context.Context) ([]Project, error) {
	coll := c.Collection("projects")
	iter := coll.Documents(ctx)

	var projects []Project
	for {
		doc, err := iter.Next()
		if err != nil {
			if err == iterator.Done {
				break
			}
			return nil, err
		}

		var p Project
		if err := doc.DataTo(&p); err != nil {
			continue
		}
		p.ID = doc.Ref.ID
		projects = append(projects, p)
	}
	return projects, nil
}

// GetStaleProjects fetches projects that haven't been touched in 10+ days
func (c *Client) GetStaleProjects(ctx context.Context) ([]Project, error) {
	projects, err := c.GetAllProjects(ctx)
	if err != nil {
		return nil, err
	}

	var stale []Project
	tenDaysAgo := time.Now().Add(-10 * 24 * time.Hour)
	for _, p := range projects {
		if p.LastTouched.Before(tenDaysAgo) {
			stale = append(stale, p)
		}
	}
	return stale, nil
}

// TouchProject records a touch for a project, appending to touchHistory (capped at 8)
func (c *Client) TouchProject(ctx context.Context, slug, reason string) error {
	doc := c.Collection("projects").Doc(slug)

	now := time.Now()

	// Use a transaction to read-modify-write safely
	err := c.RunTransaction(ctx, func(ctx context.Context, tx *firestore.Transaction) error {
		snap, err := tx.Get(doc)
		if err != nil {
			// Document doesn't exist - create it
			if isNotFound(err) {
				newTouch := TouchEntry{Timestamp: now, Reason: reason}
				return tx.Set(doc, map[string]interface{}{
					"name":            slug,
					"lastTouched":     now,
					"lastTouchReason": reason,
					"touchHistory":    []TouchEntry{newTouch},
				})
			}
			return err
		}

		// Document exists - update it
		var existing Project
		if err := snap.DataTo(&existing); err != nil {
			return err
		}

		// Append to history, cap at 8
		newHistory := append(existing.TouchHistory, TouchEntry{Timestamp: now, Reason: reason})
		if len(newHistory) > 8 {
			newHistory = newHistory[len(newHistory)-8:]
		}

		return tx.Set(doc, map[string]interface{}{
			"name":            slug,
			"lastTouched":     now,
			"lastTouchReason": reason,
			"touchHistory":    newHistory,
		}, firestore.MergeAll)
	})

	return err
}