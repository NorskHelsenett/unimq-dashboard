export function Sparkline({ data, className }: { data: number[]; className?: string }) {
  if (data.length < 2) return <svg width="100%" height={20} />
  const w = 228, h = 20, pad = 2
  const max = Math.max(...data, 1)
  const points = data
    .map((v, i) => {
      const x = pad + (i / (data.length - 1)) * (w - pad * 2)
      const y = h - pad - (v / max) * (h - pad * 2)
      return `${x.toFixed(1)},${y.toFixed(1)}`
    })
    .join(' ')
  const isFlat = Math.max(...data) === Math.min(...data)
  return (
    <svg width="100%" height={h} viewBox={`0 0 ${w} ${h}`} display="block" className={className}>
      <polyline
        points={points}
        fill="none"
        stroke={isFlat ? '#FDBA74' : '#f97316'}
        strokeWidth={1.5}
        strokeLinejoin="round"
      />
    </svg>
  )
}