import Moon from '@gravity-ui/icons/Moon'
import Sun from '@gravity-ui/icons/Sun'
import { Button } from '@heroui/react/button'
import { useTranslation } from 'react-i18next'

import { useTheme } from '@/common/lib/theme'

export function ThemeToggle() {
  const { t } = useTranslation()
  const { theme, toggle } = useTheme()
  const isDark = theme === 'dark'

  return (
    <Button
      isIconOnly
      size="sm"
      variant="ghost"
      onPress={toggle}
      aria-label={isDark ? t('theme.toLight') : t('theme.toDark')}
    >
      {isDark ? <Sun /> : <Moon />}
    </Button>
  )
}
