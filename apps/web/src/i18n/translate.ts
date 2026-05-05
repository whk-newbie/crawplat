import { defaultLocale, fallbackLocale, isSupportedLocale, type Locale, messages } from './messages'

type TranslateParams = Record<string, string | number>

function lookup(path: string, locale: Locale): string | undefined {
  const parts = path.split('.')
  let current: unknown = messages[locale]
  for (const part of parts) {
    if (!current || typeof current !== 'object' || !(part in current)) {
      return undefined
    }
    current = (current as Record<string, unknown>)[part]
  }
  return typeof current === 'string' ? current : undefined
}

function interpolate(template: string, params: TranslateParams): string {
  return template.replace(/\{(\w+)\}/g, (_, key: string) => String(params[key] ?? `{${key}}`))
}

export function normalizeLocale(value: string | null | undefined): Locale {
  return isSupportedLocale(value) ? value : defaultLocale
}

export function translate(path: string, locale: Locale, params: TranslateParams = {}): string {
  const template = lookup(path, locale) ?? lookup(path, fallbackLocale) ?? path
  return interpolate(template, params)
}
