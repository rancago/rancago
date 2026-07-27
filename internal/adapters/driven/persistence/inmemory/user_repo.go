package inmemory

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/rancago/framework/internal/domain/entities"
	"github.com/rancago/framework/internal/domain/valueobjects"
	"github.com/rancago/framework/internal/ports/driven"
)

type InMemoryUserRepo struct {
	mu     sync.RWMutex
	users  map[string]*entities.User
	roles  map[string]*entities.Role
	perms  map[string]*entities.Permission
	seq    uint64
	userRoles       map[string][]string
	userPermissions map[string][]string
}

func NewInMemoryUserRepo() driven.UserRepository {
	return &InMemoryUserRepo{
		users:           make(map[string]*entities.User),
		roles:           make(map[string]*entities.Role),
		perms:           make(map[string]*entities.Permission),
		userRoles:       make(map[string][]string),
		userPermissions: make(map[string][]string),
	}
}

func (r *InMemoryUserRepo) nextUint() uint {
	r.mu.Lock()
	r.seq++
	id := uint(r.seq)
	r.mu.Unlock()
	return id
}

func (r *InMemoryUserRepo) cloneUser(u *entities.User) *entities.User {
	cp := *u
	cp.Roles = append([]*entities.Role(nil), u.Roles...)
	cp.Permissions = append([]*entities.Permission(nil), u.Permissions...)
	return &cp
}

func (r *InMemoryUserRepo) FindByID(_ context.Context, id valueobjects.ID) (*entities.User, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	u, ok := r.users[id.String()]
	if !ok {
		return nil, nil
	}
	return r.cloneUser(u), nil
}

func (r *InMemoryUserRepo) FindAll(_ context.Context) ([]*entities.User, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	out := make([]*entities.User, 0, len(r.users))
	for _, u := range r.users {
		out = append(out, r.cloneUser(u))
	}
	return out, nil
}

func (r *InMemoryUserRepo) FindPaginated(
	_ context.Context,
	page, perPage int,
) ([]*entities.User, valueobjects.PaginationMeta, error) {
	all, _ := r.FindAll(context.Background())
	return paginate(all, page, perPage)
}

func (r *InMemoryUserRepo) Create(_ context.Context, u *entities.User) (*entities.User, error) {
	if u.ID.IsZero() {
		u.ID = valueobjects.NewIDUint(r.nextUint())
	}
	if u.CreatedAt.IsZero() {
		u.CreatedAt = time.Now()
		u.UpdatedAt = u.CreatedAt
	}
	r.mu.Lock()
	r.users[u.ID.String()] = r.cloneUser(u)
	r.mu.Unlock()
	return r.cloneUser(u), nil
}

func (r *InMemoryUserRepo) Update(_ context.Context, u *entities.User) (*entities.User, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, ok := r.users[u.ID.String()]; !ok {
		return nil, fmt.Errorf("user not found: %s", u.ID.String())
	}
	u.UpdatedAt = time.Now()
	r.users[u.ID.String()] = r.cloneUser(u)
	return r.cloneUser(u), nil
}

func (r *InMemoryUserRepo) Delete(_ context.Context, id valueobjects.ID) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	delete(r.users, id.String())
	return nil
}

func (r *InMemoryUserRepo) FindBy(_ context.Context, _ map[string]interface{}) ([]*entities.User, error) {
	return r.FindAll(context.Background())
}

func (r *InMemoryUserRepo) FirstOrCreate(
	ctx context.Context,
	conds map[string]interface{},
	defaults map[string]interface{},
) (*entities.User, bool, error) {
	if email, ok := conds["email"].(valueobjects.Email); ok {
		u, err := r.FindByEmail(ctx, email)
		if err == nil && u != nil {
			return u, false, nil
		}
	}
	name := "Unknown"
	if v, ok := defaults["name"]; ok {
		name = fmt.Sprintf("%v", v)
	}
	email, _ := valueobjects.NewEmail("unknown@example.com")
	if v, ok := defaults["email"]; ok {
		email, _ = valueobjects.NewEmail(fmt.Sprintf("%v", v))
	}
	u := entities.NewUser(name, email)
	created, err := r.Create(ctx, u)
	return created, true, err
}

func (r *InMemoryUserRepo) FindByEmail(_ context.Context, email valueobjects.Email) (*entities.User, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	for _, u := range r.users {
		if u.Email.String() == email.String() {
			return r.cloneUser(u), nil
		}
	}
	return nil, nil
}

func (r *InMemoryUserRepo) FindByProvider(_ context.Context, provider, providerID string) (*entities.User, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	for _, u := range r.users {
		if u.Provider == provider && u.ProviderID == providerID {
			return r.cloneUser(u), nil
		}
	}
	return nil, nil
}

func (r *InMemoryUserRepo) AttachRole(_ context.Context, userID, roleID valueobjects.ID) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	key := userID.String()
	r.userRoles[key] = append(r.userRoles[key], roleID.String())
	return nil
}

func (r *InMemoryUserRepo) AttachPermission(_ context.Context, userID, permID valueobjects.ID) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	key := userID.String()
	r.userPermissions[key] = append(r.userPermissions[key], permID.String())
	return nil
}
