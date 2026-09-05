package main

import (
	"database/sql"
	"errors"
	"testing"
)

func newTestDB(t *testing.T) *sql.DB {
	t.Helper()
	db, err := openDB(":memory:")
	if err != nil {
		t.Fatalf("openDB returned error: %v", err)
	}
	t.Cleanup(func() { db.Close() })
	return db
}

func TestCreateAndFindUser(t *testing.T) {
	db := newTestDB(t)

	if _, err := createUser(db, "alice", "hashed-password"); err != nil {
		t.Fatalf("createUser returned error: %v", err)
	}

	user, err := findUserByName(db, "alice")
	if err != nil {
		t.Fatalf("findUserByName returned error: %v", err)
	}
	if user.Username != "alice" {
		t.Errorf("Username = %q, want %q", user.Username, "alice")
	}
}

func TestFindUserByNameNotFound(t *testing.T) {
	db := newTestDB(t)

	if _, err := findUserByName(db, "nobody"); !errors.Is(err, sql.ErrNoRows) {
		t.Fatalf("err = %v, want sql.ErrNoRows", err)
	}
}

func TestCreateAndListBlogs(t *testing.T) {
	db := newTestDB(t)
	userId, err := createUser(db, "alice", "hashed-password")
	if err != nil {
		t.Fatalf("createUser returned error: %v", err)
	}

	if _, err := createBlogByUserId(db, int64(userId), "alice", "Hello", "World"); err != nil {
		t.Fatalf("createBlogByUserId returned error: %v", err)
	}

	blogs, err := getAllBlogs(db, 10, 0)
	if err != nil {
		t.Fatalf("getAllBlogs returned error: %v", err)
	}
	if len(blogs) != 1 {
		t.Fatalf("got %d blogs, want 1", len(blogs))
	}
	if blogs[0].Username != "alice" {
		t.Errorf("Username = %q, want %q", blogs[0].Username, "alice")
	}
}

func TestUpdateBlogRejectsWrongOwner(t *testing.T) {
	db := newTestDB(t)
	ownerId, err := createUser(db, "alice", "hashed-password")
	if err != nil {
		t.Fatalf("createUser returned error: %v", err)
	}
	otherId, err := createUser(db, "mallory", "hashed-password")
	if err != nil {
		t.Fatalf("createUser returned error: %v", err)
	}

	blog, err := createBlogByUserId(db, int64(ownerId), "alice", "Hello", "World")
	if err != nil {
		t.Fatalf("createBlogByUserId returned error: %v", err)
	}

	if _, err := updateBlogByUserId(db, blog.ID, int64(otherId), "Hacked", "Hacked"); !errors.Is(err, ErrBlogNotFound) {
		t.Fatalf("err = %v, want ErrBlogNotFound", err)
	}

	if _, err := updateBlogByUserId(db, blog.ID, int64(ownerId), "Updated", "Updated content"); err != nil {
		t.Fatalf("updateBlogByUserId returned error: %v", err)
	}
}

func TestDeleteBlogRejectsWrongOwner(t *testing.T) {
	db := newTestDB(t)
	ownerId, err := createUser(db, "alice", "hashed-password")
	if err != nil {
		t.Fatalf("createUser returned error: %v", err)
	}
	otherId, err := createUser(db, "mallory", "hashed-password")
	if err != nil {
		t.Fatalf("createUser returned error: %v", err)
	}

	blog, err := createBlogByUserId(db, int64(ownerId), "alice", "Hello", "World")
	if err != nil {
		t.Fatalf("createBlogByUserId returned error: %v", err)
	}

	if err := deleteBlogByUserId(db, blog.ID, int64(otherId)); !errors.Is(err, ErrBlogNotFound) {
		t.Fatalf("err = %v, want ErrBlogNotFound", err)
	}

	if err := deleteBlogByUserId(db, blog.ID, int64(ownerId)); err != nil {
		t.Fatalf("deleteBlogByUserId returned error: %v", err)
	}
}
