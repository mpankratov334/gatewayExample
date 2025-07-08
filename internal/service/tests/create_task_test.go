package tests

import (
	"bytes"
	"encoding/json"
	"errors"
	"github.com/gofiber/fiber/v2"
	"net/http"
	"testing"
	"trainee/internal/dto"
	"trainee/internal/repo"
	"trainee/internal/repo/mocks"
	"trainee/internal/service"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"go.uber.org/zap"
)

// TestCreateTask - тестирование метода CreateTask
func TestCreateTask(t *testing.T) {
	// Создаем мок репозитория
	mockRepo := new(mocks.Repository)
	logger := zap.NewNop().Sugar() // Без вывода логов

	// Создаем экземпляр сервиса с мок-репозиторием
	s := service.NewService(mockRepo, logger)

	// Инициализируем Fiber-контекст
	app := fiber.New()
	app.Post("/tasks", s.CreateTask)

	// Вспомогательная функция для сброса мока
	resetMock := func() {
		mockRepo.ExpectedCalls = []*mock.Call{} // Очищаем ожидания
		mockRepo.Calls = []mock.Call{}          // Очищаем историю вызовов
	}

	t.Run("ошибка валидации входных данных", func(t *testing.T) {
		resetMock()
		body := []byte(`{}`) // Пустое тело, `title` обязателен

		req, err := http.NewRequest("POST", "/tasks", bytes.NewReader(body))
		assert.NoError(t, err)
		req.Header.Set("Content-Type", "application/json")

		resp, err := app.Test(req)
		assert.NoError(t, err)
		assert.Equal(t, fiber.StatusBadRequest, resp.StatusCode)

		var response dto.Response
		json.NewDecoder(resp.Body).Decode(&response)
		assert.Equal(t, "error", response.Status)
	})

	t.Run("успешное создание задачи", func(t *testing.T) {
		resetMock()
		task := service.TaskRequest{
			Title:       "Test Task",
			Description: "Test Description",
		}
		body, _ := json.Marshal(task)

		// Ожидаем, что вызов метода `CreateTask` в репозитории вернёт ID = 1
		mockRepo.On("CreateTask", mock.Anything, repo.Task{
			Title:       task.Title,
			Description: task.Description,
		}).Return(1, nil).Once()

		// Отправляем запрос
		req, err := http.NewRequest("POST", "/tasks", bytes.NewReader(body))
		assert.NoError(t, err)
		req.Header.Set("Content-Type", "application/json")

		// Выполняем запрос
		resp, err := app.Test(req)
		assert.NoError(t, err)
		assert.Equal(t, fiber.StatusOK, resp.StatusCode)

		// Проверяем ответ
		var response dto.Response
		json.NewDecoder(resp.Body).Decode(&response)
		assert.Equal(t, "success", response.Status)
	})

	t.Run("ошибка при создании задачи в БД", func(t *testing.T) {
		resetMock()
		task := service.TaskRequest{
			Title:       "Test Task",
			Description: "Test Description",
		}
		body, _ := json.Marshal(task)

		// Ожидаем ошибку при вставке в БД
		mockRepo.On("CreateTask", mock.Anything, repo.Task{
			Title:       task.Title,
			Description: task.Description,
		}).Return(0, errors.New("DB error")).Once()

		req, err := http.NewRequest("POST", "/tasks", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")

		resp, err := app.Test(req)
		assert.NoError(t, err)
		assert.Equal(t, fiber.StatusInternalServerError, resp.StatusCode)

		var response dto.Response
		json.NewDecoder(resp.Body).Decode(&response)
		assert.Equal(t, "error", response.Status)
	})
}
