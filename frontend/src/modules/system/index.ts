import './locales'

// Публичный API модуля: снаружи импортируем только отсюда, а не из внутренних папок.
export { Footer } from './components/footer'
export { StatusBanner } from './components/status-banner'
export type { ServiceStatus } from './model/service-status'
