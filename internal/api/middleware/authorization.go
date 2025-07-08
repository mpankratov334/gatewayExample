package middleware

import (
	"fmt"
	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
	"strings"
)

// Обычный миддлваер

func Authorization(key string) fiber.Handler {
	return func(c *fiber.Ctx) error {
		// проверка токена вторизации
		authHeader := c.Get("Authorization")
		if authHeader == "" {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error":   "Unauthorized",
				"message": "Authorization header is missing",
			})
		}

		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error":   "Unauthorized",
				"message": "Authorization header is wrong format, use: Bearer <token>",
			})
		}
		token, err := jwt.ParseWithClaims(
			parts[1],
			&jwt.RegisteredClaims{},
			func(token *jwt.Token) (interface{}, error) {
				return []byte(key), nil
			},
		)

		if err != nil || !token.Valid {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error":   "Unauthorized",
				"message": "token is not valid",
			})
		}

		return c.Next()
	}
}
