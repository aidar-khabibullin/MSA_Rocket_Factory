import { render, screen } from '@testing-library/react'

import { AsyncGate } from './async-gate'

const gate = (props: { data: string | undefined; isLoading: boolean }) => (
  <AsyncGate {...props} loader={<p>загрузка</p>} fallback={<p>ошибка</p>}>
    {(value) => <p>данные: {value}</p>}
  </AsyncGate>
)

describe('AsyncGate', () => {
  it('во время загрузки показывает загрузчик, даже если есть устаревшие данные', () => {
    render(gate({ data: 'x', isLoading: true }))
    expect(screen.getByText('загрузка')).toBeInTheDocument()
  })

  it('без данных после загрузки показывает запасной вариант', () => {
    render(gate({ data: undefined, isLoading: false }))
    expect(screen.getByText('ошибка')).toBeInTheDocument()
  })

  it('с данными отдаёт их потребителю', () => {
    render(gate({ data: 'x', isLoading: false }))
    expect(screen.getByText('данные: x')).toBeInTheDocument()
  })
})
