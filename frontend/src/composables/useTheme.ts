import { ref } from 'vue'

export type Theme = 'light' | 'dark'

const STORAGE_KEY = 'gonference-theme'

function systemTheme(): Theme {
  return window.matchMedia('(prefers-color-scheme: dark)').matches ? 'dark' : 'light'
}

function storedTheme(): Theme | null {
  const value = localStorage.getItem(STORAGE_KEY)
  return value === 'light' || value === 'dark' ? value : null
}

function apply(value: Theme): void {
  document.documentElement.dataset.theme = value
}

// Singleton state so every component observes the same theme.
const theme = ref<Theme>(storedTheme() ?? systemTheme())
apply(theme.value)

// Follow the OS theme until the user makes an explicit choice.
window.matchMedia('(prefers-color-scheme: dark)').addEventListener('change', (e) => {
  if (storedTheme() === null) {
    theme.value = e.matches ? 'dark' : 'light'
    apply(theme.value)
  }
})

export function useTheme() {
  function setTheme(value: Theme): void {
    theme.value = value
    localStorage.setItem(STORAGE_KEY, value)
    apply(value)
  }

  function toggleTheme(): void {
    setTheme(theme.value === 'dark' ? 'light' : 'dark')
  }

  return { theme, setTheme, toggleTheme }
}
