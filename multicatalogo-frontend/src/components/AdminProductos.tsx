// src/components/AdminProductos.tsx
import { useState, useEffect, type FormEvent } from "react";
import {
  getProductosAdminApi,
  crearProductoAdminApi,
  actualizarProductoAdminApi,
  eliminarProductoAdminApi,
  type Producto,
} from "../services/api";

const AdminProductos = () => {
  const [productos, setProductos] = useState<Producto[]>([]);
  const [loading, setLoading] = useState<boolean>(true);
  const [busqueda, setBusqueda] = useState<string>("");
  const [error, setError] = useState<string>("");
  const [modalAbierto, setModalAbierto] = useState<boolean>(false);
  const [productoEditando, setProductoEditando] = useState<Producto | null>(null);

  // Formulario
  const [nombre, setNombre] = useState<string>("");
  const [categoria, setCategoria] = useState<string>("");
  const [precio, setPrecio] = useState<number>(0);
  const [stock, setStock] = useState<number>(0);
  const [img, setImg] = useState<string>("");
  const [descripcion, setDescripcion] = useState<string>("");

  useEffect(() => {
    cargarProductos();
  }, [busqueda]);

  const cargarProductos = () => {
    setLoading(true);
    getProductosAdminApi(1, 50, busqueda)
      .then((res) => {
        setProductos(res.data || []);
      })
      .catch((err: unknown) => {
        setError(err instanceof Error ? err.message : "Error al cargar catálogo admin");
      })
      .finally(() => {
        setLoading(false);
      });
  };

  const abrirCrear = () => {
    setProductoEditando(null);
    setNombre("");
    setCategoria("Belleza");
    setPrecio(10);
    setStock(20);
    setImg("https://images.unsplash.com/photo-1522337360788-8b13dee7a37e?w=800");
    setDescripcion("");
    setModalAbierto(true);
  };

  const abrirEditar = (prod: Producto) => {
    setProductoEditando(prod);
    setNombre(prod.nombre);
    setCategoria(prod.categoria || "");
    setPrecio(prod.precio);
    setStock(prod.stock ?? 0);
    setImg(prod.img);
    setDescripcion(prod.descripcion || "");
    setModalAbierto(true);
  };

  const handleSubmit = async (e: FormEvent) => {
    e.preventDefault();
    try {
      if (productoEditando) {
        await actualizarProductoAdminApi(productoEditando.id, {
          nombre,
          categoria,
          precio: Number(precio),
          stock: Number(stock),
          img,
          descripcion,
        });
      } else {
        await crearProductoAdminApi({
          nombre,
          categoria,
          precio: Number(precio),
          stock: Number(stock),
          img,
          descripcion,
        });
      }
      setModalAbierto(false);
      cargarProductos();
    } catch (err: unknown) {
      alert(err instanceof Error ? err.message : "Error al guardar producto");
    }
  };

  const handleToggleActivo = async (prod: Producto) => {
    const confirm = window.confirm(`¿Seguro que deseas desactivar '${prod.nombre}' del catálogo público?`);
    if (!confirm) return;
    try {
      await eliminarProductoAdminApi(prod.id);
      cargarProductos();
    } catch (err: unknown) {
      alert(err instanceof Error ? err.message : "Error al cambiar estado");
    }
  };

  return (
    <div className="space-y-6">
      <div className="flex flex-col sm:flex-row justify-between items-start sm:items-center gap-4">
        <div>
          <h1 className="text-2xl font-bold text-slate-800">Administración de Productos</h1>
          <p className="text-slate-500 text-sm">Gestiona el catálogo, precios, stock y visibilidad (Fase 2 U2)</p>
        </div>
        <button
          onClick={abrirCrear}
          className="bg-indigo-600 hover:bg-indigo-700 text-white font-semibold text-xs py-2.5 px-4 rounded-lg cursor-pointer transition shadow-sm"
        >
          + Nuevo Producto
        </button>
      </div>

      {error && <div className="p-3 bg-red-50 text-red-600 rounded-lg text-xs">{error}</div>}

      {/* Barra de búsqueda */}
      <div className="bg-white p-4 rounded-xl border border-slate-200">
        <input
          type="text"
          placeholder="Buscar por nombre o categoría (ILIKE server-side)..."
          value={busqueda}
          onChange={(e) => setBusqueda(e.target.value)}
          className="w-full text-xs px-3 py-2 border border-slate-300 rounded-lg outline-none focus:border-indigo-500"
        />
      </div>

      {/* Tabla de productos */}
      <div className="bg-white rounded-xl border border-slate-200 shadow-sm overflow-hidden">
        {loading ? (
          <div className="p-8 text-center text-xs text-slate-400">Cargando inventario...</div>
        ) : (
          <div className="overflow-x-auto">
            <table className="w-full text-left text-xs border-collapse">
              <thead>
                <tr className="border-b border-slate-200 bg-slate-50 text-slate-600 uppercase font-semibold">
                  <th className="py-3 px-4">Img</th>
                  <th className="py-3 px-4">Nombre</th>
                  <th className="py-3 px-4">Categoría</th>
                  <th className="py-3 px-4">Precio</th>
                  <th className="py-3 px-4">Stock</th>
                  <th className="py-3 px-4">Estado</th>
                  <th className="py-3 px-4 text-right">Acciones</th>
                </tr>
              </thead>
              <tbody className="divide-y divide-slate-100 text-slate-700">
                {productos.map((p) => (
                  <tr key={p.id} className="hover:bg-slate-50/60">
                    <td className="py-3 px-4">
                      <img src={p.img} alt={p.nombre} className="w-10 h-10 object-cover rounded-lg" />
                    </td>
                    <td className="py-3 px-4 font-semibold text-slate-800">{p.nombre}</td>
                    <td className="py-3 px-4">{p.categoria || "—"}</td>
                    <td className="py-3 px-4 font-bold text-slate-900">${p.precio.toFixed(2)}</td>
                    <td className="py-3 px-4">
                      <span className={`font-mono ${p.stock && p.stock < 10 ? "text-red-600 font-bold" : ""}`}>
                        {p.stock ?? 0}
                      </span>
                    </td>
                    <td className="py-3 px-4">
                      <span
                        className={`px-2 py-0.5 rounded-full text-[11px] font-semibold ${
                          p.activo !== false
                            ? "bg-emerald-100 text-emerald-700"
                            : "bg-red-100 text-red-700"
                        }`}
                      >
                        {p.activo !== false ? "Activo" : "Inactivo"}
                      </span>
                    </td>
                    <td className="py-3 px-4 text-right space-x-2">
                      <button
                        onClick={() => abrirEditar(p)}
                        className="text-indigo-600 hover:text-indigo-800 font-medium cursor-pointer"
                      >
                        Editar
                      </button>
                      {p.activo !== false && (
                        <button
                          onClick={() => handleToggleActivo(p)}
                          className="text-red-500 hover:text-red-700 font-medium cursor-pointer"
                        >
                          Desactivar
                        </button>
                      )}
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        )}
      </div>

      {/* Modal Crear/Editar */}
      {modalAbierto && (
        <div className="fixed inset-0 bg-slate-900/60 backdrop-blur-xs z-50 flex items-center justify-center p-4">
          <div className="bg-white rounded-xl shadow-xl max-w-lg w-full p-6 space-y-4">
            <h3 className="text-lg font-bold text-slate-800">
              {productoEditando ? "Editar Producto" : "Nuevo Producto"}
            </h3>

            <form onSubmit={handleSubmit} className="space-y-3 text-xs">
              <div>
                <label className="block text-slate-700 font-semibold mb-1">Nombre</label>
                <input
                  type="text"
                  required
                  value={nombre}
                  onChange={(e) => setNombre(e.target.value)}
                  className="w-full px-3 py-2 border border-slate-300 rounded-lg outline-none focus:border-indigo-600"
                />
              </div>

              <div className="grid grid-cols-2 gap-3">
                <div>
                  <label className="block text-slate-700 font-semibold mb-1">Categoría</label>
                  <input
                    type="text"
                    required
                    value={categoria}
                    onChange={(e) => setCategoria(e.target.value)}
                    className="w-full px-3 py-2 border border-slate-300 rounded-lg outline-none focus:border-indigo-600"
                  />
                </div>
                <div>
                  <label className="block text-slate-700 font-semibold mb-1">Precio ($)</label>
                  <input
                    type="number"
                    step="0.01"
                    min="0.01"
                    required
                    value={precio}
                    onChange={(e) => setPrecio(parseFloat(e.target.value) || 0)}
                    className="w-full px-3 py-2 border border-slate-300 rounded-lg outline-none focus:border-indigo-600"
                  />
                </div>
              </div>

              <div className="grid grid-cols-2 gap-3">
                <div>
                  <label className="block text-slate-700 font-semibold mb-1">Stock</label>
                  <input
                    type="number"
                    min="0"
                    required
                    value={stock}
                    onChange={(e) => setStock(parseInt(e.target.value) || 0)}
                    className="w-full px-3 py-2 border border-slate-300 rounded-lg outline-none focus:border-indigo-600"
                  />
                </div>
                <div>
                  <label className="block text-slate-700 font-semibold mb-1">URL Imagen</label>
                  <input
                    type="url"
                    required
                    value={img}
                    onChange={(e) => setImg(e.target.value)}
                    className="w-full px-3 py-2 border border-slate-300 rounded-lg outline-none focus:border-indigo-600"
                  />
                </div>
              </div>

              <div>
                <label className="block text-slate-700 font-semibold mb-1">Descripción</label>
                <textarea
                  rows={2}
                  value={descripcion}
                  onChange={(e) => setDescripcion(e.target.value)}
                  className="w-full px-3 py-2 border border-slate-300 rounded-lg outline-none focus:border-indigo-600"
                />
              </div>

              <div className="flex justify-end gap-3 pt-3">
                <button
                  type="button"
                  onClick={() => setModalAbierto(false)}
                  className="px-4 py-2 border border-slate-300 rounded-lg text-slate-700 font-semibold hover:bg-slate-50 cursor-pointer"
                >
                  Cancelar
                </button>
                <button
                  type="submit"
                  className="px-4 py-2 bg-indigo-600 text-white rounded-lg font-semibold hover:bg-indigo-700 cursor-pointer"
                >
                  Guardar
                </button>
              </div>
            </form>
          </div>
        </div>
      )}
    </div>
  );
};

export default AdminProductos;
