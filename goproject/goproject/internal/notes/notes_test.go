package notes_test

import (
	"testing"

	"goproject/internal/notes"
)

func TestCreateAndList(t *testing.T) {
	s := notes.NewStore()

	n := s.Create("user1", notes.CreateInput{Title: "Buy milk", Priority: notes.PriorityHigh})
	if n.Title != "Buy milk" {
		t.Errorf("unexpected title: %s", n.Title)
	}
	if n.Done {
		t.Error("new note should not be done")
	}

	list := s.List("user1")
	if len(list) != 1 {
		t.Errorf("expected 1 note, got %d", len(list))
	}

	// other user sees nothing
	list2 := s.List("user2")
	if len(list2) != 0 {
		t.Errorf("user2 should see 0 notes, got %d", len(list2))
	}
}

func TestGetUpdateDelete(t *testing.T) {
	s := notes.NewStore()
	n := s.Create("u1", notes.CreateInput{Title: "Test"})

	// get
	got, err := s.Get("u1", n.ID)
	if err != nil || got.ID != n.ID {
		t.Fatalf("get failed: %v", err)
	}

	// forbidden get
	_, err = s.Get("u2", n.ID)
	if err != notes.ErrForbidden {
		t.Errorf("expected ErrForbidden, got %v", err)
	}

	// update
	title := "Updated"
	done := true
	updated, err := s.Update("u1", n.ID, notes.UpdateInput{Title: &title, Done: &done})
	if err != nil {
		t.Fatalf("update failed: %v", err)
	}
	if updated.Title != "Updated" || !updated.Done {
		t.Error("update did not apply")
	}

	// delete
	if err := s.Delete("u1", n.ID); err != nil {
		t.Fatalf("delete failed: %v", err)
	}
	_, err = s.Get("u1", n.ID)
	if err != notes.ErrNotFound {
		t.Errorf("expected ErrNotFound after delete, got %v", err)
	}
}
