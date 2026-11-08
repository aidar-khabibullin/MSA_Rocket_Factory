import Check from '@gravity-ui/icons/Check'
import Xmark from '@gravity-ui/icons/Xmark'

import type { OrderStatus } from '@/shared/api/types'

import { PROGRESS_STEPS, reachedStep } from './status'

export function OrderStepper({ status }: { status: OrderStatus }) {
  const reached = reachedStep(status)

  if (reached === null) {
    return (
      <output className="flex items-center gap-3 rounded-xl border border-danger/40 bg-danger/10 p-4 text-danger">
        <Xmark className="size-5" aria-hidden="true" />
        <span className="font-medium">Заказ отменён</span>
      </output>
    )
  }

  const finished = reached === PROGRESS_STEPS.length - 1

  return (
    <ol className="flex" aria-label="Ход выполнения заказа">
      {PROGRESS_STEPS.map((step, index) => {
        const complete = index <= reached
        const current = index === reached && !finished
        const isLast = index === PROGRESS_STEPS.length - 1

        return (
          <li
            key={step.id}
            className="relative flex flex-1 flex-col items-center gap-2"
            aria-current={current ? 'step' : undefined}
          >
            {/* Линия идёт от центра этого кружка к центру следующего. */}
            {isLast ? null : (
              <span
                className={`absolute left-1/2 top-4 h-0.5 w-full -translate-y-1/2 rounded ${index < reached ? 'bg-accent' : 'bg-border'}`}
                aria-hidden="true"
              />
            )}
            <span
              className={`relative z-10 grid size-8 place-items-center rounded-full border-2 text-sm font-bold transition-colors ${
                complete ? 'border-accent bg-accent text-accent-foreground' : 'border-border bg-surface text-muted'
              } ${current ? 'ring-4 ring-accent/25' : ''}`}
            >
              {complete && !current ? <Check className="size-4" aria-hidden="true" /> : index + 1}
            </span>
            <span className={`text-center text-sm ${complete ? 'font-medium text-foreground' : 'text-muted'}`}>
              {step.label}
            </span>
          </li>
        )
      })}
    </ol>
  )
}
