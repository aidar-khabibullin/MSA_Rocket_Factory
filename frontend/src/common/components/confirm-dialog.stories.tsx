import type { Meta, StoryObj } from '@storybook/react-vite'
import { fn } from 'storybook/test'

import { ConfirmDialog } from './confirm-dialog'

const meta = {
  title: 'Общие/ConfirmDialog',
  component: ConfirmDialog,
  args: {
    isOpen: true,
    onOpenChange: fn(),
    onConfirm: fn(),
    title: 'Отменить заказ?',
    description: 'Заказ будет отменён, зарезервированные детали вернутся на склад.',
    confirmLabel: 'Да, отменить',
    cancelLabel: 'Оставить',
  },
  parameters: { layout: 'fullscreen' },
} satisfies Meta<typeof ConfirmDialog>

export default meta
type Story = StoryObj<typeof meta>

export const Danger: Story = {}

export const Warning: Story = { args: { status: 'warning' } }

export const Pending: Story = { args: { isPending: true } }
