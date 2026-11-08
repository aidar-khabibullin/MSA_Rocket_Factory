import Rocket from '@gravity-ui/icons/Rocket'

const SIZES = {
  md: { box: 'size-9', icon: 'size-5', text: 'text-lg' },
  lg: { box: 'size-12', icon: 'size-7', text: 'text-2xl' },
} as const

export function Logo({ size = 'md' }: { size?: keyof typeof SIZES }) {
  const { box, icon, text } = SIZES[size]
  return (
    <span className="flex items-center gap-3">
      <span
        className={`${box} grid place-items-center rounded-xl bg-accent text-accent-foreground shadow-[0_0_24px_-4px_var(--accent)]`}
        aria-hidden="true"
      >
        <Rocket className={icon} />
      </span>
      <span className={`font-bold tracking-tight ${text}`}>
        Rocket<span className="text-accent">Factory</span>
      </span>
    </span>
  )
}
