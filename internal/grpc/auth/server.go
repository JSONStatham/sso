package auth

import (
	"context"
	"errors"
	"fmt"

	ssov1 "github.com/JSONStatham/protos/gen/go/sso"
	"github.com/JSONStatham/sso/internal/services/auth"
	"github.com/JSONStatham/sso/internal/storage"
	"github.com/go-playground/validator/v10"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
)

var validate = validator.New()

type Auth interface {
	RegisterUser(ctx context.Context, email, password string) (int64, error)
	Login(ctx context.Context, email, password string, appID int64) (string, error)
	Logout(ctx context.Context, token string) error
	IsAdmin(ctx context.Context, userID int64) (bool, error)
}

type RegisterRequest struct {
	Email    string `validate:"required,email"`
	Password string `validate:"required,min=6"`
}

type LoginRequest struct {
	Email    string `validate:"required,email"`
	Password string `validate:"required,min=6"`
	AppID    int    `validate:"required"`
}

type serverAPI struct {
	ssov1.UnimplementedAuthServer
	auth Auth
}

func Register(gRPC *grpc.Server, auth Auth) {
	ssov1.RegisterAuthServer(gRPC, &serverAPI{auth: auth})
}

func (s *serverAPI) Register(ctx context.Context, req *ssov1.RegisterRequest) (*ssov1.RegisterResponse, error) {
	registerReq := RegisterRequest{
		Email:    req.GetEmail(),
		Password: req.GetPassword(),
	}

	if err := validate.Struct(registerReq); err != nil {
		validationErr := err.(validator.ValidationErrors)
		return nil, status.Error(codes.InvalidArgument, validationErr.Error())
	}

	userId, err := s.auth.RegisterUser(ctx, registerReq.Email, registerReq.Password)
	if userId == 0 || err != nil {
		if errors.Is(err, storage.ErrUserAlreadyExists) {
			return nil, status.Error(codes.AlreadyExists, "user already exists")
		}

		return nil, status.Error(codes.Internal, "failed to register user")
	}

	return &ssov1.RegisterResponse{
		UserId: 1,
	}, nil
}

func (s *serverAPI) Login(ctx context.Context, req *ssov1.LoginRequest) (*ssov1.LogingResponse, error) {
	loginReq := LoginRequest{
		Email:    req.GetEmail(),
		Password: req.GetPassword(),
		AppID:    int(req.GetAppId()),
	}

	if err := validate.Struct(loginReq); err != nil {
		validationErr := err.(validator.ValidationErrors)
		return nil, status.Error(codes.InvalidArgument, validationErr.Error())
	}

	token, err := s.auth.Login(ctx, loginReq.Email, loginReq.Password, int64(loginReq.AppID))
	if err != nil {
		if errors.Is(err, auth.ErrInvalidCredentials) {
			return nil, status.Error(codes.NotFound, "user not found")
		}

		return nil, status.Error(codes.Internal, fmt.Sprintf("failed to login: %v", err))
	}

	return &ssov1.LogingResponse{
		Token: token,
	}, nil
}

func (s *serverAPI) Logout(ctx context.Context, req *ssov1.LogoutRequest) (*emptypb.Empty, error) {
	return nil, nil
}

func (s *serverAPI) IsAdmin(ctx context.Context, req *ssov1.IsAdminRequest) (*ssov1.IsAdminResponse, error) {
	if req.GetUserId() == 0 {
		return nil, status.Error(codes.InvalidArgument, "user id is required")
	}

	isAdmin, err := s.auth.IsAdmin(ctx, req.GetUserId())
	if err != nil {
		if errors.Is(err, storage.ErrUserNotFound) {
			return nil, status.Error(codes.NotFound, "user not found")
		}

		return nil, status.Error(codes.Internal, fmt.Sprintf("failed to check if user is admin: %v", err))
	}

	return &ssov1.IsAdminResponse{
		IsAdmin: isAdmin,
	}, nil
}
