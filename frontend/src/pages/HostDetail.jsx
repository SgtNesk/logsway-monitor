import { useParams } from 'react-router-dom'
import { useState } from 'react'
import {
  LineChart, Line, XAxis, YAxis, Tooltip, ResponsiveContainer, CartesianGrid,
} from 'recharts'
import { useApi } from '../hooks/useApi'
import { StatusBadge, MetricBar, Spinner, ErrorBox } from '../components/ui'

const HOUR_OPTIONS = [1, 6, 24, 72]

export default function HostDetailPage() {
  const { hostname } = useParams()
  const [hours, setHours] = useState(1)

  const { data: host, loading: hl, error: he } = useApi(`/api/v1/hosts/${hostname}`)
  const { data: raw, loading: ml, error: me } = useApi(
    `/api/v1/hosts/${hostname}/metrics?hours=${hours}`,
    10000
  )

  if (hl || ml) return <Spinner />
  if (he || me) return <ErrorBox message={he || me} />
  if (!host) return <ErrorBox message="Host not found" />

  const m = host.last_metrics ?? {}
  const series = buildSeries(raw ?? [])

  return (
    <div className="space-y-6">
      {/* Header */}
      <div className="flex items-start justify-between">
        <div>
          <div className="flex items-center gap-3">
            <h1 className="text-2xl font-semibold text-gray-900">{hostname}</h1>
            <StatusBadge status={host.status} />
          </div>
          {host.tags?.length > 0 && (
            <p className="text-sm text-gray-400 mt-1">{host.tags.join(' · ')}</p>
          )}
        </div>
        <p className="text-xs text-gray-400 mt-1">
          Last seen: {formatRelative(host.last_seen)}
        </p>
      </div>

      {/* Current metrics */}
      <div className="card">
        <h2 className="text-sm font-medium text-gray-900 mb-4">Current State</h2>
        <div className="space-y-4">
          <MetricBar label="CPU" value={m.cpu_percent} warn={70} crit={85} />
          <MetricBar label="RAM" value={m.ram_percent} warn={75} crit={90} />
          <MetricBar label="Disk" value={m.disk_percent} warn={80} crit={90} />
        </div>
        <div className="grid grid-cols-3 gap-4 mt-5 pt-5 border-t border-gray-100 text-sm">
          <InfoItem label="RAM Used" value={m.ram_used_gb != null ? `${m.ram_used_gb.toFixed(1)} GB` : '—'} />
          <InfoItem label="Load (1m)" value={m.load_1?.toFixed(2) ?? '—'} />
          <InfoItem label="Load (5m)" value={m.load_5?.toFixed(2) ?? '—'} />
        </div>
      </div>

      {/* Charts */}
      <div className="card">
        <div className="flex items-center justify-between mb-5">
          <h2 className="text-sm font-medium text-gray-900">History</h2>
          <div className="flex gap-1">
            {HOUR_OPTIONS.map((h) => (
              <button
                key={h}
                onClick={() => setHours(h)}
                className={`px-2.5 py-1 text-xs rounded-md font-medium transition-colors ${
                  hours === h
                    ? 'bg-gray-900 text-white'
                    : 'bg-gray-100 text-gray-500 hover:bg-gray-200'
                }`}
              >
                {h}h
              </button>
            ))}
          </div>
        </div>

        <div className="space-y-6">
          <Chart data={series.cpu} dataKey="value" label="CPU %" color="#22c55e" />
          <Chart data={series.ram} dataKey="value" label="RAM %" color="#3b82f6" />
          <Chart data={series.disk} dataKey="value" label="Disk %" color="#f59e0b" />
        </div>
      </div>
    </div>
  )
}

function Chart({ data, dataKey, label, color }) {
  return (
    <div>
      <p className="text-xs text-gray-500 mb-2">{label}</p>
      <ResponsiveContainer width="100%" height={120}>
        <LineChart data={data} margin={{ top: 2, right: 4, bottom: 2, left: -20 }}>
          <CartesianGrid strokeDasharray="3 3" stroke="#f1f5f9" />
          <XAxis dataKey="time" tick={{ fontSize: 10, fill: '#9ca3af' }} tickLine={false} />
          <YAxis domain={[0, 100]} tick={{ fontSize: 10, fill: '#9ca3af' }} tickLine={false} />
          <Tooltip
            contentStyle={{ fontSize: 12, border: '1px solid #e5e7eb', borderRadius: 6 }}
            formatter={(v) => [`${v.toFixed(1)}%`, label]}
          />
          <Line
            type="monotone"
            dataKey={dataKey}
            stroke={color}
            strokeWidth={1.5}
            dot={false}
            isAnimationActive={false}
          />
        </LineChart>
      </ResponsiveContainer>
    </div>
  )
}

function InfoItem({ label, value }) {
  return (
    <div>
      <p className="text-xs text-gray-400">{label}</p>
      <p className="font-medium text-gray-700 mt-0.5">{value}</p>
    </div>
  )
}

function buildSeries(points) {
  const series = { cpu: [], ram: [], disk: [] }
  const map = { cpu_percent: 'cpu', ram_percent: 'ram', disk_percent: 'disk' }

  for (const p of points) {
    const key = map[p.name]
    if (!key) continue
    series[key].push({
      time: formatTime(p.timestamp),
      value: p.value,
    })
  }
  return series
}

function formatTime(ts) {
  const d = new Date(ts)
  return d.toLocaleTimeString('en-US', { hour: '2-digit', minute: '2-digit' })
}

function formatRelative(ts) {
  if (!ts) return '—'
  const diff = Math.floor((Date.now() - new Date(ts).getTime()) / 1000)
  if (diff < 60) return `${diff}s ago`
  if (diff < 3600) return `${Math.floor(diff / 60)}m ago`
  return `${Math.floor(diff / 3600)}h ago`
}
