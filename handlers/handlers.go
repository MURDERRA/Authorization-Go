package handlers

import (
	"errors"
	"net/http"
	"time"

	"auth-service/client"
	"auth-service/config"
	"auth-service/logger"
	"auth-service/models"
	"auth-service/utils"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

// AppContext содержит контекст приложения, доступный всем обработчикам
type AppContext struct {
	Config    *config.Config
	SecretKey string
	Algorithm string
	TokenTTL  time.Duration
	Logger    *logger.ColorfulLogger
}

// Claims представляет данные, хранящиеся в JWT токене
type Claims struct {
	Username string `json:"sub"`
	AgencyID int    `json:"ngy"`
	jwt.RegisteredClaims
}

// createToken создает новый JWT токен
func (ctx *AppContext) createToken(username string, agencyID int) (string, error) {
	expirationTime := time.Now().Add(ctx.TokenTTL)

	claims := &Claims{
		Username: username,
		AgencyID: agencyID,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expirationTime),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.GetSigningMethod(ctx.Algorithm), claims)
	tokenString, err := token.SignedString([]byte(ctx.SecretKey))
	if err != nil {
		ctx.Logger.Error("Ошибка подписи токена: %v", err)
		return "", err
	}

	ctx.Logger.Info("Создан новый токен для пользователя '%s' (Agency ID: %d), срок действия до: %s",
		username, agencyID, expirationTime.Format(time.RFC3339))

	return tokenString, nil
}

// ValidateToken проверяет токен и пользователя в базе данных
func (ctx *AppContext) ValidateToken(tokenString string) (*Claims, error) {
	// Сначала разбираем и проверяем токен
	claims, err := ctx.parseAndValidateToken(tokenString)
	if err != nil {
		ctx.Logger.Error("Ошибка при проверке токена: %v", err)
		return nil, errors.New("некорректный токен: " + err.Error())
	}

	// Проверяем наличие имени пользователя в токене
	if claims.Username == "" {
		ctx.Logger.Error("Ошибка при проверке токена: отсутствует имя пользователя")
		return nil, errors.New("некорректный токен: отсутствует имя пользователя")
	}

	// Проверяем ID агентства
	if claims.AgencyID < 0 {
		ctx.Logger.Error("Ошибка при проверке токена: отсутствует ID агентства")
		return nil, errors.New("некорректный токен: отсутствует ID агентства")
	}

	// Проверяем срок действия токена
	if time.Now().After(claims.ExpiresAt.Time) {
		ctx.Logger.Error("Ошибка при проверке токена: токен истек (%s)", claims.ExpiresAt.Time)
		return nil, errors.New("токен истек")
	}

	// Получаем информацию о пользователе из БД
	apiClient := client.NewAPIClient(ctx.Config)
	user, err := apiClient.GetUser(claims.Username)
	if err != nil {
		ctx.Logger.Error("Ошибка проверки токена: пользователь '%s' не найден", claims.Username)
		return nil, errors.New("пользователь не найден")
	}

	// Проверяем соответствие токена сохраненному в БД
	if user.JWTToken != tokenString {
		ctx.Logger.Error("Ошибка проверки токена: токен не соответствует сохраненному в БД для пользователя '%s'", claims.Username)
		return nil, errors.New("токен не соответствует сохраненному в БД")
	}

	ctx.Logger.Info("Токен успешно проверен для пользователя '%s'", claims.Username)
	return claims, nil
}

// Login обрабатывает запрос на аутентификацию
// @Summary Аутентификация пользователя
// @Description Выполняет вход в систему и возвращает JWT токен
// @Tags auth
// @Accept json
// @Produce json
// @Param credentials body models.User true "Учетные данные пользователя"
// @Success 200 {object} models.TokenResponse
// @Failure 400 {object} models.ErrorResponse
// @Failure 401 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /login [post]
func Login(appCtx *AppContext) gin.HandlerFunc {
	return func(c *gin.Context) {
		var userData models.User
		if err := c.ShouldBindJSON(&userData); err != nil {
			appCtx.Logger.Warn("Попытка входа с некорректными данными запроса")
			c.JSON(http.StatusBadRequest, models.ErrorResponse{Error: "Некорректные данные запроса"})
			return
		}

		appCtx.Logger.Info("Попытка входа пользователя: %s", userData.Username)

		apiClient := client.NewAPIClient(appCtx.Config)
		user, err := apiClient.GetUser(userData.Username)
		if err != nil {
			appCtx.Logger.Error("Ошибка входа: пользователь '%s' не найден", userData.Username)
			c.JSON(http.StatusUnauthorized, models.ErrorResponse{Error: "Пользователь не найден"})
			return
		}

		if !utils.VerifyPassword(userData.Password, user.Password) {
			appCtx.Logger.Error("Ошибка входа: неверный пароль для пользователя '%s'", userData.Username)
			c.JSON(http.StatusUnauthorized, models.ErrorResponse{Error: "Неверный пароль"})
			return
		}

		token, err := appCtx.createToken(user.Login, user.AgencyID)
		if err != nil {
			appCtx.Logger.Error("Ошибка создания токена для пользователя '%s': %v", userData.Username, err)
			c.JSON(http.StatusInternalServerError, models.ErrorResponse{Error: "Ошибка создания токена"})
			return
		}

		if err := apiClient.UpdateToken(userData.Username, token); err != nil {
			appCtx.Logger.Error("Ошибка обновления токена в БД для пользователя '%s': %v", userData.Username, err)
			c.JSON(http.StatusInternalServerError, models.ErrorResponse{Error: "Ошибка обновления токена в БД"})
			return
		}

		appCtx.Logger.Info("Успешный вход пользователя: %s", userData.Username)
		c.JSON(http.StatusOK, models.TokenResponse{
			AccessToken: token,
			TokenType:   "bearer",
		})
	}
}

// CreateToken обрабатывает запрос на создание токена (JWT совместимый)
// @Summary Создание токена (JWT)
// @Description Создает токен доступа в формате JWT
// @Tags auth
// @Accept x-www-form-urlencoded
// @Produce json
// @Param username formData string true "Имя пользователя"
// @Param password formData string true "Пароль"
// @Success 200 {object} models.TokenResponse
// @Failure 400 {object} models.ErrorResponse
// @Failure 401 {object} models.ErrorResponse
// @Router /token/create [post]
func CreateToken(appCtx *AppContext) gin.HandlerFunc {
	return func(c *gin.Context) {
		var form struct {
			Username string `form:"username" binding:"required"`
			Password string `form:"password" binding:"required"`
		}

		if err := c.ShouldBind(&form); err != nil {
			appCtx.Logger.Warn("Попытка создания токена с некорректными данными формы")
			c.JSON(http.StatusBadRequest, models.ErrorResponse{Error: "Некорректные данные формы"})
			return
		}

		appCtx.Logger.Info("Попытка создания токена для пользователя: %s", form.Username)

		apiClient := client.NewAPIClient(appCtx.Config)
		user, err := apiClient.GetUser(form.Username)
		if err != nil {
			appCtx.Logger.Error("Ошибка создания токена: пользователь '%s' не найден", form.Username)
			c.JSON(http.StatusUnauthorized, models.ErrorResponse{Error: "Неверное имя пользователя или пароль"})
			return
		}

		if !utils.VerifyPassword(form.Password, user.Password) {
			appCtx.Logger.Error("Ошибка создания токена: неверный пароль для пользователя '%s'", form.Username)
			c.JSON(http.StatusUnauthorized, models.ErrorResponse{Error: "Неверное имя пользователя или пароль"})
			return
		}

		token, err := appCtx.createToken(user.Login, user.AgencyID)
		if err != nil {
			appCtx.Logger.Error("Ошибка создания токена для пользователя '%s': %v", form.Username, err)
			c.JSON(http.StatusBadRequest, models.ErrorResponse{Error: "Ошибка создания токена"})
			return
		}

		if err := apiClient.UpdateToken(form.Username, token); err != nil {
			appCtx.Logger.Error("Ошибка обновления токена в БД для пользователя '%s': %v", form.Username, err)
			c.JSON(http.StatusBadRequest, models.ErrorResponse{Error: "Ошибка обновления токена в БД"})
			return
		}

		appCtx.Logger.Info("Успешно создан токен для пользователя: %s", form.Username)
		c.JSON(http.StatusOK, models.TokenResponse{
			AccessToken: token,
			TokenType:   "bearer",
		})
	}
}

// parseAndValidateToken разбирает и проверяет JWT токен
func (ctx *AppContext) parseAndValidateToken(tokenString string) (*Claims, error) {
	ctx.Logger.Debug("Проверка токена: %s...", tokenString[:10]+"...")

	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (any, error) {
		if token.Method.Alg() != ctx.Algorithm {
			return nil, errors.New("некорректный алгоритм подписи")
		}
		return []byte(ctx.SecretKey), nil
	})

	if err != nil {
		ctx.Logger.Error("Ошибка при разборе токена: %v", err)
		return nil, err
	}

	if claims, ok := token.Claims.(*Claims); ok && token.Valid {
		ctx.Logger.Info("Токен действителен для пользователя: %s (Agency ID: %d)",
			claims.Username, claims.AgencyID)
		return claims, nil
	}

	ctx.Logger.Info("Токен недействителен")
	return nil, errors.New("некорректный токен")
}

// VerifyToken обрабатывает запрос на проверку токена
// @Summary Проверка токена
// @Description Проверяет валидность JWT токена
// @Tags auth
// @Accept json
// @Produce json
// @Success 200 {object} models.TokenVerifyResponse
// @Failure 400 {object} models.ErrorResponse
// @Failure 401 {object} models.ErrorResponse
// @Failure 404 {object} models.ErrorResponse
// @Security Bearer
// @Router /token/verify [post]
func VerifyToken(appCtx *AppContext) gin.HandlerFunc {
	return func(c *gin.Context) {
		username := c.GetString("username")
		agencyID := c.GetInt("agencyID")

		appCtx.Logger.Debug("Запрос на проверку токена для пользователя: %s", username)

		c.JSON(http.StatusOK, models.TokenVerifyResponse{
			Valid:    true,
			Username: username,
			AgencyID: agencyID,
		})
	}
}

// RefreshToken обрабатывает запрос на обновление токена
// @Summary Обновление токена
// @Description Обновляет JWT токен
// @Tags auth
// @Accept json
// @Produce json
// @Success 200 {object} models.TokenResponse
// @Failure 400 {object} models.ErrorResponse
// @Failure 404 {object} models.ErrorResponse
// @Security Bearer
// @Router /token/refresh [post]
// RefreshToken обрабатывает запрос на обновление токена
func RefreshToken(appCtx *AppContext) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Получаем данные из контекста, установленные middleware
		username := c.GetString("username")
		agencyID := c.GetInt("agencyID")

		appCtx.Logger.Debug("Запрос на обновление токена для пользователя: %s", username)

		// Создаем новый токен и обновляем в БД
		newToken, err := appCtx.createToken(username, agencyID)
		if err != nil {
			appCtx.Logger.Error("Ошибка создания нового токена для пользователя '%s': %v", username, err)
			c.JSON(http.StatusBadRequest, models.ErrorResponse{Error: "Ошибка создания токена"})
			return
		}

		apiClient := client.NewAPIClient(appCtx.Config)
		if err := apiClient.UpdateToken(username, newToken); err != nil {
			appCtx.Logger.Error("Ошибка обновления токена в БД для пользователя '%s': %v", username, err)
			c.JSON(http.StatusBadRequest, models.ErrorResponse{Error: "Ошибка обновления токена в БД"})
			return
		}

		appCtx.Logger.Info("Токен успешно обновлен для пользователя '%s'", username)
		c.JSON(http.StatusOK, models.TokenResponse{
			AccessToken: newToken,
			TokenType:   "bearer",
		})
	}
}

// Logout обрабатывает запрос на выход из системы
// @Summary Выход из системы
// @Description Выполняет выход пользователя и удаляет токен
// @Tags auth
// @Accept json
// @Produce json
// @Success 200 {object} models.Message
// @Failure 400 {object} models.ErrorResponse
// @Failure 401 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Security Bearer
// @Router /logout [post]
// Logout обрабатывает запрос на выход из системы
func Logout(appCtx *AppContext) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Получаем данные из контекста, установленные middleware
		username := c.GetString("username")
		token := c.GetString("token")

		appCtx.Logger.Debug("Запрос на выход для пользователя: %s", username)

		apiClient := client.NewAPIClient(appCtx.Config)
		if err := apiClient.DeleteToken(username, token); err != nil {
			appCtx.Logger.Error("Ошибка удаления токена из БД для пользователя '%s': %v", username, err)
			c.JSON(http.StatusInternalServerError, models.ErrorResponse{Error: "Ошибка удаления токена из БД"})
			return
		}

		appCtx.Logger.Info("Успешный выход пользователя: %s", username)
		c.JSON(http.StatusOK, models.Message{Message: "Успешный выход из системы"})
	}
}
