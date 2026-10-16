package routes_test

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/gofiber/fiber/v2"
	"multicatalogo-backend/config"
	"multicatalogo-backend/models"
	"multicatalogo-backend/routes"
)

// TestMain inicializa la conexión a PostgreSQL antes de ejecutar la suite de pruebas.
func TestMain(m *testing.M) {
	_, _ = config.ConnectDB()
	code := m.Run()
	_ = config.CloseDB()
	os.Exit(code)
}

func setupTestApp() *fiber.App {
	app := fiber.New()
	routes.SetupRoutes(app)
	return app
}

func TestHealthCheck(t *testing.T) {
	app := setupTestApp()

	req := httptest.NewRequest(http.MethodGet, "/api/health", nil)
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("Error en petición /api/health: %v", err)
	}

	if resp.StatusCode != http.StatusOK {
		t.Errorf("Esperaba status 200, obtuvo %d", resp.StatusCode)
	}

	body, _ := io.ReadAll(resp.Body)
	var res map[string]string
	if err := json.Unmarshal(body, &res); err != nil {
		t.Fatalf("Error decodificando respuesta JSON: %v", err)
	}

	if res["status"] != "ok" || res["message"] != "Servidor Go/Fiber operativo" {
		t.Errorf("Respuesta inesperada: %v", res)
	}
}

func TestLoginEmptyFields(t *testing.T) {
	app := setupTestApp()

	// Caso: email y password vacíos
	payload := []byte(`{"email":"","password":""}`)
	req := httptest.NewRequest(http.MethodPost, "/api/login", bytes.NewReader(payload))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("Error en petición /api/login: %v", err)
	}

	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("Esperaba status 400 para campos vacíos, obtuvo %d", resp.StatusCode)
	}

	body, _ := io.ReadAll(resp.Body)
	var apiErr models.APIError
	if err := json.Unmarshal(body, &apiErr); err != nil {
		t.Fatalf("Respuesta no cumple con APIError: %v", err)
	}

	if apiErr.Status != http.StatusBadRequest || apiErr.Message == "" {
		t.Errorf("APIError mal formado: %+v", apiErr)
	}
}

func TestLoginInvalidCredentials(t *testing.T) {
	app := setupTestApp()

	payload := []byte(`{"email":"wrong@upse.edu.ec","password":"wrong"}`)
	req := httptest.NewRequest(http.MethodPost, "/api/login", bytes.NewReader(payload))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("Error en petición /api/login: %v", err)
	}

	if resp.StatusCode != http.StatusUnauthorized {
		t.Errorf("Esperaba status 401 para credenciales inválidas, obtuvo %d", resp.StatusCode)
	}

	body, _ := io.ReadAll(resp.Body)
	var apiErr models.APIError
	if err := json.Unmarshal(body, &apiErr); err != nil {
		t.Fatalf("Respuesta no cumple con APIError: %v", err)
	}

	if apiErr.Status != http.StatusUnauthorized {
		t.Errorf("APIError Status incorrecto: %d", apiErr.Status)
	}
}

func TestLoginSuccess(t *testing.T) {
	app := setupTestApp()

	payload := []byte(`{"email":"admin@upse.edu.ec","password":"123456"}`)
	req := httptest.NewRequest(http.MethodPost, "/api/login", bytes.NewReader(payload))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("Error en petición /api/login: %v", err)
	}

	if resp.StatusCode != http.StatusOK {
		t.Errorf("Esperaba status 200, obtuvo %d", resp.StatusCode)
	}

	body, _ := io.ReadAll(resp.Body)
	var authRes models.LoginResponse
	if err := json.Unmarshal(body, &authRes); err != nil {
		t.Fatalf("Error decodificando respuesta LoginResponse: %v", err)
	}

	if authRes.Rol != "admin" || authRes.Email != "admin@upse.edu.ec" || authRes.Token == "" {
		t.Errorf("LoginResponse inesperada: %+v", authRes)
	}
}

func TestGetProductos(t *testing.T) {
	app := setupTestApp()

	req := httptest.NewRequest(http.MethodGet, "/api/productos", nil)
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("Error en petición /api/productos: %v", err)
	}

	if resp.StatusCode != http.StatusOK {
		t.Errorf("Esperaba status 200, obtuvo %d", resp.StatusCode)
	}

	body, _ := io.ReadAll(resp.Body)
	var prods []models.Producto
	if err := json.Unmarshal(body, &prods); err != nil {
		t.Fatalf("Error decodificando lista de productos: %v", err)
	}

	if len(prods) == 0 {
		t.Errorf("Se esperaba al menos 1 producto, obtuvo 0")
	}
}

func TestGetProductoPorIDValido(t *testing.T) {
	app := setupTestApp()

	req := httptest.NewRequest(http.MethodGet, "/api/productos/1", nil)
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("Error en petición /api/productos/1: %v", err)
	}

	if resp.StatusCode != http.StatusOK {
		t.Errorf("Esperaba status 200, obtuvo %d", resp.StatusCode)
	}

	body, _ := io.ReadAll(resp.Body)
	var prod models.Producto
	if err := json.Unmarshal(body, &prod); err != nil {
		t.Fatalf("Error decodificando producto: %v", err)
	}

	if prod.ID != 1 || prod.Nombre == "" {
		t.Errorf("Producto inesperado: %+v", prod)
	}
}

func TestGetProductoPorIDInvalido(t *testing.T) {
	app := setupTestApp()

	// Enviar texto en lugar de un número entero
	req := httptest.NewRequest(http.MethodGet, "/api/productos/texto-en-lugar-de-numero", nil)
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("Error en petición: %v", err)
	}

	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("Esperaba status 400 para ID no numérico, obtuvo %d", resp.StatusCode)
	}

	body, _ := io.ReadAll(resp.Body)
	var apiErr models.APIError
	if err := json.Unmarshal(body, &apiErr); err != nil {
		t.Fatalf("Respuesta no cumple formato APIError: %v", err)
	}

	if apiErr.Status != http.StatusBadRequest {
		t.Errorf("APIError Status incorrecto: %d", apiErr.Status)
	}
}

func TestGetProductoPorIDNoEncontrado(t *testing.T) {
	app := setupTestApp()

	req := httptest.NewRequest(http.MethodGet, "/api/productos/9999", nil)
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("Error en petición: %v", err)
	}

	if resp.StatusCode != http.StatusNotFound {
		t.Errorf("Esperaba status 404 para producto inexistente, obtuvo %d", resp.StatusCode)
	}

	body, _ := io.ReadAll(resp.Body)
	var apiErr models.APIError
	if err := json.Unmarshal(body, &apiErr); err != nil {
		t.Fatalf("Respuesta no cumple formato APIError: %v", err)
	}

	if apiErr.Status != http.StatusNotFound {
		t.Errorf("APIError Status incorrecto: %d", apiErr.Status)
	}
}

func TestGetRed(t *testing.T) {
	app := setupTestApp()

	req := httptest.NewRequest(http.MethodGet, "/api/red", nil)
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("Error en petición /api/red: %v", err)
	}

	if resp.StatusCode != http.StatusOK {
		t.Errorf("Esperaba status 200, obtuvo %d", resp.StatusCode)
	}

	body, _ := io.ReadAll(resp.Body)
	var red models.Referido
	if err := json.Unmarshal(body, &red); err != nil {
		t.Fatalf("Error decodificando estructura de red: %v", err)
	}

	if red.Nombre != "Tú" || len(red.Hijos) == 0 {
		t.Errorf("Nodo raíz de red incorrecto: %+v", red)
	}
}
