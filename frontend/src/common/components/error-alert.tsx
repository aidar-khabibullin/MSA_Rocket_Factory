import { Alert } from '@heroui/react/alert'
import { useTranslation } from 'react-i18next'

export const ErrorAlert = ({ message }: { message: string | null | undefined }) => {
  const { t } = useTranslation()
  if (!message) return null
  return (
    <Alert status="danger" role="alert">
      <Alert.Indicator />
      <Alert.Content>
        <Alert.Title>{t('errors.title')}</Alert.Title>
        <Alert.Description>{message}</Alert.Description>
      </Alert.Content>
    </Alert>
  )
}
