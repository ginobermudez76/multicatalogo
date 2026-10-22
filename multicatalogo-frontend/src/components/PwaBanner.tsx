// src/components/PwaBanner.tsx
import { useState, useEffect } from 'react';

interface BeforeInstallPromptEvent extends Event {
  prompt: () => Promise<void>;
  userChoice: Promise<{ outcome: 'accepted' | 'dismissed' }>;
}

export const PwaBanner = () => {
  const [deferredPrompt, setDeferredPrompt] = useState<BeforeInstallPromptEvent | null>(null);
  const [showInstallPrompt, setShowInstallPrompt] = useState(false);
  const [isOffline, setIsOffline] = useState(!navigator.onLine);

  useEffect(() => {
    // 1. Detección de conectividad online/offline
    const handleOnline = () => setIsOffline(false);
    const handleOffline = () => setIsOffline(true);

    window.addEventListener('online', handleOnline);
    window.addEventListener('offline', handleOffline);

    // 2. Captura del evento beforeinstallprompt de PWA
    const handleBeforeInstall = (e: Event) => {
      e.preventDefault();
      setDeferredPrompt(e as BeforeInstallPromptEvent);
      // Solo mostrar si no se ejecuta ya en modo standalone
      const isStandalone = window.matchMedia('(display-mode: standalone)').matches;
      if (!isStandalone) {
        setShowInstallPrompt(true);
      }
    };

    window.addEventListener('beforeinstallprompt', handleBeforeInstall);

    return () => {
      window.removeEventListener('online', handleOnline);
      window.removeEventListener('offline', handleOffline);
      window.removeEventListener('beforeinstallprompt', handleBeforeInstall);
    };
  }, []);

  const handleInstallClick = async () => {
    if (!deferredPrompt) return;
    await deferredPrompt.prompt();
    const { outcome } = await deferredPrompt.userChoice;
    if (outcome === 'accepted') {
      setShowInstallPrompt(false);
    }
    setDeferredPrompt(null);
  };

  return (
    <>
      {/* Banner de Estado Offline */}
      {isOffline && (
        <div className="bg-amber-600 text-white text-xs font-semibold px-4 py-2 text-center flex items-center justify-center gap-2 sticky top-0 z-50 shadow-md animate-fade-in">
          <svg className="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M18.364 5.636a9 9 0 010 12.728m0 0l-2.829-2.829m2.829 2.829L21 21M15.536 8.464a5 5 0 010 7.072m0 0l-2.829-2.829m-4.243 4.243a9 9 0 01-2.829-2.829m0 0L3 21m2.829-2.829A9 9 0 013 12c0-2.485 1.007-4.735 2.636-6.364M8.464 15.536A5 5 0 017 12c0-1.38.56-2.63 1.464-3.536m3.536-3.536A9 9 0 0112 3c2.485 0 4.735 1.007 6.364 2.636" />
          </svg>
          <span>Estás navegando en modo sin conexión. El catálogo y tus datos en caché siguen disponibles.</span>
        </div>
      )}

      {/* Notificación de Instalación Nativa PWA */}
      {showInstallPrompt && (
        <div className="fixed bottom-4 right-4 z-50 max-w-sm bg-white rounded-xl shadow-2xl border border-indigo-100 p-4 flex items-center gap-3 animate-slide-up">
          <img src="/pwa-192x192.png" alt="MultiCatálogo" className="w-12 h-12 rounded-xl object-cover shrink-0 shadow-sm" />
          <div className="flex-1 min-w-0">
            <p className="text-xs font-bold text-slate-800 truncate">Instalar MultiCatálogo</p>
            <p className="text-[11px] text-slate-500 leading-tight">Accede rápido desde tu pantalla de inicio y navega offline.</p>
          </div>
          <div className="flex flex-col gap-1.5 shrink-0">
            <button
              onClick={handleInstallClick}
              className="bg-indigo-600 hover:bg-indigo-700 text-white text-xs font-bold px-3 py-1.5 rounded-lg transition cursor-pointer shadow-xs"
            >
              Instalar
            </button>
            <button
              onClick={() => setShowInstallPrompt(false)}
              className="text-[10px] text-slate-400 hover:text-slate-600 font-medium text-center cursor-pointer"
            >
              Cerrar
            </button>
          </div>
        </div>
      )}
    </>
  );
};
