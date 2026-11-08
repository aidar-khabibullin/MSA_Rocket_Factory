import type { Meta, StoryObj } from '@storybook/react-vite'
import { fn } from 'storybook/test'

import type { Part } from '../model/part'
import { PartCard, PartCardSkeleton } from './part-card'

const part: Part = {
  uuid: 'p1',
  name: 'Двигатель Raptor',
  description: 'Метано-кислородный ракетный двигатель замкнутого цикла.',
  price: 1_850_000,
  stockQuantity: 12,
  category: 'CATEGORY_ENGINE',
  manufacturer: 'SpaceX',
  tags: ['methalox', 'reusable'],
}

const meta = {
  title: 'Каталог/PartCard',
  component: PartCard,
  args: { part, selected: false, onToggle: fn(), index: 0 },
  decorators: [(Story) => <div className="w-80"><Story /></div>],
} satisfies Meta<typeof PartCard>

export default meta
type Story = StoryObj<typeof meta>

export const Default: Story = {}

export const Selected: Story = { args: { selected: true } }

export const LowStock: Story = { args: { part: { ...part, stockQuantity: 3 } } }

export const OutOfStock: Story = { args: { part: { ...part, stockQuantity: 0 } } }

export const Skeleton: Story = {
  render: () => <PartCardSkeleton />,
  decorators: [(Story) => <div className="w-80"><Story /></div>],
}
