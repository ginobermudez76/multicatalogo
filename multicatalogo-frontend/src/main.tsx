import { StrictMode } from 'react'
import { createRoot } from 'react-dom/client'
import { registerSW } from 'virtual:pwa-register'
import './index.css'
import App from './App.tsx'

// Registro automático del Service Worker con recarga periódica de verificación
registerSW({
  immediate: true,
  onNeedRefresh() {
    console.log('[PWA] Nueva versión disponible en el Service Worker.')
  },
  onOfflineReady() {
    console.log('[PWA] Aplicación lista para trabajar sin conexión (offline).')
  },
})

createRoot(document.getElementById('root')!).render(
  <StrictMode>
    <App />
  </StrictMode>,
)
