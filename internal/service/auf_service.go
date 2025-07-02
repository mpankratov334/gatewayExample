package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"gateway/internal/config"
	"gateway/internal/dto"
	repo2 "gateway/internal/repo"
	pb "gateway/pkg/auf"
	"gateway/pkg/validator"
	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"go.uber.org/zap"
	"golang.org/x/crypto/bcrypt"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"net"
	"time"
)

type AuthService struct {
	pb.UnimplementedAuthServiceServer
	Repo       repo2.Repository
	log        *zap.SugaredLogger
	Service    *grpc.Server
	listener   net.Listener
	_secretKey string
}

func NewAuthService(repo repo2.Repository, logger *zap.SugaredLogger, cfg config.Rest) *AuthService {
	lis, err := net.Listen("tcp", ":"+cfg.GRPCport)
	if err != nil {
		logger.Fatalf("NewService failed to listen: %v", err)
	}
	server := grpc.NewServer()
	auf := &AuthService{
		Repo:       repo,
		log:        logger,
		Service:    server,
		listener:   lis,
		_secretKey: cfg.SecretKey,
	}
	pb.RegisterAuthServiceServer(server, auf)
	return auf
}

func (s *AuthService) ListenAndServe() error {
	err := s.Service.Serve(s.listener)
	return err
}

func (s *AuthService) Register(ctx context.Context, req *pb.RegisterRequest) (*pb.RegisterResponse, error) {
	// sha hash
	// repo post
	// send pb response
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to hash password: %v", err)
	}

	err = s.Repo.CreateUser(ctx, req.Username, string(hashedPassword))
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			// Код ошибки unique_violation
			return nil, status.Errorf(codes.AlreadyExists, "username %s already exists", req.Username)
		}
		return nil, status.Errorf(codes.Internal, "failed to create user: %v", err)
	}

	return &pb.RegisterResponse{
		Message: fmt.Sprintf("User %s created successfully", req.Username),
	}, nil
}

func (s *AuthService) RegisterHandler(ctx *fiber.Ctx) error {
	var req LoginRequest

	// Десериализация JSON-запроса
	if err := json.Unmarshal(ctx.Body(), &req); err != nil {
		s.log.Error("Invalid request body", zap.Error(err))
		return dto.BadResponseError(ctx, dto.FieldBadFormat, "Invalid request body")
	}

	// Валидация входных данных
	if vErr := validator.Validate(ctx.Context(), req); vErr != nil {
		return dto.BadResponseError(ctx, dto.FieldIncorrect, vErr.Error())
	}

	// 2. Использовать gRPC-клиент для вызова AuthService
	grpcResp, err := s.Register(ctx.Context(), &pb.RegisterRequest{
		Username: req.Username,
		Password: req.Password,
	})
	if err != nil {
		s.log.Error("Failed to register user", zap.Error(err))
		return dto.InternalServerError(ctx)
	}

	response := dto.Response{
		Status: "success",
		Data:   grpcResp.Message,
	}

	return ctx.Status(fiber.StatusOK).JSON(response)
}

func (s *AuthService) Login(ctx context.Context, req *pb.LoginRequest) (*pb.LoginResponse, error) {
	user, err := s.Repo.GetUser(ctx, req.Username)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, status.Errorf(codes.NotFound, "user %s not found", req.Username)
		}
		return nil, status.Errorf(codes.Internal, "failed to get user: %v", err)
	}

	err = bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password))
	if err != nil {
		return nil, status.Errorf(codes.Unauthenticated, "invalid password")
	}

	claims := jwt.RegisteredClaims{
		ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(s._secretKey))
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to sign token: %v", err)
	}

	return &pb.LoginResponse{
		Token: tokenString,
	}, nil
}

func (s *AuthService) LoginHandler(ctx *fiber.Ctx) error {
	var req LoginRequest

	// Десериализация JSON-запроса
	if err := json.Unmarshal(ctx.Body(), &req); err != nil {
		s.log.Error("Invalid request body", zap.Error(err))
		return dto.BadResponseError(ctx, dto.FieldBadFormat, "Invalid request body")
	}

	// Валидация входных данных
	if vErr := validator.Validate(ctx.Context(), req); vErr != nil {
		return dto.BadResponseError(ctx, dto.FieldIncorrect, vErr.Error())
	}

	// 2. Использовать gRPC-клиент для вызова AuthService
	grpcResp, err := s.Login(ctx.Context(), &pb.LoginRequest{
		Username: req.Username,
		Password: req.Password,
	})
	if err != nil {
		s.log.Error("Failed to register user", zap.Error(err))
		return dto.InternalServerError(ctx)
	}

	response := dto.Response{
		Status: "success",
		Data:   grpcResp.Token,
	}

	return ctx.Status(fiber.StatusOK).JSON(response)
}

func (s *AuthService) ValidateToken(ctx context.Context, req *pb.Token) (*pb.IsValid, error) {
	if req.Token == "" {
		return nil, status.Error(codes.InvalidArgument, "Token is required")
	}
	token, err := jwt.ParseWithClaims(
		req.Token,
		&jwt.RegisteredClaims{},
		func(token *jwt.Token) (interface{}, error) {
			return s._secretKey, nil
		},
	)

	if err != nil || !token.Valid {
		return &pb.IsValid{IsValid: false}, nil
	}
	return &pb.IsValid{IsValid: true}, nil
}
