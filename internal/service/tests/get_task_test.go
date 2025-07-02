package tests

import (
	"encoding/json"
	"errors"
	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"go.uber.org/zap"
	"net/http"
	"testing"
	"trainee/internal/dto"
	"trainee/internal/repo"
	"trainee/internal/repo/mocks"
	"trainee/internal/service"
)

// TestGetTask - тестирование метода GetTask
func TestGetTask(t *testing.T) {
	// Создание мок-репозитория
	mockRepo := new(mocks.Repository)
	logger := zap.NewNop().Sugar() // Без вывода логов

	// Создаем экземпляр сервиса с мок-репозиторием
	s := service.NewService(mockRepo, logger)

	// Инициализируем Fiber-контекст
	app := fiber.New()
	app.Get("/tasks/:id", s.GetTask) // Маршрут с параметром :id

	t.Run("успешное получение задачи", func(t *testing.T) {
		taskID := "123"
		expectedTask := repo.Task{
			ID:          taskID,
			Title:       "Test Task",
			Description: "Test Description",
		}

		// Настройка мок-репозитория
		mockRepo.On("GetTask", mock.Anything, taskID).
			Return(expectedTask, nil). // Ожидаем успешный результат
			Once()

		// Формирование HTTP-запроса
		req, err := http.NewRequest("GET", "/tasks/"+taskID, nil)
		assert.NoError(t, err)

		// Выполнение запроса
		resp, err := app.Test(req)
		assert.NoError(t, err)
		assert.Equal(t, fiber.StatusOK, resp.StatusCode)

		// Декодирование ответа
		var response dto.Response
		err = json.NewDecoder(resp.Body).Decode(&response)
		assert.NoError(t, err)

		// Проверка структуры ответа
		assert.Equal(t, "success", response.Status)
		assert.NotNil(t, response.Data)

		// Преобразование и проверка данных задачи
		taskData, ok := response.Data.(map[string]interface{})
		assert.True(t, ok)
		assert.Equal(t, expectedTask.ID, taskData["id"])
		assert.Equal(t, expectedTask.Title, taskData["title"])
		assert.Equal(t, expectedTask.Description, taskData["description"])
	})

	t.Run("задача не найдена", func(t *testing.T) {
		invalidID := "invalid_id"

		// Настройка мока на возврат ошибки
		mockRepo.On("GetTask", mock.Anything, invalidID).
			Return(repo.Task{}, errors.New("not found")).
			Once()

		req, err := http.NewRequest("GET", "/tasks/"+invalidID, nil)
		assert.NoError(t, err)

		resp, err := app.Test(req)
		assert.NoError(t, err)
		assert.Equal(t, fiber.StatusBadRequest, resp.StatusCode)

		var response dto.Response
		err = json.NewDecoder(resp.Body).Decode(&response)
		assert.NoError(t, err)
		assert.Equal(t, "error", response.Status)
	})
}
