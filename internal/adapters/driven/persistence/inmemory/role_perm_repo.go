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

type InMemoryRoleRepo struct {
	mu       sync.RWMutex
	roles    map[string]*entities.Role
	perms    map[string]*entities.Permission
	rolePerm map[string][]string
	seq      uint64
}

func NewInMemoryRoleRepo() driven.RoleRepository {
	return &InMemoryRoleRepo{
		roles:    make(map[string]*entities.Role),
		perms:    make(map[string]*entities.Permission),
		rolePerm: make(map[string][]string),
	}
}

func (r *InMemoryRoleRepo) nextUint() uint {
	r.mu.Lock()
	r.seq++
	id := uint(r.seq)
	r.mu.Unlock()
	return id
}

func (r *InMemoryRoleRepo) cloneRole(rl *entities.Role) *entities.Role {
	cp := *rl
	cp.Permissions = append([]*entities.Permission(nil), rl.Permissions...)
	return &cp
}

func (r *InMemoryRoleRepo) FindByID(_ context.Context, id valueobjects.ID) (*entities.Role, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	rl, ok := r.roles[id.String()]
	if !ok {
		return nil, nil
	}
	return r.cloneRole(rl), nil
}

func (r *InMemoryRoleRepo) FindAll(_ context.Context) ([]*entities.Role, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	out := make([]*entities.Role, 0, len(r.roles))
	for _, rl := range r.roles {
		out = append(out, r.cloneRole(rl))
	}
	return out, nil
}

func (r *InMemoryRoleRepo) FindPaginated(ctx context.Context, page, perPage int) ([]*entities.Role, valueobjects.PaginationMeta, error) {
	all, _ := r.FindAll(ctx)
	return paginate(all, page, perPage)
}

func (r *InMemoryRoleRepo) Create(_ context.Context, e *entities.Role) (*entities.Role, error) {
	if e.ID.IsZero() {
		e.ID = valueobjects.NewIDUint(r.nextUint())
	}
	if e.CreatedAt.IsZero() {
		now := time.Now()
		e.CreatedAt = now
		e.UpdatedAt = now
	}
	r.mu.Lock()
	r.roles[e.ID.String()] = r.cloneRole(e)
	r.mu.Unlock()
	return r.cloneRole(e), nil
}

func (r *InMemoryRoleRepo) Update(_ context.Context, e *entities.Role) (*entities.Role, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, ok := r.roles[e.ID.String()]; !ok {
		return nil, fmt.Errorf("role not found: %s", e.ID.String())
	}
	e.UpdatedAt = time.Now()
	r.roles[e.ID.String()] = r.cloneRole(e)
	return r.cloneRole(e), nil
}

func (r *InMemoryRoleRepo) Delete(_ context.Context, id valueobjects.ID) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	delete(r.roles, id.String())
	return nil
}

func (r *InMemoryRoleRepo) FindBy(_ context.Context, _ map[string]interface{}) ([]*entities.Role, error) {
	return r.FindAll(context.Background())
}

func (r *InMemoryRoleRepo) FirstOrCreate(ctx context.Context, conds map[string]interface{}, defaults map[string]interface{}) (*entities.Role, bool, error) {
	if name, ok := conds["name"].(string); ok {
		found, err := r.FindByName(ctx, name)
		if err == nil && found != nil {
			return found, false, nil
		}
	}
	name := "new_role"
	if v, ok := defaults["name"]; ok {
		name = fmt.Sprintf("%v", v)
	}
	label := name
	if v, ok := defaults["label"]; ok {
		label = fmt.Sprintf("%v", v)
	}
	rl := entities.NewRole(name, label)
	created, err := r.Create(ctx, rl)
	return created, true, err
}

func (r *InMemoryRoleRepo) FindByName(_ context.Context, name string) (*entities.Role, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	for _, rl := range r.roles {
		if rl.Name == name {
			return r.cloneRole(rl), nil
		}
	}
	return nil, nil
}

func (r *InMemoryRoleRepo) AttachPermission(_ context.Context, roleID, permID valueobjects.ID) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	key := roleID.String()
	r.rolePerm[key] = append(r.rolePerm[key], permID.String())
	return nil
}

type InMemoryPermissionRepo struct {
	mu    sync.RWMutex
	perms map[string]*entities.Permission
	seq   uint64
}

func NewInMemoryPermissionRepo() driven.PermissionRepository {
	return &InMemoryPermissionRepo{
		perms: make(map[string]*entities.Permission),
	}
}

func (r *InMemoryPermissionRepo) nextUint() uint {
	r.mu.Lock()
	r.seq++
	id := uint(r.seq)
	r.mu.Unlock()
	return id
}

func (r *InMemoryPermissionRepo) FindByID(_ context.Context, id valueobjects.ID) (*entities.Permission, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	p, ok := r.perms[id.String()]
	if !ok {
		return nil, nil
	}
	cp := *p
	return &cp, nil
}

func (r *InMemoryPermissionRepo) FindAll(_ context.Context) ([]*entities.Permission, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	out := make([]*entities.Permission, 0, len(r.perms))
	for _, p := range r.perms {
		cp := *p
		out = append(out, &cp)
	}
	return out, nil
}

func (r *InMemoryPermissionRepo) FindPaginated(ctx context.Context, page, perPage int) ([]*entities.Permission, valueobjects.PaginationMeta, error) {
	all, _ := r.FindAll(ctx)
	return paginate(all, page, perPage)
}

func (r *InMemoryPermissionRepo) Create(_ context.Context, e *entities.Permission) (*entities.Permission, error) {
	if e.ID.IsZero() {
		e.ID = valueobjects.NewIDUint(r.nextUint())
	}
	if e.CreatedAt.IsZero() {
		now := time.Now()
		e.CreatedAt = now
		e.UpdatedAt = now
	}
	r.mu.Lock()
	cp := *e
	r.perms[e.ID.String()] = &cp
	r.mu.Unlock()
	cp2 := *e
	return &cp2, nil
}

func (r *InMemoryPermissionRepo) Update(_ context.Context, e *entities.Permission) (*entities.Permission, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, ok := r.perms[e.ID.String()]; !ok {
		return nil, fmt.Errorf("permission not found: %s", e.ID.String())
	}
	e.UpdatedAt = time.Now()
	cp := *e
	r.perms[e.ID.String()] = &cp
	cp2 := *e
	return &cp2, nil
}

func (r *InMemoryPermissionRepo) Delete(_ context.Context, id valueobjects.ID) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	delete(r.perms, id.String())
	return nil
}

func (r *InMemoryPermissionRepo) FindBy(_ context.Context, _ map[string]interface{}) ([]*entities.Permission, error) {
	return r.FindAll(context.Background())
}

func (r *InMemoryPermissionRepo) FirstOrCreate(ctx context.Context, conds map[string]interface{}, defaults map[string]interface{}) (*entities.Permission, bool, error) {
	if name, ok := conds["name"].(string); ok {
		found, err := r.FindByName(ctx, name)
		if err == nil && found != nil {
			return found, false, nil
		}
	}
	name := "new_perm"
	if v, ok := defaults["name"]; ok {
		name = fmt.Sprintf("%v", v)
	}
	action := "read"
	if v, ok := defaults["action"]; ok {
		action = fmt.Sprintf("%v", v)
	}
	resource := "resource"
	if v, ok := defaults["resource"]; ok {
		resource = fmt.Sprintf("%v", v)
	}
	p := entities.NewPermission(name, action, resource)
	created, err := r.Create(ctx, p)
	return created, true, err
}

func (r *InMemoryPermissionRepo) FindByName(_ context.Context, name string) (*entities.Permission, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	for _, p := range r.perms {
		if p.Name == name {
			cp := *p
			return &cp, nil
		}
	}
	return nil, nil
}
