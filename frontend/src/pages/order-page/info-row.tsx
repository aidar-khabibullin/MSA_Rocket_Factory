interface InfoRowProps {
  label: string
  value: string
  strong?: boolean
  mono?: boolean
}

export const InfoRow = ({ label, value, strong, mono }: InfoRowProps) => (
  <div className="flex flex-col gap-0.5">
    <dt className="text-muted">{label}</dt>
    <dd className={`break-all ${strong ? 'text-lg font-bold tabular-nums' : ''} ${mono ? 'font-mono text-xs' : ''}`}>
      {value}
    </dd>
  </div>
)
