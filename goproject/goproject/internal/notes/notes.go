package notes

import (
	"errors"
	"fmt"
	"sync"
	"time"
)

var (
	ErrNotFound   = errors.New("note not found")
	ErrForbidden  = errors.New("access denied")
)

type Priority string

const (
	PriorityLow    Priority = "low"
	PriorityMedium Priority = "medium"
	PriorityHigh   Priority = "high"
)

type Note struct {
	ID        string    `json:"id"`
	UserID    string    `json:"user_id"`
	Title     string    `json:"title"`
	Body      string    `json:"body"`
	Done      bool      `json:"done"`
	Priority  Priority  `json:"priority"`
	Tags      []string  `json:"tags"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type Store struct {
	mu      sync.RWMutex
	notes   map[string]*Note
	counter int
}

func NewStore() *Store {
	return &Store{notes: make(map[string]*Note)}
}

func (s *Store) newID() string {
	s.counter++
	return fmt.Sprintf("note_%d", s.counter)
}

type CreateInput struct {
	Title    string   `json:"title"`
	Body     string   `json:"body"`
	Priority Priority `json:"priority"`
	Tags     []string `json:"tags"`
}

type UpdateInput struct {
	Title    *string   `json:"title,omitempty"`
	Body     *string   `json:"body,omitempty"`
	Done     *bool     `json:"done,omitempty"`
	Priority *Priority `json:"priority,omitempty"`
	Tags     []string  `json:"tags,omitempty"`
}

func (s *Store) Create(userID string, input CreateInput) *Note {
	s.mu.Lock()
	defer s.mu.Unlock()

	now := time.Now()
	priority := input.Priority
	if priority == "" {
		priority = PriorityMedium
	}
	tags := input.Tags
	if tags == nil {
		tags = []string{}
	}

	note := &Note{
		ID:        s.newID(),
		UserID:    userID,
		Title:     input.Title,
		Body:      input.Body,
		Done:      false,
		Priority:  priority,
		Tags:      tags,
		CreatedAt: now,
		UpdatedAt: now,
	}
	s.notes[note.ID] = note
	return note
}

func (s *Store) Get(userID, noteID string) (*Note, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	note, ok := s.notes[noteID]
	if !ok {
		return nil, ErrNotFound
	}
	if note.UserID != userID {
		return nil, ErrForbidden
	}
	return note, nil
}

func (s *Store) List(userID string) []*Note {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var result []*Note
	for _, n := range s.notes {
		if n.UserID == userID {
			result = append(result, n)
		}
	}
	return result
}

func (s *Store) Update(userID, noteID string, input UpdateInput) (*Note, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	note, ok := s.notes[noteID]
	if !ok {
		return nil, ErrNotFound
	}
	if note.UserID != userID {
		return nil, ErrForbidden
	}

	if input.Title != nil {
		note.Title = *input.Title
	}
	if input.Body != nil {
		note.Body = *input.Body
	}
	if input.Done != nil {
		note.Done = *input.Done
	}
	if input.Priority != nil {
		note.Priority = *input.Priority
	}
	if input.Tags != nil {
		note.Tags = input.Tags
	}
	note.UpdatedAt = time.Now()
	return note, nil
}

func (s *Store) Delete(userID, noteID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	note, ok := s.notes[noteID]
	if !ok {
		return ErrNotFound
	}
	if note.UserID != userID {
		return ErrForbidden
	}
	delete(s.notes, noteID)
	return nil
}
