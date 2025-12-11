// Файл: middleware/auth.go
package middleware

import (
	"net/http"
	"strings"

	"auth-service/handlers"

	"github.com/gin-gonic/gin"
)

// AuthMiddleware проверяет авторизацию пользователя
func AuthMiddleware(appCtx *handlers.AppContext) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Получаем токен из заголовка Authorization
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Отсутствует заголовок авторизации"})
			c.Abort()
			return
		}

		// Извлекаем токен из заголовка
		var token string
		if strings.HasPrefix(strings.ToLower(authHeader), "bearer ") {
			token = strings.TrimPrefix(strings.ToLower(authHeader), "bearer ")
		} else {
			token = authHeader
		}

		appCtx.Logger.Debug("Проверка токена из заголовка: %s...", token[:10]+"...")

		// Проверяем токен напрямую через ValidateToken
		claims, err := appCtx.ValidateToken(token)
		if err != nil {
			appCtx.Logger.Error("Ошибка при проверке токена: %v", err)
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Недействительный токен: " + err.Error()})
			c.Abort()
			return
		}

		// Добавляем данные пользователя в контекст для использования в обработчиках
		c.Set("username", claims.Username)
		c.Set("agencyID", claims.AgencyID)
		c.Set("token", token)

		appCtx.Logger.Info("Успешная аутентификация пользователя: %s (Agency ID: %d)",
			claims.Username, claims.AgencyID)

		// Продолжаем выполнение цепочки middleware
		c.Next()
	}
}
