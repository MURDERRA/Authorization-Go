// Файл: models/models.go
package models

// User представляет данные пользователя
// @Description Данные пользователя для аутентификации
type User struct {
	Username string `json:"username" binding:"required" example:"user123"`   // Логин пользователя
	Password string `json:"password" binding:"required" example:"pass123!!"` // Пароль пользователя
}

// TokenResponse представляет ответ с токеном доступа
// @Description Ответ с токеном доступа
type TokenResponse struct {
	AccessToken string `json:"access_token" example:"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."` // JWT токен доступа
	TokenType   string `json:"token_type" example:"bearer"`                                    // Тип токена (обычно "bearer")
}

// TokenVerify представляет запрос на проверку токена
// @Description Запрос на проверку токена
type TokenVerify struct {
	Token string `json:"token" binding:"required" example:"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."` // JWT токен для проверки
}

// TokenVerifyResponse представляет ответ на проверку токена
// @Description Ответ на проверку токена
type TokenVerifyResponse struct {
	Valid    bool   `json:"valid" example:"true"`       // Флаг валидности токена
	Username string `json:"username" example:"user123"` // Имя пользователя
	AgencyID int    `json:"agency_id" example:"42"`     // ID агентства
}

// Message представляет сообщение в ответе API
// @Description Сообщение в ответе API
type Message struct {
	Message string `json:"message" example:"Успешный выход из системы"` // Текст сообщения
}

// UserData представляет данные пользователя из БД
type UserData struct {
	Login    string `json:"login"`
	Password string `json:"password"`
	AgencyID int    `json:"agency_id"`
	JWTToken string `json:"jwt_token"`
}

// LocalAPIRequest представляет запрос к локальному API
type LocalAPIRequest struct {
	MicroName struct {
		Name string `json:"name"`
	} `json:"micro_name"`
	TokenData struct {
		Login    string `json:"login"`
		JWTToken string `json:"jwt_token"`
	} `json:"token_data"`
}

// ErrorResponse представляет структуру ответа с ошибкой
type ErrorResponse struct {
	Error string `json:"error" example:"Описание ошибки"`
}
