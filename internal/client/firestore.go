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

// ---------- Workload plan (Epic D) ----------
//
// bulwork (TypeScript) owns the WorkloadPlan/WorkflowTemplate schema (src/types.ts) and changes it
// often; this client deliberately treats plan/template payloads as opaque JSON rather than
// duplicating that schema in Go structs. All values already round-trip as JSON primitives/strings
// (dates are ISO strings, never Firestore Timestamp), so map[string]interface{} is lossless.

// planDocRef is the single source of the workload-plan path — bulwork's PlanStore is a singleton
// (one active plan at a time, no history yet). Kept as one helper so a future change (e.g. a
// history collection, or per-user scoping) is a one-line edit.
func (c *Client) planDocRef() *firestore.DocumentRef {
	return c.Collection("plans").Doc("current")
}

// GetPlan returns the current workload plan. ok=false (no error) means "no active plan" — a valid
// empty state, not a failure.
func (c *Client) GetPlan(ctx context.Context) (map[string]interface{}, bool, error) {
	snap, err := c.planDocRef().Get(ctx)
	if err != nil {
		if isNotFound(err) {
			return nil, false, nil
		}
		return nil, false, err
	}
	return snap.Data(), true, nil
}

// SetPlan overwrites the current plan wholesale (matches bulwork's LocalPlanStore.save(), which
// always rewrites the entire file — never a partial merge).
func (c *Client) SetPlan(ctx context.Context, data map[string]interface{}) error {
	_, err := c.planDocRef().Set(ctx, data)
	return err
}

// ClearPlan deletes the current plan. Idempotent — deleting an absent doc is not an error.
func (c *Client) ClearPlan(ctx context.Context) error {
	_, err := c.planDocRef().Delete(ctx)
	return err
}

// ---------- Workflow templates (Epic D) ----------

func (c *Client) templatesColl() *firestore.CollectionRef {
	return c.Collection("templates")
}

// ListTemplates returns every saved workflow template (never nil — an empty result marshals to
// `[]`, not `null`, matching what bulwork's TemplateStore.list() expects).
func (c *Client) ListTemplates(ctx context.Context) ([]map[string]interface{}, error) {
	iter := c.templatesColl().Documents(ctx)
	out := []map[string]interface{}{}
	for {
		doc, err := iter.Next()
		if err != nil {
			if err == iterator.Done {
				break
			}
			return nil, err
		}
		out = append(out, doc.Data())
	}
	return out, nil
}

// GetTemplate returns one template by id. ok=false (no error) means it doesn't exist.
func (c *Client) GetTemplate(ctx context.Context, id string) (map[string]interface{}, bool, error) {
	snap, err := c.templatesColl().Doc(id).Get(ctx)
	if err != nil {
		if isNotFound(err) {
			return nil, false, nil
		}
		return nil, false, err
	}
	return snap.Data(), true, nil
}

// SetTemplate upserts a template wholesale by id.
func (c *Client) SetTemplate(ctx context.Context, id string, data map[string]interface{}) error {
	_, err := c.templatesColl().Doc(id).Set(ctx, data)
	return err
}

// DeleteTemplate deletes a template by id. existed reports whether it was there beforehand, so
// callers can distinguish "deleted" from "wasn't there" (mirrors DeleteProject's Get-then-Delete).
func (c *Client) DeleteTemplate(ctx context.Context, id string) (bool, error) {
	doc := c.templatesColl().Doc(id)
	if _, err := doc.Get(ctx); err != nil {
		if isNotFound(err) {
			return false, nil
		}
		return false, err
	}
	if _, err := doc.Delete(ctx); err != nil {
		return false, err
	}
	return true, nil
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