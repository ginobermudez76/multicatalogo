package controllers

import (
	"strconv"

	"github.com/gofiber/fiber/v2"
	"multicatalogo-backend/models"
	"multicatalogo-backend/repository"
)

// ReporteController gestiona reportes y KPIs (Fase 6).
type ReporteController struct {
	repo repository.DBRepository
}

// NewReporteController crea una nueva instancia de ReporteController.
func NewReporteController(repo repository.DBRepository) *ReporteController {
	return &ReporteController{repo: repo}
}

// Resumen genera KPIs de ventas y comisiones (Fase 6).
func (rc *ReporteController) Resumen(c *fiber.Ctx) error {
	desde := c.Query("desde", "")
	hasta := c.Query("hasta", "")

	reporte, err := rc.repo.ObtenerReporteResumen(desde, hasta)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(models.APIError{
			Status:  fiber.StatusInternalServerError,
			Message: "Error al generar reporte",
			Details: err.Error(),
		})
	}

	return c.Status(fiber.StatusOK).JSON(reporte)
}

// TopProductos lista los productos más vendidos (Fase 6).
func (rc *ReporteController) TopProductos(c *fiber.Ctx) error {
	limit, _ := strconv.Atoi(c.Query("limit", "10"))
	if limit < 1 || limit > 100 {
		limit = 10
	}

	productos, err := rc.repo.ObtenerTopProductos(limit)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(models.APIError{
			Status:  fiber.StatusInternalServerError,
			Message: "Error al obtener top productos",
			Details: err.Error(),
		})
	}

	return c.Status(fiber.StatusOK).JSON(productos)
}

// TopReferidos lista los referidos con más ventas (Fase 6).
func (rc *ReporteController) TopReferidos(c *fiber.Ctx) error {
	limit, _ := strconv.Atoi(c.Query("limit", "10"))
	if limit < 1 || limit > 100 {
		limit = 10
	}

	referidos, err := rc.repo.ObtenerTopReferidos(limit)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(models.APIError{
			Status:  fiber.StatusInternalServerError,
			Message: "Error al obtener top referidos",
			Details: err.Error(),
		})
	}

	return c.Status(fiber.StatusOK).JSON(referidos)
}

// ComisionesPendientes lista comisiones pendientes por distribuidor (Fase 6).
func (rc *ReporteController) ComisionesPendientes(c *fiber.Ctx) error {
	comisiones, err := rc.repo.ObtenerComisionesPendientes()
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(models.APIError{
			Status:  fiber.StatusInternalServerError,
			Message: "Error al obtener comisiones pendientes",
			Details: err.Error(),
		})
	}

	return c.Status(fiber.StatusOK).JSON(comisiones)
}
