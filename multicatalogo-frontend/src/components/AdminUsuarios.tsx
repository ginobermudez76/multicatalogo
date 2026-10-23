// src/components/AdminUsuarios.tsx
// Panel de administración de cuentas de usuarios (Fase 1 y 2 Back-office U2):
// - Listado de cuentas con búsqueda en tiempo real (?q=)
// - Cambio de rol (admin <-> cliente) con confirmación
// - Activación / desactivación de cuentas (bloqueo en /api/login)
import { useState, useEffect } from "react";
import {
  getUsuariosAdminApi,
  patchUsuarioAdminApi,
  type UsuarioAdminItem,
} from "../services/api";

const AdminUsuarios = () => {
  const [usuarios, setUsuarios] = useState<UsuarioAdminItem[]>([]);
  const [loading, setLoading] = useState<boolean>(true);
  const [busqueda, setBusqueda] = useState<string>("");
  const [error, setError] = useState<string>("");
  const [mensajeExito, setMensajeExito] = useState<string>("");
  const [pagina, setPagina] = useState<number>(1);
  const [total, setTotal] = useState<number>(0);
  const limite = 15;

  useEffect(() => {
    cargarUsuarios();
  }, [busqueda, pagina]);

  const cargarUsuarios = () => {
    setLoading(true);
    getUsuariosAdminApi(pagina, limite, busqueda)
      .then((res) => {
        setUsuarios(res.data || []);
        setTotal(res.total || 0);
      })
      .catch((err: unknown) => {
        setError(err instanceof Error ? err.message : "Error al cargar usuarios");
      })
      .finally(() => {
        setLoading(false);
      });
  };

  const handleCambiarRol = async (id: number, nuevoRol: string) => {
    try {
      setError("");
      await patchUsuarioAdminApi(id, { rol: nuevoRol });
      setMensajeExito(`Rol actualizado a "${nuevoRol}" con éxito.`);
      setTimeout(() => setMensajeExito(""), 3000);
      cargarUsuarios();
    } catch (err: unknown) {
      setError(err instanceof Error ? err.message : "Error al cambiar el rol");
    }
  };

  const handleToggleActivo = async (id: number, estadoActual: boolean) => {
    const accion = estadoActual ? "desactivar" : "activar";
    if (!window.confirm(`¿Estás seguro de que deseas ${accion} esta cuenta de usuario?`)) {
      return;
    }

    try {
      setError("");
      await patchUsuarioAdminApi(id, { activo: !estadoActual });
      setMensajeExito(`Cuenta ${estadoActual ? "desactivada" : "activada"} correctamente.`);
      setTimeout(() => setMensajeExito(""), 3000);
      cargarUsuarios();
    } catch (err: unknown) {
      setError(err instanceof Error ? err.message : `Error al ${accion} la cuenta`);
    }
  };

  const totalActivos = usuarios.filter((u) => u.activo).length;
  const totalInactivos = usuarios.filter((u) => !u.activo).length;

  return (
    <div className="space-y-6">
      {/* Cabecera */}
      <div className="flex flex-col sm:flex-row sm:items-center sm:justify-between gap-4">
        <div>
          <h1 className="text-2xl font-bold text-slate-800">Gestión de Usuarios</h1>
          <p className="text-slate-500 text-sm mt-0.5">
            Administra roles, estados de activación y cuentas de la plataforma
          </p>
        </div>

        {/* Resumen rápido */}
        <div className="flex items-center gap-3">
          <span className="inline-flex items-center px-3 py-1 rounded-full text-xs font-semibold bg-emerald-50 text-emerald-700 border border-emerald-200">
            {totalActivos} Activos
          </span>
          <span className="inline-flex items-center px-3 py-1 rounded-full text-xs font-semibold bg-slate-100 text-slate-600 border border-slate-200">
            {totalInactivos} Inactivos
          </span>
          <span className="inline-flex items-center px-3 py-1 rounded-full text-xs font-semibold bg-indigo-50 text-indigo-700 border border-indigo-200">
            {total} Registrados
          </span>
        </div>
      </div>

      {/* Alertas */}
      {mensajeExito && (
        <div className="bg-emerald-50 border border-emerald-200 text-emerald-800 px-4 py-3 rounded-lg text-sm flex items-center justify-between shadow-sm animate-fade-in">
          <div className="flex items-center gap-2">
            <span>✓</span>
            <span>{mensajeExito}</span>
          </div>
          <button type="button" onClick={() => setMensajeExito("")} className="font-bold">×</button>
        </div>
      )}

      {error && (
        <div className="bg-red-50 border border-red-200 text-red-700 px-4 py-3 rounded-lg text-sm flex items-center justify-between shadow-sm">
          <span>{error}</span>
          <button type="button" onClick={() => setError("")} className="font-bold">×</button>
        </div>
      )}

      {/* Barra de Filtro y Búsqueda */}
      <div className="bg-white p-4 rounded-xl border border-slate-200 shadow-sm flex flex-col sm:flex-row items-center justify-between gap-4">
        <div className="relative w-full sm:w-96">
          <span className="absolute inset-y-0 left-0 pl-3 flex items-center pointer-events-none text-slate-400">
            🔍
          </span>
          <input
            type="text"
            value={busqueda}
            onChange={(e) => {
              setBusqueda(e.target.value);
              setPagina(1);
            }}
            placeholder="Buscar por nombre, correo o rol..."
            className="w-full pl-9 pr-4 py-2 border border-slate-300 rounded-lg text-sm focus:ring-2 focus:ring-indigo-600 outline-none"
          />
        </div>
        <p className="text-xs text-slate-500 self-end sm:self-center">
          Mostrando página {pagina} de {Math.max(1, Math.ceil(total / limite))}
        </p>
      </div>

      {/* Tabla de Usuarios */}
      <div className="bg-white rounded-xl border border-slate-200 shadow-sm overflow-hidden">
        {loading ? (
          <div className="p-12 text-center text-slate-500">
            <div className="inline-block animate-spin rounded-full h-8 w-8 border-4 border-indigo-600 border-t-transparent mb-3" />
            <p>Cargando usuarios registrados...</p>
          </div>
        ) : usuarios.length === 0 ? (
          <div className="p-12 text-center text-slate-500">
            <p className="text-lg font-medium text-slate-700">No se encontraron usuarios</p>
            <p className="text-sm mt-1">Intenta cambiar el término de búsqueda</p>
          </div>
        ) : (
          <div className="overflow-x-auto">
            <table className="w-full text-left text-sm text-slate-600">
              <thead className="bg-slate-50 text-slate-700 uppercase font-semibold text-xs border-b border-slate-200">
                <tr>
                  <th className="px-6 py-4">Usuario</th>
                  <th className="px-6 py-4">Rol</th>
                  <th className="px-6 py-4">Patrocinador</th>
                  <th className="px-6 py-4">Estado</th>
                  <th className="px-6 py-4">Fecha Registro</th>
                  <th className="px-6 py-4 text-right">Acciones</th>
                </tr>
              </thead>
              <tbody className="divide-y divide-slate-100">
                {usuarios.map((u) => {
                  const inicial = (u.nombre || u.email).charAt(0).toUpperCase();
                  const fechaReg = u.creado_en ? new Date(u.creado_en).toLocaleDateString("es-EC") : "—";

                  return (
                    <tr key={u.id} className="hover:bg-slate-50/80 transition">
                      {/* Usuario (Nombre + Email + Avatar) */}
                      <td className="px-6 py-4">
                        <div className="flex items-center gap-3">
                          <div className={`w-9 h-9 rounded-full flex items-center justify-center font-bold text-white text-xs ${
                            u.rol === "admin" ? "bg-purple-600" : "bg-indigo-600"
                          }`}>
                            {inicial}
                          </div>
                          <div>
                            <p className="font-semibold text-slate-800">
                              {u.nombre || "Sin nombre registrado"}
                            </p>
                            <p className="text-xs text-slate-500 font-mono">{u.email}</p>
                          </div>
                        </div>
                      </td>

                      {/* Rol */}
                      <td className="px-6 py-4">
                        <select
                          value={u.rol}
                          onChange={(e) => handleCambiarRol(u.id, e.target.value)}
                          className={`text-xs font-semibold px-2.5 py-1 rounded-full border outline-none cursor-pointer ${
                            u.rol === "admin"
                              ? "bg-purple-50 text-purple-700 border-purple-200"
                              : "bg-blue-50 text-blue-700 border-blue-200"
                          }`}
                        >
                          <option value="cliente">cliente</option>
                          <option value="admin">admin</option>
                        </select>
                      </td>

                      {/* Patrocinador / Referido ID */}
                      <td className="px-6 py-4">
                        {u.referido_id ? (
                          <span className="inline-flex items-center gap-1.5 text-xs font-mono bg-slate-100 px-2 py-0.5 rounded border border-slate-200 text-slate-700">
                            ID #{u.referido_id}
                          </span>
                        ) : (
                          <span className="text-xs text-slate-400">Directo (Raíz)</span>
                        )}
                      </td>

                      {/* Estado Activo */}
                      <td className="px-6 py-4">
                        <button
                          type="button"
                          onClick={() => handleToggleActivo(u.id, u.activo)}
                          title="Haz clic para activar o desactivar la cuenta"
                          className={`inline-flex items-center gap-1.5 px-3 py-1 rounded-full text-xs font-semibold transition ${
                            u.activo
                              ? "bg-emerald-100 text-emerald-800 hover:bg-emerald-200 border border-emerald-300"
                              : "bg-red-100 text-red-800 hover:bg-red-200 border border-red-300"
                          }`}
                        >
                          <span className={`w-2 h-2 rounded-full ${u.activo ? "bg-emerald-600" : "bg-red-600"}`} />
                          {u.activo ? "Activo" : "Desactivado"}
                        </button>
                      </td>

                      {/* Fecha de Registro */}
                      <td className="px-6 py-4 text-xs text-slate-500">{fechaReg}</td>

                      {/* Acciones */}
                      <td className="px-6 py-4 text-right">
                        <button
                          type="button"
                          onClick={() => handleToggleActivo(u.id, u.activo)}
                          className="text-xs font-medium text-slate-600 hover:text-indigo-600 hover:underline"
                        >
                          {u.activo ? "Suspender cuenta" : "Reactivar cuenta"}
                        </button>
                      </td>
                    </tr>
                  );
                })}
              </tbody>
            </table>
          </div>
        )}

        {/* Paginación */}
        {total > limite && (
          <div className="p-4 border-t border-slate-200 flex items-center justify-between text-sm text-slate-600">
            <button
              type="button"
              disabled={pagina <= 1}
              onClick={() => setPagina((p) => Math.max(1, p - 1))}
              className="px-3 py-1.5 border border-slate-300 rounded-lg disabled:opacity-40 hover:bg-slate-50 transition"
            >
              ← Anterior
            </button>
            <span className="text-xs">
              Página {pagina} de {Math.ceil(total / limite)}
            </span>
            <button
              type="button"
              disabled={pagina >= Math.ceil(total / limite)}
              onClick={() => setPagina((p) => p + 1)}
              className="px-3 py-1.5 border border-slate-300 rounded-lg disabled:opacity-40 hover:bg-slate-50 transition"
            >
              Siguiente →
            </button>
          </div>
        )}
      </div>
    </div>
  );
};

export default AdminUsuarios;
