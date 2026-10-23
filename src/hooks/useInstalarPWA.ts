import { useEffect, useState } from "react";

// Evento que lanzan Chrome, Edge y Android cuando la PWA se puede instalar.
// No está en los tipos estándar del DOM, por eso se declara aquí.
interface BeforeInstallPromptEvent extends Event {
  prompt: () => Promise<void>;
  userChoice: Promise<{ outcome: "accepted" | "dismissed" }>;
}

// El evento puede llegar antes de que se monte el componente (por ejemplo,
// en el login), así que se captura a nivel de módulo y se guarda.
let eventoGuardado: BeforeInstallPromptEvent | null = null;
const suscriptores = new Set<() => void>();
const avisar = () => suscriptores.forEach((fn) => fn());

window.addEventListener("beforeinstallprompt", (e) => {
  // Evita el aviso automático del navegador: la instalación la pide el botón.
  e.preventDefault();
  eventoGuardado = e as BeforeInstallPromptEvent;
  avisar();
});

window.addEventListener("appinstalled", () => {
  eventoGuardado = null;
  avisar();
});

const yaInstalada = () =>
  window.matchMedia("(display-mode: standalone)").matches ||
  (navigator as Navigator & { standalone?: boolean }).standalone === true;

// Instrucciones para cuando el navegador no ofrece la instalación directa
// (Safari, Firefox, o Chrome si el aviso se cerró antes y quedó silenciado).
const instruccionesManuales = () => {
  const ua = navigator.userAgent;
  if (/iphone|ipad|ipod/i.test(ua)) {
    return 'En iPhone o iPad: toca "Compartir" y luego "Agregar a inicio".';
  }
  if (/android/i.test(ua)) {
    return 'En Android: abre el menú ⋮ del navegador y elige "Instalar app" o "Agregar a pantalla de inicio".';
  }
  if (/firefox/i.test(ua)) {
    return "Firefox no permite instalar apps web en computadora. Abre esta página en Chrome o Edge para instalarla.";
  }
  if (/safari/i.test(ua) && !/chrome|chromium|edg/i.test(ua)) {
    return 'En Safari: menú Archivo > "Agregar al Dock".';
  }
  return 'En Chrome o Edge: usa el ícono de instalar a la derecha de la barra de direcciones, o el menú ⋮ > "Transmitir, guardar y compartir" > "Instalar MultiCatálogo".';
};

export function useInstalarPWA() {
  const [, forzar] = useState(0);

  useEffect(() => {
    const actualizar = () => forzar((n) => n + 1);
    suscriptores.add(actualizar);
    return () => {
      suscriptores.delete(actualizar);
    };
  }, []);

  const instalada = yaInstalada();
  // El botón se muestra siempre que la app no esté instalada.
  const puedeInstalar = !instalada;

  const instalar = async () => {
    if (eventoGuardado) {
      await eventoGuardado.prompt();
      await eventoGuardado.userChoice;
      eventoGuardado = null;
      avisar();
    } else {
      alert(`Para instalar MultiCatálogo:\n\n${instruccionesManuales()}`);
    }
  };

  return { puedeInstalar, instalar };
}
