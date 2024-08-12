package auth

import (
	"context"
	"errors"
	"fmt"
	"golang.org/x/crypto/bcrypt"
	"log/slog"
	"ssov2/internal/domain/models"
	"ssov2/internal/lib/jwt"
	"time"
)

type Auth struct {
	log         *slog.Logger
	newUser     NewUser
	provideUser ProvideUser
	tokenTTL    time.Duration
}

var (
	ErrInvalidCredentials = errors.New("invalid credentials")
)

type NewUser interface {
	SaveUser(
		ctx context.Context,
		email string,
		passwdHash []byte,
	) (userID int64, err error)
}

type ProvideUser interface {
	User(ctx context.Context, email string) (user models.User, err error)
}

func New(log *slog.Logger, tokenTTL time.Duration, newUser NewUser, provideUser ProvideUser) *Auth {
	return &Auth{log: log, tokenTTL: tokenTTL, newUser: newUser, provideUser: provideUser}
}

func (a *Auth) RegisterNewUser(ctx context.Context, email string, password string) (userID int64, err error) {
	const op = "Auth.RegisterNewUser"

	log := a.log.With(
		slog.String("op", op),
		slog.String("email", email),
	)

	e, err := a.provideUser.User(ctx, email)
	if err != nil {
		log.Error("failed to check user", err)
		return 0, fmt.Errorf("%s: %w", op, err)
	}

	if e.Email == email {
		log.Error("email already in use")
		return 0, fmt.Errorf("%s: %w", op, err)
	}

	log.Info("registering user")

	passwdHash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		log.Error("failed to generate password hash", err)
		return 0, fmt.Errorf("%s: %w", op, err)
	}

	uid, err := a.newUser.SaveUser(ctx, email, passwdHash)
	if err != nil {
		log.Error("failed to save user", err)
		return 0, fmt.Errorf("%s: %w", op, err)
	}

	return uid, nil
}

func (a *Auth) Login(ctx context.Context, email string, password string) (string, error) {
	const op = "Auth.Login"

	log := a.log.With(
		slog.String("op", op),
		slog.String("email", email),
	)

	log.Info("attempting to login user")

	user, err := a.provideUser.User(ctx, email)
	if err != nil {
		log.Error("failed to check user", err)
		return "", fmt.Errorf("%s: %w", op, ErrInvalidCredentials)
	}

	if err := bcrypt.CompareHashAndPassword(user.PassHash, []byte(password)); err != nil {
		a.log.Info("failed to compare password", err)

		return "", fmt.Errorf("%s: %w", op, ErrInvalidCredentials)
	}

	token, err := jwt.NewToken(user, a.tokenTTL)
	if err != nil {
		a.log.Error("failed to generate token", err)

		return "", fmt.Errorf("%s: %w", op, err)
	}

	return token, nil
}

func (a *Auth) IsAdmin(ctx context.Context, userID int64) (bool, error) {
	return true, nil
}
