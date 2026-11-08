import type { Meta, StoryObj } from '@storybook/react-vite'

import { OrderStepper } from './order-stepper'

const meta = {
  title: 'Заказы/OrderStepper',
  component: OrderStepper,
  parameters: { layout: 'padded' },
  decorators: [(Story) => <div className="w-[34rem] max-w-full"><Story /></div>],
} satisfies Meta<typeof OrderStepper>

export default meta
type Story = StoryObj<typeof meta>

export const PendingPayment: Story = { args: { status: 'PENDING_PAYMENT' } }

export const Paid: Story = { args: { status: 'PAID' } }

export const Assembled: Story = { args: { status: 'ASSEMBLED' } }

export const Cancelled: Story = { args: { status: 'CANCELLED' } }
