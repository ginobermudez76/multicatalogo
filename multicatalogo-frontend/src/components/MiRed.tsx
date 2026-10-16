// src/components/MiRed.tsx
// Red multinivel visual (Fase 4 U2):
// - Resumen: referidos activos, ventas de la red, comisiones del mes y nivel alcanzado
// - Link de referido propio (?ref=CODIGO) con botón copiar
// - Árbol jerárquico recursivo de referidos con comisión calculada por nivel
// - Ledger de comisiones auditable (Mis comisiones: monto, nivel, estado)
import { useState, useEffect } from "react";
import {
  redInicial,
  contarRed,
  sumarVentasRed,
  sumarComisiones,
  nivelAlcanzado,
  comisionDeReferido,
  TASA_COMISION,
  type Referido,
} from "../data/red";
import { getMiRedPropiaApi, getMisComisionesApi } from "../services/api";

const NIVEL_ESTILO: Record<number, { badge: string; border: string; tasa: string }> = {
  1: { badge: "bg-indigo-100 text-indigo-700", border: "border-indigo-200", tasa: "10 %" },
  2: { badge: "bg-purple-100 text-purple-700", border: "border-purple-200", tasa: "5 %" },
  3: { badge: "bg-pink-100 text-pink-700", border: "border-pink-200", tasa: "2 %" },
};

interface ComisionItem {
  id: number;
  pedido_id: number;
  referido_id: number;
  nivel: number;
  porcentaje: number;
  monto: number;
  estado: string;
  creado_en?: string;
}

const NodoReferido = ({ nodo }: { nodo: Referido }) => {
  const estilo = NIVEL_ESTILO[nodo.nivel] ?? NIVEL_ESTILO[1];
  const comision = comisionDeReferido(nodo);

  return (
    <li>
      <div className={`bg-white rounded-lg border ${estilo.border} shadow-sm p-4 flex flex-wrap items-center justify-between gap-3`}>
        <div className="flex items-center gap-3">
          <div className="w-10 h-10 rounded-full bg-slate-200 flex items-center justify-center font-bold text-slate-600">
            {nodo.nombre.charAt(0)}
          </div>
          <div>
            <div className="flex items-center gap-2">
              <p className="font-semibold text-slate-800">{nodo.nombre}</p>
              {nodo.codigo && (
                <span className="text-[10px] bg-slate-100 font-mono px-1.5 py-0.5 rounded border border-slate-200 text-slate-600">
                  REF: {nodo.codigo}
                </span>
              )}
            </div>
            <span className={`inline-block mt-1 text-xs font-semibold px-2 py-0.5 rounded-full ${estilo.badge}`}>
              Nivel {nodo.nivel} · {estilo.tasa} comisión
            </span>
          </div>
        </div>
        <div className="text-right">
          <p className="text-sm text-slate-500">Ventas</p>
          <p className="font-bold text-slate-800">${nodo.ventas.toLocaleString()}</p>
          <p className="text-xs text-indigo-600 font-semibold mt-1">
            +${comision.toFixed(2)} para ti
          </p>
        </div>
      </div>

      {nodo.hijos && nodo.hijos.length > 0 && (
        <ul className="mt-3 ml-6 space-y-3 border-l-2 border-slate-200 pl-4">
          {nodo.hijos.map((hijo) => (
            <NodoReferido key={hijo.id} nodo={hijo} />
          ))}
        </ul>
      )}
    </li>
  );
};

const MiRed = () => {
  const [red, setRed] = useState<Referido>(redInicial);
  const [comisiones, setComisiones] = useState<ComisionItem[]>([]);
  const [totalPendiente, setTotalPendiente] = useState<number>(0);
  const [totalPagado, setTotalPagado] = useState<number>(0);
  const [copiado, setCopiado] = useState<boolean>(false);

  useEffect(() => {
    // Intentamos cargar la red viva del usuario desde la API U2
    getMiRedPropiaApi()
      .then((data) => {
        if (data && data.nombre) {
          setRed(data);
        }
      })
      .catch(() => {
        // Fallback gracioso a redInicial
      });

    // Cargar ledger de comisiones
    getMisComisionesApi()
      .then((data) => {
        if (data) {
          setComisiones(data.comisiones || []);
          setTotalPendiente(data.total_pendiente || 0);
          setTotalPagado(data.total_pagado || 0);
        }
      })
      .catch(() => {
        // No bloqueante
      });
  }, []);

  const referidosActivos = contarRed(red);
  const ventasRed = sumarVentasRed(red);
  const comisionesMes = sumarComisiones(red);
  const nivel = nivelAlcanzado(red.hijos?.length ?? 0);

  const codigoReferido = red.codigo || "TU";
  const linkReferido = `${window.location.origin}/registro?ref=${codigoReferido}`;

  const handleCopiarLink = () => {
    navigator.clipboard.writeText(linkReferido).then(() => {
      setCopiado(true);
      setTimeout(() => setCopiado(false), 2000);
    });
  };

  const resumen = [
    { etiqueta: "Referidos Activos", valor: referidosActivos.toString(), color: "text-indigo-600" },
    { etiqueta: "Ventas de la Red", valor: `$${ventasRed.toLocaleString()}`, color: "text-indigo-600" },
    { etiqueta: "Comisiones Estimadas", valor: `$${comisionesMes.toFixed(2)}`, color: "text-emerald-600" },
    { etiqueta: "Nivel Alcanzado", valor: nivel, color: "text-amber-600" },
  ];

  return (
    <div className="space-y-8">
      <div>
        <h1 className="text-2xl font-bold text-slate-800 mb-2">Mi Red de Referidos</h1>
        <p className="text-slate-500">
          Estructura multinivel: gana el {TASA_COMISION[1] * 100} % de nivel 1, {TASA_COMISION[2] * 100} % de nivel 2 y {TASA_COMISION[3] * 100} % de nivel 3.
        </p>
      </div>

      {/* ===== ENLACE DE INVITACIÓN / REFERIDO ===== */}
      <div className="bg-gradient-to-r from-indigo-50 to-purple-50 rounded-xl p-5 border border-indigo-100 flex flex-col md:flex-row items-start md:items-center justify-between gap-4">
        <div>
          <span className="text-xs uppercase font-bold tracking-wider text-indigo-700 bg-indigo-100 px-2.5 py-1 rounded-md">
            Tu Link de Afiliado
          </span>
          <p className="text-sm font-semibold text-slate-800 mt-2">
            Invita a distribuidores y gana comisiones de por vida sobre sus compras y ventas
          </p>
          <p className="text-xs text-slate-500 mt-0.5">
            Código: <span className="font-mono font-bold text-indigo-600">{codigoReferido}</span>
          </p>
        </div>

        <div className="flex items-center gap-2 w-full md:w-auto">
          <input
            type="text"
            readOnly
            value={linkReferido}
            className="text-xs font-mono bg-white border border-slate-300 rounded-lg px-3 py-2 text-slate-600 w-full md:w-80 select-all"
          />
          <button
            type="button"
            onClick={handleCopiarLink}
            className={`px-4 py-2 text-xs font-bold rounded-lg cursor-pointer transition shrink-0 ${
              copiado
                ? "bg-emerald-600 text-white"
                : "bg-indigo-600 text-white hover:bg-indigo-700"
            }`}
          >
            {copiado ? "¡Copiado!" : "Copiar Link"}
          </button>
        </div>
      </div>

      {/* ===== TARJETAS DE RESUMEN ===== */}
      <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-6">
        {resumen.map((item) => (
          <div key={item.etiqueta} className="bg-white p-6 rounded-lg shadow-sm border border-slate-100">
            <p className="text-sm text-slate-500 uppercase font-semibold">{item.etiqueta}</p>
            <p className={`text-3xl font-bold mt-2 ${item.color}`}>{item.valor}</p>
          </div>
        ))}
      </div>

      {/* ===== ÁRBOL DE REFERIDOS ===== */}
      <div className="bg-white rounded-lg border border-slate-200 shadow-sm p-6">
        <h2 className="text-lg font-bold text-slate-800 mb-4">Jerarquía de tu Red</h2>

        {/* Raíz: el usuario */}
        <div className="bg-slate-900 text-white rounded-lg p-4 mb-4 flex flex-wrap items-center justify-between gap-3">
          <div className="flex items-center gap-3">
            <div className="w-10 h-10 rounded-full bg-indigo-500 flex items-center justify-center font-bold">
              {red.nombre.charAt(0)}
            </div>
            <div>
              <p className="font-semibold">{red.nombre}</p>
              <p className="text-xs text-slate-300">Nivel alcanzado: {nivel}</p>
            </div>
          </div>
          <div className="text-right">
            <p className="text-xs text-slate-400">Total ventas red</p>
            <p className="font-bold text-emerald-400">${ventasRed.toLocaleString()}</p>
          </div>
        </div>

        {red.hijos && red.hijos.length > 0 ? (
          <ul className="space-y-3">
            {red.hijos.map((hijo) => (
              <NodoReferido key={hijo.id} nodo={hijo} />
            ))}
          </ul>
        ) : (
          <div className="p-8 text-center text-slate-500 border border-dashed border-slate-200 rounded-lg">
            <p className="text-sm">Aún no tienes referidos registrados bajo tu código.</p>
            <p className="text-xs text-slate-400 mt-1">Comparte tu link de afiliado arriba para empezar a construir tu red.</p>
          </div>
        )}
      </div>

      {/* ===== LEDGER AUDITABLE DE COMISIONES (FASE 4) ===== */}
      <div className="bg-white rounded-lg border border-slate-200 shadow-sm p-6">
        <div className="flex flex-wrap items-center justify-between gap-3 mb-4">
          <div>
            <h2 className="text-lg font-bold text-slate-800">Ledger de Comisiones</h2>
            <p className="text-xs text-slate-500">Historial transaccional auditable de comisiones devengadas</p>
          </div>
          <div className="flex items-center gap-4 text-xs font-semibold">
            <span className="text-amber-700 bg-amber-50 px-3 py-1.5 rounded-lg border border-amber-200">
              Pendiente: ${totalPendiente.toFixed(2)}
            </span>
            <span className="text-emerald-700 bg-emerald-50 px-3 py-1.5 rounded-lg border border-emerald-200">
              Liquidado/Pagado: ${totalPagado.toFixed(2)}
            </span>
          </div>
        </div>

        {comisiones.length > 0 ? (
          <div className="overflow-x-auto">
            <table className="w-full text-left text-xs border-collapse">
              <thead>
                <tr className="border-b border-slate-200 bg-slate-50 text-slate-600 uppercase font-semibold">
                  <th className="py-2.5 px-3">#</th>
                  <th className="py-2.5 px-3">Pedido ID</th>
                  <th className="py-2.5 px-3">Nivel MLM</th>
                  <th className="py-2.5 px-3">Tasa (%)</th>
                  <th className="py-2.5 px-3">Monto</th>
                  <th className="py-2.5 px-3">Estado</th>
                </tr>
              </thead>
              <tbody className="divide-y divide-slate-100 text-slate-700">
                {comisiones.map((c) => (
                  <tr key={c.id} className="hover:bg-slate-50/60">
                    <td className="py-2.5 px-3 font-mono text-slate-400">{c.id}</td>
                    <td className="py-2.5 px-3 font-semibold">#{c.pedido_id}</td>
                    <td className="py-2.5 px-3">
                      <span className="bg-indigo-50 text-indigo-700 px-2 py-0.5 rounded font-semibold">
                        Nivel {c.nivel}
                      </span>
                    </td>
                    <td className="py-2.5 px-3">{c.porcentaje}%</td>
                    <td className="py-2.5 px-3 font-bold text-emerald-600">+${c.monto.toFixed(2)}</td>
                    <td className="py-2.5 px-3">
                      <span
                        className={`px-2 py-0.5 rounded font-medium ${
                          c.estado === "pagada"
                            ? "bg-emerald-100 text-emerald-700"
                            : c.estado === "revertida"
                            ? "bg-red-100 text-red-700"
                            : "bg-amber-100 text-amber-700"
                        }`}
                      >
                        {c.estado}
                      </span>
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        ) : (
          <p className="text-xs text-slate-400 text-center py-6">
            Aún no hay comisiones registradas en tu ledger. Se generarán automáticamente con las compras de tu red.
          </p>
        )}
      </div>
    </div>
  );
};

export default MiRed;
