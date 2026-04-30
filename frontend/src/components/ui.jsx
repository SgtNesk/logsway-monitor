export function StatusDot({ status }) {
  return <span className={`status-dot ${status}`} />
}

export function StatusBadge({ status }) {
  const labels = {
    healthy: 'Healthy',
    warning: 'Warning',
    critical: 'Critical',
    offline: 'Offline',
  }
  return (
    <span className={`badge ${status}`}>
      <StatusDot status={status} />
      {labels[status] ?? status}
    </span>
  )
}

export function MetricBar({ label, value, warn = 70, crit = 85, unit = '%' }) {
  const color =
    value >= crit ? 'bg-critical' : value >= warn ? 'bg-warning' : 'bg-healthy'

  return (
    <div>
      <div className="flex justify-between text-xs text-gray-500 mb-1">
        <span>{label}</span>
        <span className="font-medium text-gray-700">
          {value?.toFixed(1)}{unit}
        </span>
      </div>
      <div className="h-1.5 bg-gray-100 rounded-full overflow-hidden">
        <div
          className={`h-full rounded-full transition-all ${color}`}
          style={{ width: `${Math.min(value ?? 0, 100)}%` }}
        />
      </div>
    </div>
  )
}

export function Spinner() {
  return (
    <div className="flex items-center justify-center py-16">
      <div className="w-6 h-6 border-2 border-gray-200 border-t-gray-500 rounded-full animate-spin" />
    </div>
  )
}

export function ErrorBox({ message }) {
  return (
    <div className="rounded-lg border border-red-200 bg-red-50 px-4 py-3 text-sm text-red-700">
      Error: {message}
    </div>
  )
}

export function EmptyState({ icon, title, description }) {
  return (
    <div className="flex flex-col items-center justify-center py-20 text-center">
      {icon && <div className="text-4xl mb-4">{icon}</div>}
      <p className="font-medium text-gray-900">{title}</p>
      {description && <p className="text-sm text-gray-500 mt-1">{description}</p>}
    </div>
  )
}
