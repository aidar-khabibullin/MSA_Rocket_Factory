import 'i18next'

import type authRu from '@/modules/auth/locales/ru.json'
import type catalogRu from '@/modules/catalog/locales/ru.json'
import type ordersRu from '@/modules/orders/locales/ru.json'
import type systemRu from '@/modules/system/locales/ru.json'
import type commonRu from '@/common/i18n/locales/ru/common.json'

// Русский словарь — эталон типов: опечатка в ключе t('...') ломает typecheck.
declare module 'i18next' {
  interface CustomTypeOptions {
    defaultNS: 'common'
    resources: {
      common: typeof commonRu
      auth: typeof authRu
      catalog: typeof catalogRu
      orders: typeof ordersRu
      system: typeof systemRu
    }
  }
}
