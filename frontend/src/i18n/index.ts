import { createI18n } from 'vue-i18n'
import en from './en.json'
import zh from './zh.json'

// Detect system language
function detectLocale(): string {
  const nav = (globalThis as any).navigator
  if (!nav) return 'zh'
  const lang = (nav.language || nav.userLanguage || 'zh').toLowerCase()
  if (lang.startsWith('zh')) return 'zh'
  return 'en'
}

// Load saved preference or detect
function getSavedLocale(): string {
  try {
    const saved = localStorage.getItem('hey-nanobot-locale')
    if (saved && (saved === 'zh' || saved === 'en')) return saved
  } catch {}
  return detectLocale()
}

const detectedLocale = getSavedLocale()

const i18n = createI18n({
  legacy: false,
  locale: detectedLocale,
  fallbackLocale: 'en',
  messages: { en, zh },
})

export default i18n

export function setLocale(locale: string) {
  i18n.global.locale.value = locale as any
  try { localStorage.setItem('hey-nanobot-locale', locale) } catch {}
}

export function getLocale(): string {
  return i18n.global.locale.value as string
}

export function toggleLocale() {
  const next = getLocale() === 'zh' ? 'en' : 'zh'
  setLocale(next)
  return next
}
