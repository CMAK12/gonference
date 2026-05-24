package conference

import (
	"errors"
	"sort"
	"sync"
	"time"

	"github.com/google/uuid"
)

var (
	ErrNotFound      = errors.New("conference not found")
	ErrUsernameTaken = errors.New("username already taken")
	ErrEmptyName     = errors.New("empty conference name")
	ErrEmptyUsername = errors.New("empty username")
)

type Conference struct {
	ID        string
	Name      string
	Members   []string
	CreatedAt time.Time
}

type entry struct {
	id        string
	name      string
	members   map[string]struct{}
	createdAt time.Time
}

func (e *entry) snapshot() Conference {
	members := make([]string, 0, len(e.members))
	for m := range e.members {
		members = append(members, m)
	}
	sort.Strings(members)
	return Conference{
		ID:        e.id,
		Name:      e.name,
		Members:   members,
		CreatedAt: e.createdAt,
	}
}

type Registry struct {
	mu          sync.RWMutex
	conferences map[string]*entry
}

func New() *Registry {
	return &Registry{conferences: make(map[string]*entry)}
}

func (r *Registry) Create(name string) (Conference, error) {
	if name == "" {
		return Conference{}, ErrEmptyName
	}

	e := &entry{
		id:        uuid.NewString(),
		name:      name,
		members:   make(map[string]struct{}),
		createdAt: time.Now(),
	}

	r.mu.Lock()
	r.conferences[e.id] = e
	r.mu.Unlock()

	return e.snapshot(), nil
}

func (r *Registry) List() []Conference {
	r.mu.RLock()
	defer r.mu.RUnlock()

	out := make([]Conference, 0, len(r.conferences))
	for _, e := range r.conferences {
		out = append(out, e.snapshot())
	}
	sort.Slice(out, func(i, j int) bool {
		return out[i].CreatedAt.Before(out[j].CreatedAt)
	})
	return out
}

func (r *Registry) Get(id string) (Conference, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	e, ok := r.conferences[id]
	if !ok {
		return Conference{}, false
	}
	return e.snapshot(), true
}

func (r *Registry) Join(id, username string) (Conference, error) {
	if username == "" {
		return Conference{}, ErrEmptyUsername
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	e, ok := r.conferences[id]
	if !ok {
		return Conference{}, ErrNotFound
	}
	if _, taken := e.members[username]; taken {
		return Conference{}, ErrUsernameTaken
	}
	e.members[username] = struct{}{}
	return e.snapshot(), nil
}

func (r *Registry) Leave(id, username string) (empty bool, err error) {
	if username == "" {
		return false, ErrEmptyUsername
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	e, ok := r.conferences[id]
	if !ok {
		return false, ErrNotFound
	}
	delete(e.members, username)
	return len(e.members) == 0, nil
}

func (r *Registry) Delete(id string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	delete(r.conferences, id)
}
