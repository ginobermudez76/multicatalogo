// src/components/Registro.tsx
import { useState, type FormEvent, useEffect } from "react";
import { useNavigate, useSearchParams, Link } from "react-router-dom";
import { registerApi } from "../services/api";

const Registro = () => {
  const [searchParams] = useSearchParams();
  const [nombre, setNombre] = useState<string>("");
  const [email, setEmail] = useState<string>("");
  const [password, setPassword] = useState<string>("");
  const [ref, setRef] = useState<string>("");
  const [error, setError] = useState<string>("");
  const [success, setSuccess] = useState<string>("");
  const [loading, setLoading] = useState<boolean>(false);

  const navigate = useNavigate();

  useEffect(() => {
    const refParam = searchParams.get("ref");
    if (refParam) {
      setRef(refParam.toUpperCase());
    }
  }, [searchParams]);

  const handleSubmit = async (e: FormEvent<HTMLFormElement>) => {
    e.preventDefault();
    setError("");
    setSuccess("");
    setLoading(true);

    try {
      await registerApi({
        nombre: nombre.trim(),
        email: email.trim(),
        password,
        ref: ref.trim() ? ref.trim().toUpperCase() : undefined,
      });

      setSuccess("¡Cuenta creada con éxito! Redirigiendo al inicio de sesión...");
      setTimeout(() => {
        navigate("/login");
      }, 1500);
    } catch (err: unknown) {
      setError(err instanceof Error ? err.message : "Error al registrar la cuenta");
    } finally {
      setLoading(false);
    }
  };

  return (
    <div className="min-h-screen flex items-center justify-center bg-slate-100 px-4 py-8">
      <div className="max-w-md w-full bg-white rounded-xl shadow-md p-8 border border-slate-200">
        <div className="text-center mb-8">
          <h2 className="text-3xl font-bold text-slate-900">Crear Cuenta</h2>
          <p className="text-slate-500 mt-2">
            Únete a la red y catálogo de MultiCatálogo
          </p>
        </div>

        {error && (
          <div className="bg-red-50 text-red-600 p-3 rounded-lg text-sm mb-6 text-center border border-red-200">
            {error}
          </div>
        )}

        {success && (
          <div className="bg-emerald-50 text-emerald-700 p-3 rounded-lg text-sm mb-6 text-center border border-emerald-200 font-medium">
            {success}
          </div>
        )}

        <form onSubmit={handleSubmit} className="space-y-4">
          <div>
            <label className="block text-sm font-medium text-slate-700 mb-1">
              Nombre Completo
            </label>
            <input
              type="text"
              value={nombre}
              onChange={(e) => setNombre(e.target.value)}
              className="w-full px-4 py-2.5 rounded-lg border border-slate-300 focus:ring-2 focus:ring-indigo-600 focus:border-indigo-600 outline-none transition"
              placeholder="Juan Pérez"
              required
            />
          </div>

          <div>
            <label className="block text-sm font-medium text-slate-700 mb-1">
              Correo Electrónico
            </label>
            <input
              type="email"
              value={email}
              onChange={(e) => setEmail(e.target.value)}
              className="w-full px-4 py-2.5 rounded-lg border border-slate-300 focus:ring-2 focus:ring-indigo-600 focus:border-indigo-600 outline-none transition"
              placeholder="usuario@upse.edu.ec"
              required
            />
          </div>

          <div>
            <label className="block text-sm font-medium text-slate-700 mb-1">
              Contraseña
            </label>
            <input
              type="password"
              value={password}
              onChange={(e) => setPassword(e.target.value)}
              className="w-full px-4 py-2.5 rounded-lg border border-slate-300 focus:ring-2 focus:ring-indigo-600 focus:border-indigo-600 outline-none transition"
              placeholder="••••••"
              minLength={6}
              required
            />
          </div>

          <div>
            <label className="block text-sm font-medium text-slate-700 mb-1">
              Código de Referido (opcional)
            </label>
            <input
              type="text"
              value={ref}
              onChange={(e) => setRef(e.target.value.toUpperCase())}
              className="w-full px-4 py-2.5 rounded-lg border border-slate-300 focus:ring-2 focus:ring-indigo-600 focus:border-indigo-600 outline-none transition uppercase tracking-wider font-mono text-sm"
              placeholder="LUIS / ANA / TU"
            />
            <p className="text-xs text-slate-500 mt-1">
              Si fuiste invitado por un distribuidor, ingresa su código aquí.
            </p>
          </div>

          <button
            type="submit"
            disabled={loading}
            className="w-full mt-2 bg-indigo-600 text-white font-bold py-3 rounded-lg hover:bg-indigo-700 transition disabled:opacity-50 disabled:cursor-not-allowed cursor-pointer shadow-md"
          >
            {loading ? "Creando cuenta..." : "Registrarme"}
          </button>
        </form>

        <div className="mt-6 text-center text-sm text-slate-600 border-t border-slate-200 pt-4">
          ¿Ya tienes una cuenta?{" "}
          <Link to="/login" className="font-semibold text-indigo-600 hover:text-indigo-800 transition">
            Iniciar Sesión
          </Link>
        </div>
      </div>
    </div>
  );
};

export default Registro;
