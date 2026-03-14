// Package api реализует HTTP-клиент для взаимодействия с сервером GophKeeper.
package api

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"

	"github.com/google/uuid"

	"github.com/MarkelovSergey/goph-keeper/internal/model"
)

// ErrUnauthorized возникает при ответе сервера 401.
var ErrUnauthorized = errors.New("не авторизован")

// ErrNotFound возникает при ответе сервера 404.
var ErrNotFound = errors.New("запись не найдена")

// ErrConflict возникает при ответе сервера 409 (пользователь уже существует).
var ErrConflict = errors.New("пользователь уже существует")

// Client — HTTP-клиент для API сервера GophKeeper.
type Client struct {
	baseURL    string
	httpClient *http.Client
	token      string
}

// New создаёт новый Client с заданным адресом сервера.
// Если insecure=true, пропускает проверку TLS-сертификата.
func New(serverAddress string, insecure bool) *Client {
	transport := http.DefaultTransport
	if insecure {
		transport = &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true}, //nolint:gosec
		}
	}
	return &Client{
		baseURL:    serverAddress,
		httpClient: &http.Client{Transport: transport},
	}
}

// SetToken устанавливает JWT-токен для последующих запросов.
func (c *Client) SetToken(token string) {
	c.token = token
}

// authResponse представляет ответ на запросы регистрации/входа (токен в заголовке).
type authRequest struct {
	Login    string `json:"login"`
	Password string `json:"password"`
}

// Register регистрирует нового пользователя и возвращает JWT-токен.
func (c *Client) Register(ctx context.Context, login, password string) (string, error) {
	return c.authRequest(ctx, "/api/register", login, password)
}

// Login выполняет вход пользователя и возвращает JWT-токен.
func (c *Client) Login(ctx context.Context, login, password string) (string, error) {
	return c.authRequest(ctx, "/api/login", login, password)
}

func (c *Client) authRequest(ctx context.Context, path, login, password string) (string, error) {
	body := authRequest{Login: login, Password: password}
	resp, err := c.doJSON(ctx, http.MethodPost, path, body)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	switch resp.StatusCode {
	case http.StatusOK:
		token := resp.Header.Get("Authorization")
		if len(token) > 7 {
			token = token[7:] // убираем "Bearer "
		}
		return token, nil
	case http.StatusConflict:
		return "", ErrConflict
	case http.StatusUnauthorized:
		return "", ErrUnauthorized
	default:
		return "", readError(resp)
	}
}

// CreateCredential создаёт новую запись учётных данных на сервере.
func (c *Client) CreateCredential(ctx context.Context, credType model.CredentialType, name, metadata string, data []byte) (*model.Credential, error) {
	body := map[string]any{
		"type":     credType,
		"name":     name,
		"metadata": metadata,
		"data":     data,
	}
	resp, err := c.doJSON(ctx, http.MethodPost, "/api/credentials", body)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		return nil, readError(resp)
	}

	var cred model.Credential
	if err := json.NewDecoder(resp.Body).Decode(&cred); err != nil {
		return nil, err
	}
	return &cred, nil
}

// ListCredentials возвращает список всех записей текущего пользователя.
func (c *Client) ListCredentials(ctx context.Context) ([]*model.Credential, error) {
	resp, err := c.doJSON(ctx, http.MethodGet, "/api/credentials", nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusUnauthorized {
		return nil, ErrUnauthorized
	}
	if resp.StatusCode != http.StatusOK {
		return nil, readError(resp)
	}

	var creds []*model.Credential
	if err := json.NewDecoder(resp.Body).Decode(&creds); err != nil {
		return nil, err
	}
	return creds, nil
}

// GetCredential возвращает запись по UUID.
func (c *Client) GetCredential(ctx context.Context, id uuid.UUID) (*model.Credential, error) {
	resp, err := c.doJSON(ctx, http.MethodGet, "/api/credentials/"+id.String(), nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	switch resp.StatusCode {
	case http.StatusOK:
	case http.StatusNotFound:
		return nil, ErrNotFound
	case http.StatusUnauthorized:
		return nil, ErrUnauthorized
	default:
		return nil, readError(resp)
	}

	var cred model.Credential
	if err := json.NewDecoder(resp.Body).Decode(&cred); err != nil {
		return nil, err
	}
	return &cred, nil
}

// UpdateCredential обновляет запись по UUID.
func (c *Client) UpdateCredential(ctx context.Context, id uuid.UUID, name, metadata string, data []byte) (*model.Credential, error) {
	body := map[string]any{
		"name":     name,
		"metadata": metadata,
		"data":     data,
	}
	resp, err := c.doJSON(ctx, http.MethodPut, "/api/credentials/"+id.String(), body)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	switch resp.StatusCode {
	case http.StatusOK:
	case http.StatusNotFound:
		return nil, ErrNotFound
	case http.StatusUnauthorized:
		return nil, ErrUnauthorized
	default:
		return nil, readError(resp)
	}

	var cred model.Credential
	if err := json.NewDecoder(resp.Body).Decode(&cred); err != nil {
		return nil, err
	}
	return &cred, nil
}

// DeleteCredential удаляет запись по UUID.
func (c *Client) DeleteCredential(ctx context.Context, id uuid.UUID) error {
	resp, err := c.doJSON(ctx, http.MethodDelete, "/api/credentials/"+id.String(), nil)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	switch resp.StatusCode {
	case http.StatusNoContent:
		return nil
	case http.StatusNotFound:
		return ErrNotFound
	case http.StatusUnauthorized:
		return ErrUnauthorized
	default:
		return readError(resp)
	}
}

// doJSON выполняет HTTP-запрос с JSON-телом и Bearer-токеном.
func (c *Client) doJSON(ctx context.Context, method, path string, body any) (*http.Response, error) {
	var bodyReader io.Reader
	if body != nil {
		data, err := json.Marshal(body)
		if err != nil {
			return nil, err
		}
		bodyReader = bytes.NewReader(data)
	}

	req, err := http.NewRequestWithContext(ctx, method, c.baseURL+path, bodyReader)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	if c.token != "" {
		req.Header.Set("Authorization", "Bearer "+c.token)
	}

	return c.httpClient.Do(req)
}

// readError читает тело ответа ошибки и формирует error.
func readError(resp *http.Response) error {
	data, _ := io.ReadAll(resp.Body)
	msg := string(data)
	if msg == "" {
		msg = resp.Status
	}
	return fmt.Errorf("сервер вернул %d: %s", resp.StatusCode, msg)
}
