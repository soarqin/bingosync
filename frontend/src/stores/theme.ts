import { defineStore } from 'pinia';
import { ref, watch } from 'vue';

export type Theme = 'dark' | 'light';

export const useThemeStore = defineStore('theme', () => {
  const theme = ref<Theme>('dark');

  // Load theme from localStorage
  function loadTheme() {
    const saved = localStorage.getItem('bingosync-theme');
    if (saved === 'light' || saved === 'dark') {
      theme.value = saved;
    }
    applyTheme();
  }

  // Apply theme to DOM
  function applyTheme() {
    if (theme.value === 'light') {
      document.documentElement.classList.add('light');
    } else {
      document.documentElement.classList.remove('light');
    }
  }

  // Toggle theme
  function toggleTheme() {
    theme.value = theme.value === 'dark' ? 'light' : 'dark';
    applyTheme();
    localStorage.setItem('bingosync-theme', theme.value);
  }

  // Watch theme changes
  watch(theme, applyTheme);

  return {
    theme,
    loadTheme,
    toggleTheme,
  };
});
