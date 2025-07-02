package service

import (
	"encoding/json"
	"gateway/internal/dto"
	repo2 "gateway/internal/repo"
	"gateway/pkg/validator"
	"github.com/gofiber/fiber/v2"
	"go.uber.org/zap"
)

// Слой бизнес-логики. Тут должна быть основная логика сервиса

// Service - интерфейс для бизнес-логики
type Service interface {
	CreateTask(ctx *fiber.Ctx) error
	GetTask(ctx *fiber.Ctx) error
}

type service struct {
	repo repo2.Repository
	log  *zap.SugaredLogger
}

// NewService - конструктор сервиса
func NewService(repo repo2.Repository, logger *zap.SugaredLogger) Service {
	return &service{
		repo: repo,
		log:  logger,
	}
}

// CreateTask - обработчик запроса на создание задачи
func (s *service) CreateTask(ctx *fiber.Ctx) error {
	var req TaskRequest

	// Десериализация JSON-запроса
	if err := json.Unmarshal(ctx.Body(), &req); err != nil {
		s.log.Error("Invalid request body", zap.Error(err))
		return dto.BadResponseError(ctx, dto.FieldBadFormat, "Invalid request body")
	}

	// Валидация входных данных
	if vErr := validator.Validate(ctx.Context(), req); vErr != nil {
		return dto.BadResponseError(ctx, dto.FieldIncorrect, vErr.Error())
	}

	// Вставка задачи в БД через репозиторий
	task := repo2.Task{
		Title:       req.Title,
		Description: req.Description,
	}
	taskID, err := s.repo.CreateTask(ctx.Context(), task)
	if err != nil {
		s.log.Error("Failed to insert task", zap.Error(err))
		return dto.InternalServerError(ctx)
	}

	// Формирование ответа
	response := dto.Response{
		Status: "success",
		Data:   map[string]int{"task_id": taskID},
	}

	return ctx.Status(fiber.StatusOK).JSON(response)
}

func (s *service) GetTask(ctx *fiber.Ctx) error {
	var taskReq TaskGetRequest

	// Получение id задачи
	taskReq = TaskGetRequest{ID: ctx.Params("id")}

	// Валидация входных данных
	if vErr := validator.Validate(ctx.Context(), taskReq); vErr != nil {
		return dto.BadResponseError(ctx, dto.FieldIncorrect, vErr.Error())
	}

	// Получение задачи из БД через репозиторий
	taskPtr, err := s.repo.GetTask(ctx.Context(), taskReq.ID)
	if err != nil {
		s.log.Error("Failed to get task", zap.Error(err))
		return dto.InternalServerError(ctx)
	}

	// Формирование ответа
	response := dto.Response{
		Status: "success",
		Data:   taskPtr,
	}
	return ctx.Status(fiber.StatusOK).JSON(response)
}
