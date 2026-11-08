import { AlertDialog } from '@heroui/react/alert-dialog'
import { Button } from '@heroui/react/button'
import { Spinner } from '@heroui/react/spinner'
import type { ReactNode } from 'react'

interface ConfirmDialogProps {
  isOpen: boolean
  onOpenChange: (isOpen: boolean) => void
  title: string
  description: ReactNode
  confirmLabel: string
  cancelLabel: string
  onConfirm: () => void
  isPending?: boolean
  /** Опасное действие красится в danger и получает соответствующую иконку. */
  status?: 'danger' | 'warning' | 'accent'
}

/**
 * Подтверждение необратимого действия. Управляется снаружи (isOpen), поэтому открывать его может
 * любой код, а не только кнопка-триггер. Escape и клик по фону закрывают окно, пока идёт запрос — нет.
 */
export function ConfirmDialog({
  isOpen,
  onOpenChange,
  title,
  description,
  confirmLabel,
  cancelLabel,
  onConfirm,
  isPending = false,
  status = 'danger',
}: ConfirmDialogProps) {
  return (
    <AlertDialog.Backdrop
      isOpen={isOpen}
      onOpenChange={(next) => {
        if (!isPending) onOpenChange(next)
      }}
      isKeyboardDismissDisabled={isPending}
    >
      <AlertDialog.Container>
        <AlertDialog.Dialog>
          <AlertDialog.Header>
            <AlertDialog.Icon status={status} />
            <AlertDialog.Heading>{title}</AlertDialog.Heading>
          </AlertDialog.Header>
          <AlertDialog.Body>{description}</AlertDialog.Body>
          <AlertDialog.Footer>
            <Button variant="tertiary" isDisabled={isPending} onPress={() => onOpenChange(false)}>
              {cancelLabel}
            </Button>
            <Button variant={status === 'danger' ? 'danger' : 'primary'} isDisabled={isPending} onPress={onConfirm}>
              {isPending ? <Spinner size="sm" color="current" /> : null}
              {confirmLabel}
            </Button>
          </AlertDialog.Footer>
        </AlertDialog.Dialog>
      </AlertDialog.Container>
    </AlertDialog.Backdrop>
  )
}
