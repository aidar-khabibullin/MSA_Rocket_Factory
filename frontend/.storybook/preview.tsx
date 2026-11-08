import type { Decorator, Preview } from '@storybook/react-vite'

import { i18n } from '@/common/i18n' // язык определяется до первой отрисовки
import '@/modules/auth/locales'
import '@/modules/catalog/locales'
import '@/modules/orders/locales'
import '@/modules/system/locales'
import '@/global/index.css'

/** Тему приложения ставит класс на <html>; переключатель на панели Storybook делает то же самое. */
const withTheme: Decorator = (Story, { globals }) => {
  const theme = globals.theme === 'light' ? 'light' : 'dark'
  const root = document.documentElement
  root.classList.remove('light', 'dark')
  root.classList.add(theme)
  root.setAttribute('data-theme', theme)

  return (
    <div className="bg-background p-6 text-foreground">
      <Story />
    </div>
  )
}

/** Язык интерфейса: в Storybook он не берётся из браузера, а задаётся переключателем (по умолчанию русский). */
const withLanguage: Decorator = (Story, { globals }) => {
  const language = globals.language === 'en' ? 'en' : 'ru'
  if (i18n.language !== language) void i18n.changeLanguage(language)
  return <Story />
}

const preview: Preview = {
  decorators: [withLanguage, withTheme],
  globalTypes: {
    language: {
      description: 'Язык',
      toolbar: {
        icon: 'globe',
        items: [
          { value: 'ru', title: 'Русский' },
          { value: 'en', title: 'English' },
        ],
        dynamicTitle: true,
      },
    },
    theme: {
      description: 'Тема',
      toolbar: {
        icon: 'circlehollow',
        items: [
          { value: 'dark', title: 'Тёмная' },
          { value: 'light', title: 'Светлая' },
        ],
        dynamicTitle: true,
      },
    },
  },
  initialGlobals: { theme: 'dark', language: 'ru' },
  parameters: {
    layout: 'centered',
    controls: { matchers: { date: /Date$/i } },
  },
}

export default preview
