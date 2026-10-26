package post

import (
	"context"
	"errors"
	"testing"

	"backend-core-go/internal/shared/pagination"
)

type mockRepo struct {
	posts      map[int64]Post
	lastFilter ListFilter
	updated    bool
	deleted    bool
}

func newMockRepo() *mockRepo {
	return &mockRepo{posts: map[int64]Post{
		1: {ID: 1, AuthorID: 10, Title: "Published", Slug: "published", Status: StatusPublished},
		2: {ID: 2, AuthorID: 10, Title: "Draft", Slug: "draft", Status: StatusDraft},
	}}
}

func (m *mockRepo) List(_ context.Context, f ListFilter, _ pagination.Params) ([]Post, int64, error) {
	m.lastFilter = f
	return nil, 0, nil
}

func (m *mockRepo) FindByID(_ context.Context, id int64) (*Post, error) {
	p, ok := m.posts[id]
	if !ok {
		return nil, ErrPostNotFound
	}
	return &p, nil
}

func (m *mockRepo) Create(_ context.Context, p *Post) error {
	for _, existing := range m.posts {
		if existing.Slug == p.Slug {
			return ErrSlugTaken
		}
	}
	p.ID = int64(len(m.posts) + 1)
	m.posts[p.ID] = *p
	return nil
}

func (m *mockRepo) Update(_ context.Context, id int64, fields UpdateFields) (*Post, error) {
	m.updated = true
	p := m.posts[id]
	if fields.Title != nil {
		p.Title = *fields.Title
	}
	return &p, nil
}

func (m *mockRepo) Publish(_ context.Context, id int64) (*Post, error) {
	p := m.posts[id]
	p.Status = StatusPublished
	return &p, nil
}

func (m *mockRepo) Unpublish(_ context.Context, id int64) (*Post, error) {
	p := m.posts[id]
	p.Status = StatusDraft
	return &p, nil
}

func (m *mockRepo) Delete(context.Context, int64) error {
	m.deleted = true
	return nil
}

type mockAuthz map[int64][]string

func (a mockAuthz) HasPermission(_ context.Context, userID int64, perm string) (bool, error) {
	for _, p := range a[userID] {
		if p == perm {
			return true, nil
		}
	}
	return false, nil
}

// 10 = author, 20 = author lain, 30 = editor (post:manage)
var authz = mockAuthz{30: {PermManage}}

func TestCreate_DraftWithSlug(t *testing.T) {
	repo := newMockRepo()
	p, err := NewService(repo, authz).Create(context.Background(), 20, CreatePostRequest{Title: "  Belajar Go: Dasar & Tips!  "})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	if p.Slug != "belajar-go-dasar-tips" || p.Status != StatusDraft || p.AuthorID != 20 {
		t.Fatalf("got slug=%q status=%q author=%d", p.Slug, p.Status, p.AuthorID)
	}
}

func TestCreate_InvalidTitleAndDuplicate(t *testing.T) {
	svc := NewService(newMockRepo(), authz)
	if _, err := svc.Create(context.Background(), 20, CreatePostRequest{Title: "!!!"}); !errors.Is(err, ErrInvalidTitle) {
		t.Fatalf("err = %v, want ErrInvalidTitle", err)
	}
	if _, err := svc.Create(context.Background(), 20, CreatePostRequest{Title: "Draft"}); !errors.Is(err, ErrSlugTaken) {
		t.Fatalf("err = %v, want ErrSlugTaken", err)
	}
}

func TestList_VisibilityByPermission(t *testing.T) {
	repo := newMockRepo()
	svc := NewService(repo, authz)
	ctx := context.Background()

	if _, _, err := svc.List(ctx, 20, ListQuery{}, pagination.New(1, 20)); err != nil {
		t.Fatalf("List: %v", err)
	}
	if repo.lastFilter.VisibleTo != 20 {
		t.Fatalf("author should only see published + own drafts, VisibleTo = %d", repo.lastFilter.VisibleTo)
	}

	if _, _, err := svc.List(ctx, 30, ListQuery{}, pagination.New(1, 20)); err != nil {
		t.Fatalf("List: %v", err)
	}
	if repo.lastFilter.VisibleTo != 0 {
		t.Fatal("editor with post:manage should see every post")
	}
}

func TestGet_DraftVisibility(t *testing.T) {
	tests := []struct {
		name    string
		actorID int64
		wantErr error
	}{
		{"author sees own draft", 10, nil},
		{"other author gets 404", 20, ErrPostNotFound},
		{"editor sees any draft", 30, nil},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := NewService(newMockRepo(), authz).Get(context.Background(), tt.actorID, 2)
			if !errors.Is(err, tt.wantErr) {
				t.Fatalf("err = %v, want %v", err, tt.wantErr)
			}
		})
	}
}

func TestUpdate_Ownership(t *testing.T) {
	title := "Judul Baru"
	tests := []struct {
		name    string
		actorID int64
		postID  int64
		wantErr error
	}{
		{"author", 10, 1, nil},
		{"other author on published post", 20, 1, ErrNotAuthor},
		{"other author on draft is hidden", 20, 2, ErrPostNotFound},
		{"editor", 30, 1, nil},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := newMockRepo()
			_, err := NewService(repo, authz).Update(context.Background(), tt.actorID, tt.postID, UpdatePostRequest{Title: &title})
			if !errors.Is(err, tt.wantErr) {
				t.Fatalf("err = %v, want %v", err, tt.wantErr)
			}
			if repo.updated != (tt.wantErr == nil) {
				t.Fatalf("updated = %v, want %v", repo.updated, tt.wantErr == nil)
			}
		})
	}
}

func TestDelete_NonAuthorForbidden(t *testing.T) {
	repo := newMockRepo()
	if err := NewService(repo, authz).Delete(context.Background(), 20, 1); !errors.Is(err, ErrNotAuthor) {
		t.Fatalf("err = %v, want ErrNotAuthor", err)
	}
	if repo.deleted {
		t.Fatal("post must not be deleted by non-author")
	}
}

func TestSlugify(t *testing.T) {
	tests := map[string]string{
		"Hello World":          "hello-world",
		"  Go 1.26 -- Rilis! ": "go-1-26-rilis",
		"Ünïcode ✓":            "n-code",
		"***":                  "",
	}
	for in, want := range tests {
		if got := slugify(in); got != want {
			t.Errorf("slugify(%q) = %q, want %q", in, got, want)
		}
	}
}
