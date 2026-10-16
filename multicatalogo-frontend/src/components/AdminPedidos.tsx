// src/components/AdminPedidos.tsx
import { useState, useEffect } from "react";
import { getPedidosAdminApi, patchEstadoPedidoApi } from "../services/api";

interface ItemPedido {
  nombre: string;
  cantidad: number;
  precio: number;
}

interface PedidoAdmin {
  id: number;
  numero: string;
  cliente_nombre: string;
  cliente_email: string;
  cliente_ciudad: string;
  metodo_pago: string;
  total: number;
  estado: string;
  fecha: string;
  items?: ItemPedido[];
}

const ESTADOS = ["", "pendiente", "preparando", "enviado", "entregado", "cancelado"];

const AdminPedidos = () => {
  const [pedidos, setPedidos] = useState<PedidoAdmin[]>([]);
  const [loading, setLoading] = useState<boolean>(true);
  const [filtroEstado, setFiltroEstado] = useState<string>("");
  const [error, setError] = useState<string>("");

  useEffect(() => {
    cargarPedidos();
  }, [filtroEstado]);

  const cargarPedidos = () => {
    setLoading(true);
    getPedidosAdminApi(1, 50, filtroEstado)
      .then((res) => {
        setPedidos(res.data || []);
      })
      .catch((err: unknown) => {
        setError(err instanceof Error ? err.message : "Error al cargar pedidos");
      })
      .finally(() => {
        setLoading(false);
      });
  };

  const handleCambiarEstado = async (id: number, nuevoEstado: string) => {
    try {
      await patchEstadoPedidoApi(id, nuevoEstado);
      cargarPedidos();
    } catch (err: unknown) {
      alert(err instanceof Error ? err.message : "Error al actualizar estado");
    }
  };

  return (
    <div className="space-y-6">
      <div className="flex flex-col sm:flex-row justify-between items-start sm:items-center gap-4">
        <div>
          <h1 className="text-2xl font-bold text-slate-800">Gestión de Pedidos</h1>
          <p className="text-slate-500 text-sm">Flujo de estados de compras, reversión de stock y comisiones (Fase 3 U2)</p>
        </div>

        {/* Filtro por estado */}
        <select
          value={filtroEstado}
          onChange={(e) => setFiltroEstado(e.target.value)}
          className="text-xs px-3 py-2 border border-slate-300 rounded-lg bg-white outline-none focus:border-indigo-600 font-semibold text-slate-700"
        >
          <option value="">Todos los estados</option>
          <option value="pendiente">Pendiente</option>
          <option value="preparando">Preparando</option>
          <option value="enviado">Enviado</option>
          <option value="entregado">Entregado</option>
          <option value="cancelado">Cancelado</option>
        </select>
      </div>

      {error && <div className="p-3 bg-red-50 text-red-600 rounded-lg text-xs">{error}</div>}

      <div className="bg-white rounded-xl border border-slate-200 shadow-sm overflow-hidden">
        {loading ? (
          <div className="p-8 text-center text-xs text-slate-400">Cargando órdenes...</div>
        ) : (
          <div className="overflow-x-auto">
            <table className="w-full text-left text-xs border-collapse">
              <thead>
                <tr className="border-b border-slate-200 bg-slate-50 text-slate-600 uppercase font-semibold">
                  <th className="py-3 px-4">Orden</th>
                  <th className="py-3 px-4">Cliente</th>
                  <th className="py-3 px-4">Ciudad</th>
                  <th className="py-3 px-4">Total</th>
                  <th className="py-3 px-4">Fecha</th>
                  <th className="py-3 px-4">Estado</th>
                  <th className="py-3 px-4 text-right">Cambiar Estado</th>
                </tr>
              </thead>
              <tbody className="divide-y divide-slate-100 text-slate-700">
                {pedidos.map((p) => (
                  <tr key={p.id} className="hover:bg-slate-50/60">
                    <td className="py-3 px-4 font-mono font-bold text-slate-800">{p.numero}</td>
                    <td className="py-3 px-4">
                      <p className="font-semibold text-slate-800">{p.cliente_nombre}</p>
                      <p className="text-[11px] text-slate-400">{p.cliente_email}</p>
                    </td>
                    <td className="py-3 px-4">{p.cliente_ciudad || "—"}</td>
                    <td className="py-3 px-4 font-bold text-slate-900">${p.total.toFixed(2)}</td>
                    <td className="py-3 px-4 text-slate-400">
                      {new Date(p.fecha).toLocaleDateString()}
                    </td>
                    <td className="py-3 px-4">
                      <span
                        className={`px-2.5 py-0.5 rounded-full text-[11px] font-semibold border ${
                          p.estado === "entregado"
                            ? "bg-emerald-100 text-emerald-700 border-emerald-200"
                            : p.estado === "cancelado"
                            ? "bg-red-100 text-red-700 border-red-200"
                            : p.estado === "enviado"
                            ? "bg-blue-100 text-blue-700 border-blue-200"
                            : "bg-amber-100 text-amber-700 border-amber-200"
                        }`}
                      >
                        {p.estado}
                      </span>
                    </td>
                    <td className="py-3 px-4 text-right">
                      <select
                        value={p.estado}
                        disabled={p.estado === "cancelado" || p.estado === "entregado"}
                        onChange={(e) => handleCambiarEstado(p.id, e.target.value)}
                        className="text-xs px-2 py-1 border border-slate-300 rounded bg-white font-medium text-slate-700 outline-none cursor-pointer disabled:opacity-40 disabled:cursor-not-allowed"
                      >
                        {ESTADOS.filter(Boolean).map((est) => (
                          <option key={est} value={est}>
                            {est}
                          </option>
                        ))}
                      </select>
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        )}
      </div>
    </div>
  );
};

export default AdminPedidos;
