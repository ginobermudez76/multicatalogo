package routes_test

import (
	"bytes"
	"context"
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

// TestMain inicializa las conexiones a PostgreSQL y Redis antes de ejecutar la suite de pruebas.
func TestMain(m *testing.M) {
	_, _ = config.ConnectDB()
	_, _ = config.ConnectRedis()
	code := m.Run()
	_ = config.CloseRedis()
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

func TestGetProductosCacheAside(t *testing.T) {
	app := setupTestApp()
	rdb := config.GetRedis()
	if rdb != nil {
		_ = rdb.Del(context.Background(), "productos:all").Err()
	}

	// 1ra Petición: Cache Miss
	req1 := httptest.NewRequest(http.MethodGet, "/api/productos", nil)
	resp1, err := app.Test(req1)
	if err != nil {
		t.Fatalf("Error en 1ra petición: %v", err)
	}
	if rdb != nil && resp1.Header.Get("X-Cache") != "MISS" {
		t.Errorf("Esperaba X-Cache: MISS en primera llamada, obtuvo: %s", resp1.Header.Get("X-Cache"))
	}

	// 2da Petición: Cache Hit
	req2 := httptest.NewRequest(http.MethodGet, "/api/productos", nil)
	resp2, err := app.Test(req2)
	if err != nil {
		t.Fatalf("Error en 2da petición: %v", err)
	}
	if rdb != nil && resp2.Header.Get("X-Cache") != "HIT" {
		t.Errorf("Esperaba X-Cache: HIT en segunda llamada, obtuvo: %s", resp2.Header.Get("X-Cache"))
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

func TestGetRedCacheAside(t *testing.T) {
	app := setupTestApp()
	rdb := config.GetRedis()
	if rdb != nil {
		_ = rdb.Del(context.Background(), "red:tree:admin@upse.edu.ec").Err()
	}

	// 1ra Petición: Cache Miss (construcción costosa desde PostgreSQL)
	req1 := httptest.NewRequest(http.MethodGet, "/api/red?email=admin@upse.edu.ec", nil)
	resp1, err := app.Test(req1)
	if err != nil {
		t.Fatalf("Error en 1ra petición /api/red: %v", err)
	}
	if resp1.StatusCode != http.StatusOK {
		t.Errorf("Esperaba status 200, obtuvo %d", resp1.StatusCode)
	}
	if rdb != nil && resp1.Header.Get("X-Cache") != "MISS" {
		t.Errorf("Esperaba X-Cache: MISS en primera llamada, obtuvo: %s", resp1.Header.Get("X-Cache"))
	}

	// 2da Petición: Cache Hit (retorno inmediato desde Redis)
	req2 := httptest.NewRequest(http.MethodGet, "/api/red?email=admin@upse.edu.ec", nil)
	resp2, err := app.Test(req2)
	if err != nil {
		t.Fatalf("Error en 2da petición /api/red: %v", err)
	}
	if resp2.StatusCode != http.StatusOK {
		t.Errorf("Esperaba status 200, obtuvo %d", resp2.StatusCode)
	}
	if rdb != nil && resp2.Header.Get("X-Cache") != "HIT" {
		t.Errorf("Esperaba X-Cache: HIT en segunda llamada, obtuvo: %s", resp2.Header.Get("X-Cache"))
	}
}

func TestCheckoutSuccess(t *testing.T) {
	app := setupTestApp()

	payload := models.CheckoutRequest{
		Cliente: models.ClienteCheckout{
			Nombre:    "María López",
			Email:     "maria.lopez@example.com",
			Ciudad:    "Guayaquil",
			Direccion: "Av. 9 de Octubre 123",
			Telefono:  "0991234567",
		},
		MetodoPago: "Tarjeta de crédito",
		Items: []models.ItemCarrito{
			{ID: 1, Nombre: "Serum Revitalizante", Cantidad: 2, Precio: 45.00},
			{ID: 2, Nombre: "Crema Hidratante Pro", Cantidad: 1, Precio: 32.50},
		},
		Total: 122.50,
	}

	bodyBytes, _ := json.Marshal(payload)
	req := httptest.NewRequest(http.MethodPost, "/api/checkout", bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("Error en petición /api/checkout: %v", err)
	}

	if resp.StatusCode != http.StatusCreated {
		respBody, _ := io.ReadAll(resp.Body)
		t.Fatalf("Esperaba status 201 Created, obtuvo %d. Respuesta: %s", resp.StatusCode, string(respBody))
	}

	body, _ := io.ReadAll(resp.Body)
	var confirmado models.PedidoConfirmado
	if err := json.Unmarshal(body, &confirmado); err != nil {
		t.Fatalf("Error decodificando PedidoConfirmado: %v", err)
	}

	if confirmado.Numero == "" || len(confirmado.Numero) < 4 || confirmado.Numero[:3] != "MC-" {
		t.Errorf("Número de pedido inválido: %s", confirmado.Numero)
	}

	if confirmado.Total != 122.50 {
		t.Errorf("Total del pedido incorrecto: %f", confirmado.Total)
	}

	if len(confirmado.Items) != 2 {
		t.Errorf("Cantidad de ítems incorrecta: %d", len(confirmado.Items))
	}
}

func TestCheckoutEmptyItems(t *testing.T) {
	app := setupTestApp()

	payload := models.CheckoutRequest{
		Cliente: models.ClienteCheckout{
			Nombre: "Juan Pérez",
			Email:  "juan@example.com",
			Ciudad: "Quito",
		},
		MetodoPago: "Efectivo contra entrega",
		Items:      []models.ItemCarrito{}, // Carrito vacío
		Total:      100.0,
	}

	bodyBytes, _ := json.Marshal(payload)
	req := httptest.NewRequest(http.MethodPost, "/api/checkout", bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("Error en petición /api/checkout: %v", err)
	}

	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("Esperaba status 400 para carrito vacío, obtuvo %d", resp.StatusCode)
	}
}

func TestCheckoutMissingClientInfo(t *testing.T) {
	app := setupTestApp()

	payload := models.CheckoutRequest{
		Cliente: models.ClienteCheckout{
			Nombre: "", // Nombre vacío
			Email:  "",
		},
		MetodoPago: "Transferencia bancaria",
		Items: []models.ItemCarrito{
			{ID: 1, Nombre: "Serum Revitalizante", Cantidad: 1, Precio: 45.00},
		},
		Total: 45.0,
	}

	bodyBytes, _ := json.Marshal(payload)
	req := httptest.NewRequest(http.MethodPost, "/api/checkout", bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("Error en petición /api/checkout: %v", err)
	}

	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("Esperaba status 400 para cliente con campos requeridos vacíos, obtuvo %d", resp.StatusCode)
	}
}

