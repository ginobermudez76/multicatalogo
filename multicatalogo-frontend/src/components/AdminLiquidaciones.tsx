// src/components/AdminLiquidaciones.tsx
// Panel de administración de liquidaciones a distribuidores MLM (Fase 5 Back-office U2):
// - Consulta de comisiones acumuladas pendientes por distribuidor
// - Proceso transaccional de liquidación (POST /api/liquidaciones)
// - Historial auditable de liquidaciones emitidas
import { useState, useEffect } from "react";
import {
  getDeudaDistribuidoresApi,
  getLiquidacionesAdminApi,
  postCrearLiquidacionApi,
  type DeudaDistribuidorItem,
  type LiquidacionItem,
} from "../services/api";

const AdminLiquidaciones = () => {
  const [deudas, setDeudas] = useState<DeudaDistribuidorItem[]>([]);
  const [liquidaciones, setLiquidaciones] = useState<LiquidacionItem[]>([]);
  const [loading, setLoading] = useState<boolean>(true);
  const [procesandoId, setProcesandoId] = useState<number | null>(null);
  const [error, setError] = useState<string>("");
  const [mensajeExito, setMensajeExito] = useState<string>("");

  useEffect(() => {
    cargarDatos();
  }, []);

  const cargarDatos = async () => {
    setLoading(true);
    setError("");
    try {
      const [resDeudas, resLiq] = await Promise.all([
        getDeudaDistribuidoresApi(),
        getLiquidacionesAdminApi(),
      ]);
      setDeudas(resDeudas || []);
      setLiquidaciones(resLiq || []);
    } catch (err: unknown) {
      setError(err instanceof Error ? err.message : "Error al cargar datos de liquidaciones");
    } finally {
      setLoading(false);
    }
  };

  const handleLiquidar = async (referidoId: number, nombre: string, total: number) => {
    const totalFormateado = total.toLocaleString("es-EC", { minimumFractionDigits: 2, maximumFractionDigits: 2 });
    if (!window.confirm(`¿Confirmas la liquidación de $${totalFormateado} para el distribuidor "${nombre}"?`)) {
      return;
    }

    setProcesandoId(referidoId);
    setError("");
    try {
      const liq = await postCrearLiquidacionApi(referidoId);
      setMensajeExito(`Liquidación #${liq.id} por $${liq.total.toFixed(2)} procesada exitosamente.`);
      setTimeout(() => setMensajeExito(""), 4000);
      await cargarDatos();
    } catch (err: unknown) {
      setError(err instanceof Error ? err.message : "Error al procesar la liquidación");
    } finally {
      setProcesandoId(null);
    }
  };

  const totalDeudaPendiente = deudas.reduce((acc, d) => acc + d.total_pendiente, 0);
  const totalHistoricoLiquidado = liquidaciones.reduce((acc, l) => acc + l.total, 0);

  return (
    <div className="space-y-8">
      {/* Cabecera */}
      <div className="flex flex-col sm:flex-row sm:items-center sm:justify-between gap-4">
        <div>
          <h1 className="text-2xl font-bold text-slate-800">Liquidaciones y Pagos</h1>
          <p className="text-slate-500 text-sm mt-0.5">
            Gestión y cierre del ciclo de comisiones para la red de distribuidores
          </p>
        </div>
        <button
          type="button"
          onClick={cargarDatos}
          disabled={loading}
          className="inline-flex items-center gap-2 px-4 py-2 bg-white border border-slate-300 rounded-lg text-sm font-semibold text-slate-700 hover:bg-slate-50 transition shadow-sm"
        >
          <span>↻</span>
          <span>Actualizar saldos</span>
        </button>
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

      {/* Tarjetas de Métricas KPI */}
      <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-5">
        <div className="bg-white p-5 rounded-xl border border-slate-200 shadow-sm">
          <p className="text-xs font-semibold text-slate-500 uppercase tracking-wider">Deuda por Liquidar</p>
          <p className="text-2xl font-bold text-amber-600 mt-2">
            ${totalDeudaPendiente.toLocaleString("es-EC", { minimumFractionDigits: 2, maximumFractionDigits: 2 })}
          </p>
          <p className="text-xs text-slate-400 mt-1">Comisiones pendientes acumuladas</p>
        </div>

        <div className="bg-white p-5 rounded-xl border border-slate-200 shadow-sm">
          <p className="text-xs font-semibold text-slate-500 uppercase tracking-wider">Distribuidores con Deuda</p>
          <p className="text-2xl font-bold text-indigo-600 mt-2">{deudas.length}</p>
          <p className="text-xs text-slate-400 mt-1">Acreedores de la red MLM</p>
        </div>

        <div className="bg-white p-5 rounded-xl border border-slate-200 shadow-sm">
          <p className="text-xs font-semibold text-slate-500 uppercase tracking-wider">Total Liquidado</p>
          <p className="text-2xl font-bold text-emerald-600 mt-2">
            ${totalHistoricoLiquidado.toLocaleString("es-EC", { minimumFractionDigits: 2, maximumFractionDigits: 2 })}
          </p>
          <p className="text-xs text-slate-400 mt-1">Pagos completados históricamente</p>
        </div>

        <div className="bg-white p-5 rounded-xl border border-slate-200 shadow-sm">
          <p className="text-xs font-semibold text-slate-500 uppercase tracking-wider">Liquidaciones Emitidas</p>
          <p className="text-2xl font-bold text-slate-800 mt-2">{liquidaciones.length}</p>
          <p className="text-xs text-slate-400 mt-1">Registros en el ledger</p>
        </div>
      </div>

      {/* SECCIÓN 1: Deuda por Distribuidor y Botón Liquidar */}
      <div className="bg-white rounded-xl border border-slate-200 shadow-sm overflow-hidden">
        <div className="px-6 py-4 bg-slate-50 border-b border-slate-200 flex items-center justify-between">
          <div>
            <h2 className="text-base font-bold text-slate-800">Comisiones Pendientes por Liquidar</h2>
            <p className="text-xs text-slate-500 mt-0.5">
              Distribuidores con saldo devengado listo para corte y pago
            </p>
          </div>
          <span className="text-xs font-semibold px-2.5 py-1 rounded-full bg-amber-100 text-amber-800 border border-amber-200">
            {deudas.length} pendientes
          </span>
        </div>

        {loading ? (
          <div className="p-10 text-center text-slate-500">
            <div className="inline-block animate-spin rounded-full h-8 w-8 border-4 border-indigo-600 border-t-transparent mb-3" />
            <p>Calculando deuda acumulada...</p>
          </div>
        ) : deudas.length === 0 ? (
          <div className="p-10 text-center text-slate-500">
            <div className="text-4xl mb-2">🎉</div>
            <p className="text-base font-semibold text-slate-700">¡Al día! No hay comisiones pendientes por liquidar</p>
            <p className="text-xs text-slate-400 mt-1">
              Las comisiones se generarán automáticamente cuando los referidos realicen compras en la tienda
            </p>
          </div>
        ) : (
          <div className="overflow-x-auto">
            <table className="w-full text-left text-sm text-slate-600">
              <thead className="bg-slate-50/50 text-slate-600 uppercase font-semibold text-xs border-b border-slate-200">
                <tr>
                  <th className="px-6 py-3.5">Distribuidor</th>
                  <th className="px-6 py-3.5">Código Referido</th>
                  <th className="px-6 py-3.5 text-center">Nº Comisiones</th>
                  <th className="px-6 py-3.5 text-right">Saldo Pendiente</th>
                  <th className="px-6 py-3.5 text-right">Acción</th>
                </tr>
              </thead>
              <tbody className="divide-y divide-slate-100">
                {deudas.map((d) => (
                  <tr key={d.referido_id} className="hover:bg-slate-50/80 transition">
                    <td className="px-6 py-4">
                      <div className="flex items-center gap-3">
                        <div className="w-9 h-9 rounded-full bg-indigo-100 text-indigo-700 font-bold flex items-center justify-center text-xs">
                          {d.nombre.charAt(0)}
                        </div>
                        <div>
                          <p className="font-semibold text-slate-800">{d.nombre}</p>
                          <p className="text-xs text-slate-400">ID Referido: #{d.referido_id}</p>
                        </div>
                      </div>
                    </td>
                    <td className="px-6 py-4 font-mono text-xs">
                      <span className="bg-slate-100 px-2 py-0.5 rounded border border-slate-200 text-slate-700 font-semibold">
                        {d.codigo || `REF-${d.referido_id}`}
                      </span>
                    </td>
                    <td className="px-6 py-4 text-center">
                      <span className="inline-block px-2 py-0.5 rounded-full text-xs font-medium bg-slate-100 text-slate-700">
                        {d.num_comisiones} {d.num_comisiones === 1 ? "comisión" : "comisiones"}
                      </span>
                    </td>
                    <td className="px-6 py-4 text-right">
                      <span className="font-bold text-base text-emerald-600">
                        ${d.total_pendiente.toFixed(2)}
                      </span>
                    </td>
                    <td className="px-6 py-4 text-right">
                      <button
                        type="button"
                        disabled={procesandoId === d.referido_id}
                        onClick={() => handleLiquidar(d.referido_id, d.nombre, d.total_pendiente)}
                        className="inline-flex items-center gap-1.5 px-3.5 py-1.5 bg-emerald-600 hover:bg-emerald-700 text-white rounded-lg text-xs font-semibold shadow-sm transition disabled:opacity-50"
                      >
                        {procesandoId === d.referido_id ? (
                          <>
                            <span className="animate-spin text-xs">⏳</span>
                            <span>Procesando...</span>
                          </>
                        ) : (
                          <>
                            <span>💵</span>
                            <span>Liquidar pago</span>
                          </>
                        )}
                      </button>
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        )}
      </div>

      {/* SECCIÓN 2: Historial de Liquidaciones Emitidas */}
      <div className="bg-white rounded-xl border border-slate-200 shadow-sm overflow-hidden">
        <div className="px-6 py-4 bg-slate-50 border-b border-slate-200 flex items-center justify-between">
          <div>
            <h2 className="text-base font-bold text-slate-800">Historial de Liquidaciones Emitidas</h2>
            <p className="text-xs text-slate-500 mt-0.5">
              Registro contable de todos los cortes y pagos de comisiones procesados
            </p>
          </div>
          <span className="text-xs font-semibold px-2.5 py-1 rounded-full bg-slate-200 text-slate-700">
            {liquidaciones.length} registros
          </span>
        </div>

        {liquidaciones.length === 0 ? (
          <div className="p-8 text-center text-slate-400 text-sm">
            Aún no se han emitido liquidaciones en el sistema.
          </div>
        ) : (
          <div className="overflow-x-auto">
            <table className="w-full text-left text-sm text-slate-600">
              <thead className="bg-slate-50/50 text-slate-600 uppercase font-semibold text-xs border-b border-slate-200">
                <tr>
                  <th className="px-6 py-3.5">Nº Liquidación</th>
                  <th className="px-6 py-3.5">Distribuidor</th>
                  <th className="px-6 py-3.5">Monto Liquidado</th>
                  <th className="px-6 py-3.5">Estado</th>
                  <th className="px-6 py-3.5">Fecha de Emisión</th>
                </tr>
              </thead>
              <tbody className="divide-y divide-slate-100">
                {liquidaciones.map((l) => {
                  const fecha = new Date(l.creado_en).toLocaleString("es-EC", {
                    dateStyle: "medium",
                    timeStyle: "short",
                  });

                  return (
                    <tr key={l.id} className="hover:bg-slate-50/80 transition">
                      <td className="px-6 py-3.5 font-mono text-xs font-bold text-slate-700">
                        #LIQ-{l.id.toString().padStart(4, "0")}
                      </td>
                      <td className="px-6 py-3.5">
                        <span className="font-semibold text-slate-800">
                          {l.referido_nombre || `Distribuidor #${l.referido_id}`}
                        </span>
                      </td>
                      <td className="px-6 py-3.5 font-bold text-slate-800">
                        ${l.total.toFixed(2)}
                      </td>
                      <td className="px-6 py-3.5">
                        <span className="inline-flex items-center gap-1 text-xs font-semibold px-2.5 py-0.5 rounded-full bg-emerald-50 text-emerald-700 border border-emerald-200">
                          <span>✓</span>
                          <span>Procesada</span>
                        </span>
                      </td>
                      <td className="px-6 py-3.5 text-xs text-slate-500">{fecha}</td>
                    </tr>
                  );
                })}
              </tbody>
            </table>
          </div>
        )}
      </div>
    </div>
  );
};

export default AdminLiquidaciones;
