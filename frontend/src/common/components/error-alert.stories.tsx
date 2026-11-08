import type { Meta, StoryObj } from '@storybook/react-vite'

import { ErrorAlert } from './error-alert'

const meta = {
  title: 'Общие/ErrorAlert',
  component: ErrorAlert,
  decorators: [(Story) => <div className="w-96 max-w-full"><Story /></div>],
} satisfies Meta<typeof ErrorAlert>

export default meta
type Story = StoryObj<typeof meta>

export const Default: Story = { args: { message: 'Не удалось загрузить заказ' } }

export const LongMessage: Story = {
  args: {
    message:
      'Сервис инвентаризации временно недоступен: превышено время ожидания ответа. Повторите попытку через минуту.',
  },
}

/** Пустое сообщение не рисует ничего: так страницы показывают блок только при ошибке. */
export const Hidden: Story = { args: { message: null } }
