import type { Meta, StoryObj } from '@storybook/react-vite'
import { fn } from 'storybook/test'

import { CategoryFilter } from './category-filter'

const meta = {
  title: 'Каталог/CategoryFilter',
  component: CategoryFilter,
  args: {
    value: null,
    onChange: fn(),
    total: 14,
    counts: new Map([
      ['CATEGORY_ENGINE', 5],
      ['CATEGORY_FUEL', 4],
      ['CATEGORY_PORTHOLE', 3],
      ['CATEGORY_WING', 2],
    ]),
  },
  parameters: { layout: 'padded' },
} satisfies Meta<typeof CategoryFilter>

export default meta
type Story = StoryObj<typeof meta>

export const All: Story = {}

export const EngineSelected: Story = { args: { value: 'CATEGORY_ENGINE' } }

export const Loading: Story = { args: { total: undefined, counts: new Map() } }
