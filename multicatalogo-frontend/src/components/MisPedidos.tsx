// src/components/MisPedidos.tsx
import { useState, useEffect } from "react";
import { getMisPedidosApi } from "../services/api";

interface ItemPedido {
  id?: number;
  nombre: string;
  cantidad: number;
  precio: number;
  subtotal?: number;
}

interface Pedido {
  id: number;
  numero: string;
  cliente_nombre: string;
  cliente_email: string;
  cliente_ciudad: string;
  cliente_direccion: string;
  cliente_telefono: string;
  metodo_pago: string;
  total: number;
  estado: string;
  fecha: string;
  items?: ItemPedido[];
}

const MisPedidos = () => {
  const [pedidos, setPedidos] = useState<Pedido[]>([]);
  const [loading, setLoading] = useState<boolean>(true);
  const [error, setError] = useState<string>("");
  const [expandido, setExpandido] = useState<number | null>(null);

  useEffect(() => {
    cargarPedidos();
  }, []);

  const cargarPedidos = () => {
    setLoading(true);
    getMisPedidosApi(1, 20)
      .then((res) => {
        setPedidos(res.data || []);
      })
      .catch((err: unknown) => {
        setError(err instanceof Error ? err.message : "Error al cargar tus pedidos");
      })
      .finally(() => {
        setLoading(false);
      });
  };

  const getEstadoBadge = (estado: string) => {
    switch (estado) {
      case "entregado":
        return "bg-emerald-100 text-emerald-700 border-emerald-200";
      case "enviado":
        return "bg-blue-100 text-blue-700 border-blue-200";
      case "preparando":
        return "bg-purple-100 text-purple-700 border-purple-200";
      case "cancelado":
        return "bg-red-100 text-red-700 border-red-200";
      default:
        return "bg-amber-100 text-amber-700 border-amber-200";
    }
  };

  return (
    <div className="space-y-6">
      <div>
        <h1 className="text-2xl font-bold text-slate-800 mb-2">Mis Pedidos</h1>
        <p className="text-slate-500">Historial de compras realizadas en la plataforma</p>
      </div>

      {error && (
        <div className="bg-red-50 text-red-600 p-4 rounded-lg border border-red-200 text-sm">
          {error}
        </div>
      )}

      {loading ? (
        <div className="p-8 text-center text-slate-500 bg-white rounded-lg border border-slate-200">
          Cargando pedidos...
        </div>
      ) : pedidos.length === 0 ? (
        <div className="p-12 text-center bg-white rounded-lg border border-dashed border-slate-300">
          <p className="text-slate-600 font-semibold mb-1">Aún no has realizado ninguna compra</p>
          <p className="text-xs text-slate-400">Tus órdenes aparecerán aquí con su estado en tiempo real.</p>
        </div>
      ) : (
        <div className="space-y-4">
          {pedidos.map((p) => {
            const isExpanded = expandido === p.id;
            return (
              <div key={p.id} className="bg-white rounded-xl border border-slate-200 shadow-sm overflow-hidden">
                <div
                  onClick={() => setExpandido(isExpanded ? null : p.id)}
                  className="p-5 flex flex-wrap items-center justify-between gap-4 cursor-pointer hover:bg-slate-50/70 transition"
                >
                  <div>
                    <div className="flex items-center gap-3">
                      <span className="font-mono font-bold text-slate-800 text-sm">{p.numero}</span>
                      <span className={`text-xs px-2.5 py-0.5 rounded-full border font-semibold ${getEstadoBadge(p.estado)}`}>
                        {p.estado}
                      </span>
                    </div>
                    <p className="text-xs text-slate-400 mt-1">
                      {new Date(p.fecha).toLocaleDateString()} · Método: {p.metodo_pago}
                    </p>
                  </div>

                  <div className="flex items-center gap-6">
                    <div className="text-right">
                      <p className="text-xs text-slate-400">Total</p>
                      <p className="text-lg font-bold text-slate-900">${p.total.toFixed(2)}</p>
                    </div>
                    <span className="text-slate-400 text-sm font-mono">
                      {isExpanded ? "▲" : "▼"}
                    </span>
                  </div>
                </div>

                {isExpanded && (
                  <div className="border-t border-slate-100 p-5 bg-slate-50/50 space-y-4 text-xs">
                    <div>
                      <p className="font-semibold text-slate-700 mb-1">Datos de entrega:</p>
                      <p className="text-slate-600">{p.cliente_nombre} ({p.cliente_email})</p>
                      <p className="text-slate-500">{p.cliente_direccion}, {p.cliente_ciudad}</p>
                      {p.cliente_telefono && <p className="text-slate-500">Tel: {p.cliente_telefono}</p>}
                    </div>

                    {p.items && p.items.length > 0 && (
                      <div>
                        <p className="font-semibold text-slate-700 mb-2">Artículos comprados:</p>
                        <div className="divide-y divide-slate-200 bg-white rounded-lg border border-slate-200 p-3">
                          {p.items.map((item, idx) => (
                            <div key={idx} className="py-2 flex justify-between items-center">
                              <div>
                                <span className="font-medium text-slate-800">{item.nombre}</span>
                                <span className="text-slate-400 ml-2">x{item.cantidad}</span>
                              </div>
                              <span className="font-semibold text-slate-700">
                                ${(item.precio * item.cantidad).toFixed(2)}
                              </span>
                            </div>
                          ))}
                        </div>
                      </div>
                    )}
                  </div>
                )}
              </div>
            );
          })}
        </div>
      )}
    </div>
  );
};

export default MisPedidos;
