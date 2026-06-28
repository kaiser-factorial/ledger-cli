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
	LastTouchReason string       `json:"lastTouchReason"`
	NextAction      string       `json:"nextAction"`
	StatusNote      string       `json:"statusNote"`
	TouchHistory    []TouchEntry `json:"touchHistory"`
	Archived        bool         `json:"archived"`
	ArchivedAt      time.Time    `json:"archivedAt,omitempty"`
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

// IsNotFound reports whether err is a Firestore "not found" error.
func IsNotFound(err error) bool { return isNotFound(err) }

// fetchProjects returns every project document (archived or not).
func (c *Client) fetchProjects(ctx context.Context) ([]Project, error) {
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

// GetAllProjects fetches active (non-archived) projects. Archived projects are
// excluded from status, review, analyze, export, and stale checks by default.
func (c *Client) GetAllProjects(ctx context.Context) ([]Project, error) {
	all, err := c.fetchProjects(ctx)
	if err != nil {
		return nil, err
	}
	var active []Project
	for _, p := range all {
		if !p.Archived {
			active = append(active, p)
		}
	}
	return active, nil
}

// GetAllProjectsIncludingArchived fetches every project (active + archived).
func (c *Client) GetAllProjectsIncludingArchived(ctx context.Context) ([]Project, error) {
	return c.fetchProjects(ctx)
}

// GetArchivedProjects fetches only archived projects.
func (c *Client) GetArchivedProjects(ctx context.Context) ([]Project, error) {
	all, err := c.fetchProjects(ctx)
	if err != nil {
		return nil, err
	}
	var archived []Project
	for _, p := range all {
		if p.Archived {
			archived = append(archived, p)
		}
	}
	return archived, nil
}

// ArchiveProject soft-archives a project. Errors NotFound if it doesn't exist
// (unlike touch/note, it never creates the doc).
func (c *Client) ArchiveProject(ctx context.Context, id string) error {
	now := time.Now()
	_, err := c.Collection("projects").Doc(id).Update(ctx, []firestore.Update{
		{Path: "archived", Value: true},
		{Path: "archivedAt", Value: now},
		{Path: "updatedAt", Value: now},
	})
	return err
}

// RestoreProject clears the archived flag, returning a project to the active list.
func (c *Client) RestoreProject(ctx context.Context, id string) error {
	_, err := c.Collection("projects").Doc(id).Update(ctx, []firestore.Update{
		{Path: "archived", Value: firestore.Delete},
		{Path: "archivedAt", Value: firestore.Delete},
		{Path: "updatedAt", Value: time.Now()},
	})
	return err
}

// DeleteProject permanently deletes a project. Returns NotFound if it doesn't
// exist (Firestore Delete is otherwise idempotent).
func (c *Client) DeleteProject(ctx context.Context, id string) error {
	doc := c.Collection("projects").Doc(id)
	if _, err := doc.Get(ctx); err != nil {
		return err
	}
	_, err := doc.Delete(ctx)
	return err
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