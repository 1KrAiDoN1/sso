package auth

import (
	"context"
	"errors"
	"fmt"
	"sso/internal/lib/jwt"
	"sso/internal/storage"
	"sso/internal/storage/postgres"
	"sso/pkg/logger"
	"time"

	"golang.org/x/crypto/bcrypt"
)

type Auth struct {
	log         *logger.Logger
	usrSaver    postgres.UserSaver
	usrProvider postgres.UserProvider
	appProvider postgres.AppProvider
	tokenTTL    time.Duration
}

var (
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrInvalidAppID       = errors.New("invalid app id")
	ErrUserExists         = errors.New("user already exists")
)

func New(
	log *logger.Logger,
	userSaver postgres.UserSaver,
	userProvider postgres.UserProvider,
	appProvider postgres.AppProvider,
	tokenTTL time.Duration,
) *Auth {
	return &Auth{
		log:         log,
		usrSaver:    userSaver,
		usrProvider: userProvider,
		appProvider: appProvider,
		tokenTTL:    tokenTTL,
	}
}

type AuthInterface interface {
	Login(ctx context.Context, email string, password string, appID int) (token string, err error)
	RegisterNewUser(ctx context.Context, email string, password string) (userID int64, err error)
	IsAdmin(ctx context.Context, userID int64) (bool, error)
}

func (a *Auth) Login(ctx context.Context, email string, password string, appID int) (string, error) {
	const op = "auth.Login"

	a.log.Info("Attempting to login user", email)

	// Получаем пользователя
	user, err := a.usrProvider.User(ctx, email)
	if err != nil {
		if errors.Is(err, storage.ErrUserNotFound) {
			a.log.Info(op, "User not found")
			// Возвращаем auth ошибку без обёртки
			return "", ErrInvalidCredentials
		}
		// Только неизвестные ошибки оборачиваем
		a.log.Error(op, err.Error())
		return "", fmt.Errorf("%s: failed to get user: %w", op, err)
	}

	// Проверяем пароль
	if err := bcrypt.CompareHashAndPassword(user.PassHash, []byte(password)); err != nil {
		a.log.Info(op, "Invalid password")
		return "", ErrInvalidCredentials
	}

	// Получаем приложение
	app, err := a.appProvider.App(ctx, appID)
	if err != nil {
		if errors.Is(err, storage.ErrAppNotFound) {
			a.log.Info(op, "App not found")
			// Возвращаем storage ошибку без обёртки
			return "", storage.ErrAppNotFound
		}
		a.log.Error(op, err.Error())
		return "", fmt.Errorf("%s: failed to get app: %w", op, err)
	}

	a.log.Info(op, "User logged in successfully")

	// Генерируем токен
	token, err := jwt.NewToken(user, app, a.tokenTTL)
	if err != nil {
		a.log.Error(op, err.Error())
		return "", fmt.Errorf("%s: failed to generate token: %w", op, err)
	}

	return token, nil
}

func (a *Auth) RegisterNewUser(ctx context.Context, email string, pass string) (int64, error) {
	const op = "auth.RegisterNewUser"

	a.log.Info(op, "Registering user")

	passHash, err := bcrypt.GenerateFromPassword([]byte(pass), bcrypt.DefaultCost)
	if err != nil {
		a.log.Error("Failed to generate password hash", err.Error())

		return 0, fmt.Errorf("%s: %w", op, err)
	}

	id, err := a.usrSaver.SaveUser(ctx, email, passHash)
	if err != nil {
		if errors.Is(err, storage.ErrUserExists) {
			a.log.Warn("User already exists", err.Error())

			return 0, fmt.Errorf("%s: %w", op, storage.ErrUserExists)
		}

		a.log.Error("Failed to save user", err.Error())

		return 0, fmt.Errorf("%s: %w", op, err)
	}

	a.log.Info(op, "User registered")

	return id, nil
}

func (a *Auth) IsAdmin(ctx context.Context, userID int64) (bool, error) {
	const op = "auth.IsAdmin"

	a.log.Info(op, "Checking if user is admin")

	isAdmin, err := a.usrProvider.IsAdmin(ctx, userID)
	if err != nil {
		if errors.Is(err, storage.ErrAppNotFound) {
			a.log.Warn("User not found", err.Error())

			return false, fmt.Errorf("%s: %w", op, ErrInvalidAppID)
		}
		return false, fmt.Errorf("%s: %w", op, err)
	}

	a.log.Info(op, "Checked if user is admin")

	return isAdmin, nil
}
