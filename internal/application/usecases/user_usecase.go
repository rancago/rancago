package usecases

import (
	"context"
	"fmt"

	"github.com/rancago/framework/internal/domain/entities"
	"github.com/rancago/framework/internal/domain/valueobjects"
	derrors "github.com/rancago/framework/internal/domain/errors"
	"github.com/rancago/framework/internal/ports/driven"
	"github.com/rancago/framework/internal/ports/driving"
)

type UserInteractor struct {
	users      driven.UserRepository
	roles      driven.RoleRepository
	perms      driven.PermissionRepository
	socialite  driven.SocialitePort
}

func NewUserInteractor(
	users driven.UserRepository,
	roles driven.RoleRepository,
	perms driven.PermissionRepository,
	socialite driven.SocialitePort,
) driving.UserUseCase {
	return &UserInteractor{
		users:     users,
		roles:     roles,
		perms:     perms,
		socialite: socialite,
	}
}

func (uc *UserInteractor) Register(ctx context.Context, name, email, password string) (*entities.User, error) {
	e, err := valueobjects.NewEmail(email)
	if err != nil {
		return nil, derrors.New("user.register", derrors.ErrValidation, err.Error())
	}
	if name == "" {
		return nil, derrors.New("user.register", derrors.ErrValidation, "name is required")
	}
	existing, _ := uc.users.FindByEmail(ctx, e)
	if existing != nil {
		return nil, derrors.New("user.register", derrors.ErrAlreadyExists, fmt.Sprintf("email %s already registered", email))
	}
	u := entities.NewUser(name, e)
	u.PasswordHash = password
	return uc.users.Create(ctx, u)
}

func (uc *UserInteractor) Login(ctx context.Context, email, password string) (*entities.User, error) {
	e, err := valueobjects.NewEmail(email)
	if err != nil {
		return nil, derrors.New("user.login", derrors.ErrValidation, err.Error())
	}
	u, err := uc.users.FindByEmail(ctx, e)
	if err != nil || u == nil {
		return nil, derrors.New("user.login", derrors.ErrUnauthorized, "invalid credentials")
	}
	if u.PasswordHash != password {
		return nil, derrors.New("user.login", derrors.ErrUnauthorized, "invalid credentials")
	}
	return u, nil
}

func (uc *UserInteractor) FindByID(ctx context.Context, id valueobjects.ID) (*entities.User, error) {
	return uc.users.FindByID(ctx, id)
}

func (uc *UserInteractor) FindByEmail(ctx context.Context, email string) (*entities.User, error) {
	e, err := valueobjects.NewEmail(email)
	if err != nil {
		return nil, err
	}
	return uc.users.FindByEmail(ctx, e)
}

func (uc *UserInteractor) GetAuthURL(ctx context.Context, provider, state string) (string, error) {
	if uc.socialite == nil {
		return "", derrors.New("user.socialite", derrors.ErrInternal, "socialite not configured")
	}
	p, err := uc.socialite.Provider(provider)
	if err != nil {
		return "", derrors.New("user.socialite", derrors.ErrNotFound, err.Error())
	}
	return p.GetAuthURL(state), nil
}

func (uc *UserInteractor) LoginWithProvider(ctx context.Context, providerName, code string) (*entities.User, error) {
	if uc.socialite == nil {
		return nil, derrors.New("user.socialite", derrors.ErrInternal, "socialite not configured")
	}
	p, err := uc.socialite.Provider(providerName)
	if err != nil {
		return nil, derrors.New("user.socialite", derrors.ErrNotFound, err.Error())
	}
	tok, err := p.ExchangeCode(ctx, code)
	if err != nil {
		return nil, derrors.New("user.socialite", derrors.ErrUnauthorized, err.Error())
	}
	info, err := p.GetUserInfo(ctx, tok)
	if err != nil {
		return nil, derrors.New("user.socialite", derrors.ErrInternal, err.Error())
	}
	user, err := uc.users.FindByProvider(ctx, providerName, info.ID)
	if err == nil && user != nil {
		return user, nil
	}
	email, _ := valueobjects.NewEmail(info.Email)
	user = entities.NewUser(info.Name, email)
	user.Provider = providerName
	user.ProviderID = info.ID
	user.AvatarURL = info.AvatarURL
	user.EmailVerifiedAt = &user.CreatedAt
	return uc.users.Create(ctx, user)
}

func (uc *UserInteractor) AssignRole(ctx context.Context, userID valueobjects.ID, roleName string) error {
	role, err := uc.roles.FindByName(ctx, roleName)
	if err != nil {
		return derrors.New("user.assign_role", derrors.ErrNotFound, err.Error())
	}
	return uc.users.AttachRole(ctx, userID, role.ID)
}

func (uc *UserInteractor) HasPermission(ctx context.Context, userID valueobjects.ID, permName string) (bool, error) {
	u, err := uc.users.FindByID(ctx, userID)
	if err != nil {
		return false, err
	}
	return u.HasPermission(permName), nil
}
