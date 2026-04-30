import { useParams, Link } from 'react-router-dom'
import { useState } from 'react'
import {
  LineChart, Line, XAxis, YAxis, Tooltip, ResponsiveContainer, CartesianGrid,
  ReferenceLine,
} from 'recharts'
import { useApi } from '../hooks/useApi'
import { StatusBadge, Spinner, ErrorBox } from '../components/ui'

const HOUR_OPTIONS = [1, 6, 24, 72]

const SERVICE_LABELS = {
  cpu:     'CPU',
  memory:  'Memory',
  disk:    'Disk',
  load:    'Load',
  network: 'Network',
  ping:    'Ping',
}

const SERVICE_UNIT = {
  cpu:    '%',
  memory: '%',
  disk:   '%',
  load:   '',
  network:'',
  ping:   '',
}

export default function ServiceDetailPage() {
  const { hostname, service } = useParams()
  const [hours, setHours] = useState(24)

  const { data: detail, loading: dl, error: de } = useApi(
    `/api/v1/hosts/${hostname}/services/${service}`,
    15000
  )
  const { data: history, loading: hl, error: he } = useApi(
    `/api/v1/hosts/${hostname}/services/${service}/history?hours=${hours}`,
    15000
  )

  if (dl || hl) return <Spinner />
  if (de || he) return <ErrorBox message={de || he} />
  if (!detail)  return <ErrorBox message="Service not found" />

  const serviceLabel = SERVICE_LABELS[service] ?? service
  const unit = SERVICE_UNIT[service] ?? ''
  const points = (history?.points ?? []).map(p => ({
    time: formatTime(p.time),
    value: p.value,
    status: p.status,
  }))

  return (
    <div className="space-y-6">
      {/* Breadcrumb + header */}
      <div>
        <div className="flex items-center gap-1 text-sm text-gray-400 mb-3">
          <Link to="/matrix" className="hover:text-gray-600">Matrix</Link>
          <span>/</span>
          <Link to={`/hosts/${hostname}`} className="hover:text-gray-600">{hostname}</Link>
          <span>/</span>
          <span className="text-gray-700 font-medium">{serviceLabel}</span>
        </div>
        <div className="flex items-center gap-3">
          <h1 className="text-2xl font-semibold text-gray-900">
            {hostname} — {serviceLabel}
          </h1>
          <StatusBadge status={detail.status} />
        </div>
      </div>

      {/* Metriche correnti */}
      <div className="card">
        <h2 className="text-sm font-medium text-gray-900 mb-4">Stato corrente</h2>
        <div className="grid grid-cols-2 md:grid-cols-4 gap-6">
          <InfoItem
            label="Valore attuale"
            value={detail.current_value != null
              ? `${detail.current_value.toFixed(1)}${unit}`
              : '—'}
            highlight
          />
          {detail.thresholds && (
            <>
              <InfoItem label="Soglia warning"  value={`${detail.thresholds.warning}${unit}`}  />
              <InfoItem label="Soglia critical" value={`${detail.thresholds.critical}${unit}`} />
            </>
          )}
          {detail.last_change && (
            <InfoItem
              label="Ultimo cambio"
              value={`${detail.last_change.from} → ${detail.last_change.to}`}
              sub={formatRelative(detail.last_change.at)}
            />
          )}
        </div>
      </div>

      {/* Grafico storico */}
      {points.length > 0 && (
        <div className="card">
          <div className="flex items-center justify-between mb-5">
            <h2 className="text-sm font-medium text-gray-900">Storico</h2>
            <div className="flex gap-1">
              {HOUR_OPTIONS.map(h => (
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
          <ServiceChart
            data={points}
            unit={unit}
            warning={detail.thresholds?.warning}
            critical={detail.thresholds?.critical}
          />
        </div>
      )}

      {/* Raw output */}
      {detail.raw_output && (
        <div className="card">
          <h2 className="text-sm font-medium text-gray-900 mb-3">Raw Output</h2>
          <pre className="text-xs font-mono bg-gray-50 rounded-md p-4 overflow-x-auto text-gray-700 leading-relaxed whitespace-pre-wrap">
            {detail.raw_output}
          </pre>
        </div>
      )}
    </div>
  )
}

function ServiceChart({ data, unit, warning, critical }) {
  const domain = unit === '%' ? [0, 100] : ['auto', 'auto']
  return (
    <ResponsiveContainer width="100%" height={160}>
      <LineChart data={data} margin={{ top: 4, right: 4, bottom: 2, left: -20 }}>
        <CartesianGrid strokeDasharray="3 3" stroke="#f1f5f9" />
        <XAxis dataKey="time" tick={{ fontSize: 10, fill: '#9ca3af' }} tickLine={false} />
        <YAxis domain={domain} tick={{ fontSize: 10, fill: '#9ca3af' }} tickLine={false} />
        <Tooltip
          contentStyle={{ fontSize: 12, border: '1px solid #e5e7eb', borderRadius: 6 }}
          formatter={(v) => [`${typeof v === 'number' ? v.toFixed(2) : v}${unit}`, 'Valore']}
        />
        {warning != null && (
          <ReferenceLine y={warning} stroke="#eab308" strokeDasharray="4 3" strokeWidth={1.5}
            label={{ value: 'warn', fill: '#eab308', fontSize: 9, position: 'insideTopLeft' }} />
        )}
        {critical != null && (
          <ReferenceLine y={critical} stroke="#ef4444" strokeDasharray="4 3" strokeWidth={1.5}
            label={{ value: 'crit', fill: '#ef4444', fontSize: 9, position: 'insideTopLeft' }} />
        )}
        <Line
          type="monotone"
          dataKey="value"
          stroke="#6366f1"
          strokeWidth={1.5}
          dot={false}
          isAnimationActive={false}
        />
      </LineChart>
    </ResponsiveContainer>
  )
}

function InfoItem({ label, value, sub, highlight }) {
  return (
    <div>
      <p className="text-xs text-gray-400">{label}</p>
      <p className={`font-semibold mt-0.5 ${highlight ? 'text-lg text-gray-900' : 'text-gray-700'}`}>
        {value}
      </p>
      {sub && <p className="text-xs text-gray-400 mt-0.5">{sub}</p>}
    </div>
  )
}

function formatTime(ts) {
  const d = new Date(ts)
  return d.toLocaleTimeString('it-IT', { hour: '2-digit', minute: '2-digit' })
}

function formatRelative(ts) {
  if (!ts) return '—'
  const diff = Math.floor((Date.now() - new Date(ts).getTime()) / 1000)
  if (diff < 60)   return `${diff}s fa`
  if (diff < 3600) return `${Math.floor(diff / 60)}m fa`
  if (diff < 86400)return `${Math.floor(diff / 3600)}h fa`
  return `${Math.floor(diff / 86400)}d fa`
}
