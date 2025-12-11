// Package docs генерирует Swagger документацию.
package docs

// "github.com/swaggo/swag"

// @title Auth Service API
// @version 1.0
// @description API сервиса аутентификации
// @BasePath /
// @schemes http https
// @securityDefinitions.apikey Bearer
// @in header
// @name Authorization
type swaggerInfo struct {
	Version     string
	Host        string
	BasePath    string
	Schemes     []string
	Title       string
	Description string
}

// var SwaggerInfo = swaggerInfo{
// 	Version:     "1.0",
// 	Host:        "",
// 	BasePath:    "/",
// 	Schemes:     []string{"http", "https"},
// 	Title:       "Auth Service API",
// 	Description: "API сервиса аутентификации",
// }

// func init() {
// 	swag.Register(swag.Name, &swag.Spec{
// 		InfoInstanceName: "swagger",
// 		SwaggerTemplate:  SwaggerInfo.Description,
// 	})
// }
