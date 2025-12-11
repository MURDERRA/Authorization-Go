package client

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"

	"auth-service/config"
	"auth-service/models"
)

// APIClient предоставляет методы для взаимодействия с локальным API
type APIClient struct {
	BaseURL     string
	ServiceName string
	HTTPClient  *http.Client
}

// NewAPIClient создает новый экземпляр клиента API
func NewAPIClient(cfg *config.Config) *APIClient {
	return &APIClient{
		BaseURL:     cfg.LocalAPIURL,
		ServiceName: cfg.ServiceName,
		HTTPClient:  &http.Client{},
	}
}

// GetUser получает данные пользователя из БД
func (c *APIClient) GetUser(username string) (*models.UserData, error) {
	url := fmt.Sprintf("%s/get_user_data/?username=%s", c.BaseURL, username)

	request := map[string]string{
		"name": c.ServiceName,
	}

	reqBody, err := json.Marshal(request)
	if err != nil {
		return nil, fmt.Errorf("ошибка маршалинга запроса: %w", err)
	}

	log.Printf("Request to %s with body %s", url, string(reqBody))

	resp, err := c.HTTPClient.Post(url, "application/json", bytes.NewBuffer(reqBody))

	if err != nil {
		log.Printf("ошибка сетевого запроса: %v", err)
		return nil, fmt.Errorf("ошибка сетевого запроса: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		log.Printf("API вернул ошибку: %d - %s", resp.StatusCode, string(bodyBytes))
		return nil, fmt.Errorf("API вернул ошибку: %d - %s", resp.StatusCode, string(bodyBytes))
	}

	var response struct {
		Data models.UserData `json:"data"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		log.Printf("ошибка декодирования ответа: %v", err)
		return nil, fmt.Errorf("ошибка декодирования ответа: %w", err)
	}

	if response.Data.Login == "" {
		log.Printf("пользователь не найден. %v", response.Data)
		return nil, errors.New("пользователь не найден")
	}

	return &response.Data, nil
}

// UpdateToken обновляет токен пользователя в БД
func (c *APIClient) UpdateToken(username, token string) error {
	url := fmt.Sprintf("%s/token/update", c.BaseURL)

	request := models.LocalAPIRequest{}
	request.MicroName.Name = c.ServiceName
	request.TokenData.Login = username
	request.TokenData.JWTToken = token

	reqBody, err := json.Marshal(request)
	if err != nil {
		return fmt.Errorf("ошибка маршалинга запроса: %w", err)
	}

	resp, err := c.HTTPClient.Post(url, "application/json", bytes.NewBuffer(reqBody))
	if err != nil {
		return fmt.Errorf("ошибка сетевого запроса: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("API вернул ошибку: %d - %s", resp.StatusCode, string(bodyBytes))
	}

	return nil
}

// DeleteToken удаляет токен пользователя из БД
func (c *APIClient) DeleteToken(username, token string) error {
	url := fmt.Sprintf("%s/token/delete", c.BaseURL)

	request := models.LocalAPIRequest{}
	request.MicroName.Name = c.ServiceName
	request.TokenData.Login = username
	request.TokenData.JWTToken = token

	reqBody, err := json.Marshal(request)
	if err != nil {
		return fmt.Errorf("ошибка маршалинга запроса: %w", err)
	}

	req, err := http.NewRequest(http.MethodDelete, url, bytes.NewBuffer(reqBody))
	if err != nil {
		return fmt.Errorf("ошибка создания запроса: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return fmt.Errorf("ошибка сетевого запроса: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("API вернул ошибку: %d - %s", resp.StatusCode, string(bodyBytes))
	}

	return nil
}
