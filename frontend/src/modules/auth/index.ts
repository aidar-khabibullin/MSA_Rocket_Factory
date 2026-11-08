import './locales'

// Публичный API модуля: снаружи импортируем только отсюда, а не из внутренних папок.
export { AuthForm } from './components/auth-form'
export { authenticate, type AuthMode } from './services/authenticate'
export {
  getSession,
  signIn,
  signOut,
  useRequiredSession,
  useSession,
  useSessionStore,
} from './store/session.store'
