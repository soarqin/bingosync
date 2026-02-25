import { defineStore } from 'pinia';
import { ref, computed } from 'vue';
import { locales, localeNames, type LocaleCode } from '../locales';

export const useLocaleStore = defineStore('locale', () => {
  const locale = ref<LocaleCode>('zh-CN');

  // Load locale from localStorage
  function loadLocale() {
    const saved = localStorage.getItem('bingosync-locale') as LocaleCode | null;
    if (saved && locales[saved]) {
      locale.value = saved;
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
