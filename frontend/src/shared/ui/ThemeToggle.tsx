import Moon from '@gravity-ui/icons/Moon'
import Sun from '@gravity-ui/icons/Sun'
import { Button } from '@heroui/react/button'

import { useTheme } from '@/shared/lib/theme'

export function ThemeToggle() {
  const { theme, toggle } = useTheme()
  const isDark = theme === 'dark'

  return (
    <Button
      isIconOnly
      size="sm"
      variant="ghost"
      onPress={toggle}
      aria-label={isDark ? 'Включить светлую тему' : 'Включить тёмную тему'}
    >
      {isDark ? <Sun /> : <Moon />}
    </Button>
  )
}
