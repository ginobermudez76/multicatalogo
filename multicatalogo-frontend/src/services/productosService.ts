// src/services/productosService.ts
// Capa de servicios del catálogo conectado a la API Go / Fiber (Fase 2 y 3).
// Obtiene los datos vivos desde PostgreSQL / Redis con fallback automático a datos locales.

import { productosMock, type Producto } from "../data/productos";
import { getProductosApi } from "./api";

const API_BASE_URL = import.meta.env.VITE_API_URL || "http://localhost:3000";

// Normaliza el producto recibido desde la API para garantizar compatibilidad con la interfaz frontend
const normalizarProducto = (p: any): Producto => ({
  id: p.id,
  nombre: p.nombre,
  descripcion: p.descripcion || "",
  precio: typeof p.precio === "number" ? p.precio : parseFloat(p.precio) || 0,
  categoria: p.categoria || "General",
  img: p.img || "https://picsum.photos/seed/default/600",
  galeria: Array.isArray(p.galeria) && p.galeria.length > 0 ? p.galeria : [p.img || "https://picsum.photos/seed/default/600"],
});

// GET /api/productos (lista completa con Cache-Aside en Redis y persistencia en PostgreSQL)
export const getProductos = async (): Promise<Producto[]> => {
  try {
    const data = await getProductosApi();
    if (Array.isArray(data) && data.length > 0) {
      return data.map(normalizarProducto);
    }
    return productosMock;
  } catch (error) {
    console.warn("Fallo al consultar GET /api/productos, usando datos locales:", error);
    return productosMock;
  }
};

// GET /api/productos/:id (detalle de un producto individual)
export const getProductoById = async (id: number): Promise<Producto | undefined> => {
  try {
    const response = await fetch(`${API_BASE_URL}/api/productos/${id}`, {
      headers: { Accept: "application/json" },
    });
    if (response.ok) {
      const item = await response.json();
      return normalizarProducto(item);
    }
    return productosMock.find((p) => p.id === id);
  } catch (error) {
    console.warn(`Fallo al consultar GET /api/productos/${id}, usando datos locales:`, error);
    return productosMock.find((p) => p.id === id);
  }
};


