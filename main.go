package main

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"log"
	"time"

	"auth-service/config"
	"auth-service/docs"
	"auth-service/handlers"
	"auth-service/logger"
	"auth-service/middleware"

	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

// @title Auth Service API
// @version 1.0
// @description API для аутентификации и управления токенами
// @BasePath /
// @securityDefinitions.apikey Bearer
// @in header
// @name Authorization

func generateSecretKey() string {
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		panic(err)
	}
	return hex.EncodeToString(bytes)
}

func main() {
	// Загрузка конфигурации
	cfg, err := config.LoadConfig("config.json")
	if err != nil {
		log.Fatalf("Ошибка загрузки конфигурации: %v", err)
	}
	logger := logger.NewColorfulLogger(cfg)

	// if !cfg.LogLevel {
	// 	gin.SetMode(gin.ReleaseMode)
	// }

	// И нициализация роутера Gin
	r := gin.Default()

	// Инициализация контекста приложения
	secretKey := generateSecretKey()
	appCtx := &handlers.AppContext{
		Config:    cfg,
		SecretKey: secretKey,
		Algorithm: "HS256",
		TokenTTL:  time.Hour * 24 * 7, // 7 дней
		Logger:    logger,
	}

	// Настройка Swagger
	docs.SwaggerInfo.Host = fmt.Sprintf("localhost:%d", cfg.ServerPort)
	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	// Настройка роутов
	r.POST("/login", handlers.Login(appCtx))
	r.POST("/token/create", handlers.CreateToken(appCtx))
	r.POST("/token/verify", middleware.AuthMiddleware(appCtx), handlers.VerifyToken(appCtx))
	r.POST("/token/refresh", middleware.AuthMiddleware(appCtx), handlers.RefreshToken(appCtx))
	r.POST("/logout", middleware.AuthMiddleware(appCtx), handlers.Logout(appCtx))

	// Запуск сервера
	serverAddr := fmt.Sprintf(":%d", cfg.ServerPort)
	logger.Debug("Сервер запущен на http://localhost%s", serverAddr)
	logger.Debug("Swagger UI доступен по адресу: http://localhost:%d/swagger/index.html", cfg.ServerPort)

	if err := r.Run(serverAddr); err != nil {
		logger.Error("Ошибка запуска сервера: %v", err)
	}
}
