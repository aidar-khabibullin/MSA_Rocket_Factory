import { Alert } from '@heroui/react/alert'

export function ErrorAlert({ message }: { message: string | null | undefined }) {
  if (!message) return null
  return (
    <Alert status="danger" role="alert">
      <Alert.Indicator />
      <Alert.Content>
        <Alert.Title>Не получилось</Alert.Title>
        <Alert.Description>{message}</Alert.Description>
      </Alert.Content>
    </Alert>
  )
}
