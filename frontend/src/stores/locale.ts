import { defineStore } from 'pinia';
import { ref, computed } from 'vue';
import { locales, localeNames, type LocaleCode } from '../locales';

// Language mapping: system language prefix -> app locale code
// Add new language mappings here when adding new locales
const LANGUAGE_MAP: Record<string, LocaleCode> = {
  'zh': 'zh-CN',    // Chinese -> Simplified Chinese
  'zh-CN': 'zh-CN', // Simplified Chinese
  'zh-TW': 'zh-CN', // Traditional Chinese -> Simplified Chinese (can add zh-TW locale later)
  'zh-HK': 'zh-CN', // Hong Kong Chinese
  // 'ja': 'ja-JP',  // Japanese (add when Japanese locale is available)
  // 'ko': 'ko-KR',  // Korean (add when Korean locale is available)
  // Add more mappings as needed
};

// Default locale when no mapping is found
const DEFAULT_LOCALE: LocaleCode = 'en-US';

// Detect system language and map to supported locale
function detectSystemLocale(): LocaleCode {
  // Get browser language (e.g., 'zh-CN', 'en-US', 'ja', etc.)
  const browserLang = navigator.language || (navigator as any).userLanguage;
  
  if (!browserLang) {
    return DEFAULT_LOCALE;
  }
  
  // Try exact match first (e.g., 'zh-CN' -> 'zh-CN')
  if (LANGUAGE_MAP[browserLang]) {
    return LANGUAGE_MAP[browserLang];
  }
  
  // Try language prefix match (e.g., 'zh-TW' -> 'zh', 'en-GB' -> 'en')
  const langPrefix = browserLang.split('-')[0];
  if (LANGUAGE_MAP[langPrefix]) {
    return LANGUAGE_MAP[langPrefix];
  }
  
  // No match found, use default (English)
  return DEFAULT_LOCALE;
}

export const useLocaleStore = defineStore('locale', () => {
  const locale = ref<LocaleCode>('zh-CN');

  // Load locale from localStorage or detect from system
  function loadLocale() {
    const saved = localStorage.getItem('bingosync-locale') as LocaleCode | null;
    if (saved && locales[saved]) {
      locale.value = saved;
    } else {
      // No saved locale, detect from system language
      const detected = detectSystemLocale();
      locale.value = detected;
      // Save detected locale to localStorage
      localStorage.setItem('bingosync-locale', detected);
    }
  }

  // Set locale
  function setLocale(code: LocaleCode) {
    if (locales[code]) {
      locale.value = code;
      localStorage.setItem('bingosync-locale', code);
    }
  }

  // Get translations
  const messages = computed(() => locales[locale.value]);

  // Translation function
  function t(path: string): string {
    const keys = path.split('.');
    let result: unknown = messages.value;
    
    for (const key of keys) {
      if (result && typeof result === 'object' && key in result) {
        result = (result as Record<string, unknown>)[key];
      } else {
        return path; // Return original path as fallback
      }
    }
    
    return typeof result === 'string' ? result : path;
  }

  // Get current locale display name
  const localeName = computed(() => localeNames[locale.value]);

  // Available locales list
  const availableLocales = Object.entries(localeNames).map(([code, name]) => ({
    code: code as LocaleCode,
    name,
  }));

  return {
    locale,
    localeName,
    availableLocales,
    loadLocale,
    setLocale,
    t,
  };
});
