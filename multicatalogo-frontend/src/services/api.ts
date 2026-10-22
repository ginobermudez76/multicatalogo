// Configuración centralizada del cliente API para conectar con el backend Go (Fiber)
const API_BASE_URL = import.meta.env.VITE_API_URL || 'http://localhost:3000';

export interface LoginCredentials {
  email: string;
  password: string;
}

export interface RegisterPayload {
  nombre: string;
  email: string;
  password: string;
  ref?: string;
}

export interface LoginResponse {
  token: string;
  email: string;
  rol?: string;
}

export interface ApiErrorResponse {
  error?: string;
  message?: string;
  status?: number;
  details?: unknown;
}

export interface Producto {
  id: number;
  nombre: string;
  descripcion?: string;
  precio: number;
  categoria?: string;
  img: string;
  galeria?: string[];
  activo?: boolean;
  stock?: number;
}

const getAuthHeaders = (): Record<string, string> => {
  const token = localStorage.getItem('multicatalogo_token');
  const headers: Record<string, string> = {
    'Content-Type': 'application/json',
  };
  if (token) {
    headers['Authorization'] = `Bearer ${token}`;
  }
  return headers;
};

/**
 * Realiza la petición de autenticación hacia el endpoint POST /api/login del backend Go.
 */
export const loginApi = async (credentials: LoginCredentials): Promise<LoginResponse> => {
  try {
    const response = await fetch(`${API_BASE_URL}/api/login`, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
      },
      body: JSON.stringify(credentials),
    });

    const data = await response.json().catch(() => null);

    if (!response.ok) {
      const errorMsg = data?.message || data?.error || `Error en el servidor (${response.status})`;
      throw new Error(errorMsg);
    }

    if (data?.token) {
      localStorage.setItem('multicatalogo_token', data.token);
    }

    return data as LoginResponse;
  } catch (error) {
    if (error instanceof TypeError && error.message.includes('fetch')) {
      throw new Error('No se pudo conectar con el servidor Go backend (puerto 3000). Verifica que esté encendido.', {
        cause: error,
      });
    }
    throw error;
  }
};

/**
 * Registra un nuevo usuario con código de referido opcional (Fase 1 U2).
 */
export const registerApi = async (payload: RegisterPayload): Promise<{ id: number; email: string; nombre: string; rol: string }> => {
  try {
    const response = await fetch(`${API_BASE_URL}/api/register`, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
      },
      body: JSON.stringify(payload),
    });

    const data = await response.json().catch(() => null);

    if (!response.ok) {
      const errorMsg = data?.message || data?.error || `Error en el registro (${response.status})`;
      throw new Error(errorMsg);
    }

    return data;
  } catch (error) {
    if (error instanceof TypeError && error.message.includes('fetch')) {
      throw new Error('No se pudo conectar con el servidor Go backend.', { cause: error });
    }
    throw error;
  }
};

/**
 * Obtiene el catálogo de productos disponibles desde el endpoint GET /api/productos del backend Go.
 */
export const getProductosApi = async (): Promise<Producto[]> => {
  try {
    const response = await fetch(`${API_BASE_URL}/api/productos`, {
      method: 'GET',
      headers: {
        'Accept': 'application/json',
      },
    });

    if (!response.ok) {
      throw new Error(`Error al obtener productos (${response.status})`);
    }

    return await response.json();
  } catch (error) {
    if (error instanceof TypeError && error.message.includes('fetch')) {
      throw new Error('No se pudo conectar con el servidor Go backend. Verifica que esté en ejecución en http://localhost:3000.', {
        cause: error,
      });
    }
    throw error;
  }
};

/**
 * Obtiene los pedidos del usuario autenticado (Fase 3 U2).
 */
export const getMisPedidosApi = async (page = 1, limit = 10) => {
  const response = await fetch(`${API_BASE_URL}/api/pedidos/mios?page=${page}&limit=${limit}`, {
    headers: getAuthHeaders(),
  });
  if (!response.ok) {
    const data = await response.json().catch(() => null);
    throw new Error(data?.message || 'Error al obtener tus pedidos');
  }
  return response.json();
};

/**
 * Obtiene todos los pedidos para el rol admin (Fase 3 U2).
 */
export const getPedidosAdminApi = async (page = 1, limit = 10, estado = '') => {
  const url = `${API_BASE_URL}/api/pedidos?page=${page}&limit=${limit}${estado ? `&estado=${estado}` : ''}`;
  const response = await fetch(url, {
    headers: getAuthHeaders(),
  });
  if (!response.ok) {
    const data = await response.json().catch(() => null);
    throw new Error(data?.message || 'Error al obtener pedidos');
  }
  return response.json();
};

/**
 * Cambia el estado de un pedido (admin).
 */
export const patchEstadoPedidoApi = async (id: number, nuevoEstado: string) => {
  const response = await fetch(`${API_BASE_URL}/api/pedidos/${id}/estado`, {
    method: 'PATCH',
    headers: getAuthHeaders(),
    body: JSON.stringify({ estado: nuevoEstado }),
  });
  if (!response.ok) {
    const data = await response.json().catch(() => null);
    throw new Error(data?.message || 'Error al actualizar estado del pedido');
  }
  return response.json();
};

/**
 * Obtiene el subárbol de la red del usuario autenticado (Fase 4 U2).
 */
export const getMiRedPropiaApi = async () => {
  const response = await fetch(`${API_BASE_URL}/api/red/mia`, {
    headers: getAuthHeaders(),
  });
  if (!response.ok) {
    const data = await response.json().catch(() => null);
    throw new Error(data?.message || 'Error al obtener tu red');
  }
  return response.json();
};

/**
 * Obtiene el ledger de comisiones del usuario (Fase 4 U2).
 */
export const getMisComisionesApi = async () => {
  const response = await fetch(`${API_BASE_URL}/api/comisiones/mias`, {
    headers: getAuthHeaders(),
  });
  if (!response.ok) {
    const data = await response.json().catch(() => null);
    throw new Error(data?.message || 'Error al obtener comisiones');
  }
  return response.json();
};

/**
 * CRUD de productos para el administrador (Fase 2 U2).
 */
export const getProductosAdminApi = async (page = 1, limit = 20, q = '') => {
  const url = `${API_BASE_URL}/api/admin/productos?page=${page}&limit=${limit}${q ? `&q=${encodeURIComponent(q)}` : ''}`;
  const response = await fetch(url, { headers: getAuthHeaders() });
  if (!response.ok) throw new Error('Error al listar productos admin');
  return response.json();
};

export const crearProductoAdminApi = async (prod: Partial<Producto>) => {
  const response = await fetch(`${API_BASE_URL}/api/admin/productos`, {
    method: 'POST',
    headers: getAuthHeaders(),
    body: JSON.stringify(prod),
  });
  if (!response.ok) {
    const data = await response.json().catch(() => null);
    throw new Error(data?.message || 'Error al crear producto');
  }
  return response.json();
};

export const actualizarProductoAdminApi = async (id: number, prod: Partial<Producto>) => {
  const response = await fetch(`${API_BASE_URL}/api/admin/productos/${id}`, {
    method: 'PUT',
    headers: getAuthHeaders(),
    body: JSON.stringify(prod),
  });
  if (!response.ok) {
    const data = await response.json().catch(() => null);
    throw new Error(data?.message || 'Error al actualizar producto');
  }
  return response.json();
};

export const eliminarProductoAdminApi = async (id: number) => {
  const response = await fetch(`${API_BASE_URL}/api/admin/productos/${id}`, {
    method: 'DELETE',
    headers: getAuthHeaders(),
  });
  if (!response.ok) {
    const data = await response.json().catch(() => null);
    throw new Error(data?.message || 'Error al eliminar producto');
  }
  return response.json();
};

/**
 * Reportes de Backoffice (Fase 6 U2).
 */
export const getReporteResumenApi = async (desde = '', hasta = '') => {
  const params = new URLSearchParams();
  if (desde) params.append('desde', desde);
  if (hasta) params.append('hasta', hasta);
  const response = await fetch(`${API_BASE_URL}/api/reportes/resumen?${params.toString()}`, {
    headers: getAuthHeaders(),
  });
  if (!response.ok) throw new Error('Error al obtener reporte resumen');
  return response.json();
};

export interface CheckoutPayload {
  cliente: {
    nombre: string;
    email: string;
    ciudad?: string;
    direccion?: string;
    telefono?: string;
  };
  metodoPago: string;
  items: {
    id?: number;
    nombre: string;
    cantidad: number;
    precio: number;
  }[];
  total: number;
}

export interface PedidoConfirmadoResponse {
  numero: string;
  items: { nombre: string; cantidad: number; precio: number }[];
  total: number;
  cliente: { nombre: string; email: string; ciudad: string };
  fecha: string;
}

/**
 * Procesa la orden de compra en el backend Go con transacciones ACID y Goroutines (Fase 4 U2).
 */
export const checkoutApi = async (payload: CheckoutPayload): Promise<PedidoConfirmadoResponse> => {
  const response = await fetch(`${API_BASE_URL}/api/checkout`, {
    method: 'POST',
    headers: getAuthHeaders(),
    body: JSON.stringify(payload),
  });

  const data = await response.json().catch(() => null);

  if (!response.ok) {
    const errorMsg = data?.message || data?.error || `Error en el procesamiento del pedido (${response.status})`;
    throw new Error(errorMsg);
  }

  return data as PedidoConfirmadoResponse;
};

