package middleware

import (
	"aswadwk/messaging-task-go/internal/utils"
	"slices"
	"strings"

	"github.com/gofiber/fiber/v2"
)

func isAllowedPath(path string) bool {
	allowed := []string{
		"/docs",
		"/rapidoc",
		"/api-docs",
		"/v1/auth/login",
		"/auth/register",
		"/docs/openapi.json",
		"/v1/ocr",
		"/scalar",
	}
	return slices.Contains(allowed, path)
}

// AuthMiddleware adalah middleware untuk memverifikasi token JWT
func AuthMiddleware() fiber.Handler {
	return func(c *fiber.Ctx) error {
		// Skip middleware untuk endpoint yang tidak memerlukan autentikasi
		path := c.Path()

		if strings.HasPrefix(path, "/storage/") {
			return c.Next()
		}

		if isAllowedPath(path) {
			return c.Next()
		}

		// Ambil token dari header Authorization (case insensitive)
		authHeader := c.Get("Authorization")
		if authHeader == "" {
			authHeader = c.Get("authorization")
		}

		if authHeader == "" {
			return fiber.NewError(fiber.StatusUnauthorized, "Authorization header is missing")
		}

		// Format token harus "Bearer <token>" (case insensitive)
		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
			return fiber.NewError(fiber.StatusUnauthorized, "Invalid token format")
		}

		token := parts[1]

		claims, err := utils.VerifyToken(token)

		if err != nil {
			return fiber.NewError(fiber.StatusUnauthorized, "Unauthorized: Invalid token")
		}

		c.Locals("user", claims)

		return c.Next()
	}
}
