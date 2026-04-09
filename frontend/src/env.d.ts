/// <reference types="vite/client" />

declare global {
  interface Window {
    __notify?: (text: string, color?: string, icon?: string) => void
  }
}

export {}
