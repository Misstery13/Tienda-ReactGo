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

// iPhone y iPad no lanzan el evento: allí se instala desde "Compartir".
const esIOS = () => /iphone|ipad|ipod/i.test(navigator.userAgent);

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
  const puedeInstalar = !instalada && (eventoGuardado !== null || esIOS());

  const instalar = async () => {
    if (eventoGuardado) {
      await eventoGuardado.prompt();
      await eventoGuardado.userChoice;
      eventoGuardado = null;
      avisar();
    } else if (esIOS()) {
      alert('Para instalar MultiCatálogo: toca "Compartir" y luego "Agregar a inicio".');
    }
  };

  return { puedeInstalar, instalar };
}
