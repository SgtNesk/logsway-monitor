import { useApi } from '../hooks/useApi'
import { StatusBadge, Spinner, ErrorBox } from '../components/ui'
import { Link } from 'react-router-dom'
import EventLog from '../components/EventLog'

const THRESHOLDS = {
  cpu_percent:  { warn: 70, crit: 85, label: 'CPU',  unit: '%' },
  ram_percent:  { warn: 75, crit: 90, label: 'RAM',  unit: '%' },
  disk_percent: { warn: 80, crit: 90, label: 'Disk', unit: '%' },
}

function getProblem(host) {
  const problems = []
  const m = host.last_metrics ?? {}

  if (host.status === 'offline') {
    return [{ metric: 'Connectivity', value: null, severity: 'offline', label: 'Host offline' }]
  }

  for (const [key, thr] of Object.entries(THRESHOLDS)) {
    const v = m[key]
    if (v == null) continue
    if (v >= thr.crit) {
      problems.push({ metric: thr.label, value: v, unit: thr.unit, severity: 'critical', label: `${thr.label} critical` })
    } else if (v >= thr.warn) {
      problems.push({ metric: thr.label, value: v, unit: thr.unit, severity: 'warning', label: `${thr.label} high` })
    }
  }
  return problems
}

const SEVERITY_ORDER = { critical: 0, offline: 1, warning: 2 }

export default function ProblemsPage() {
  const { data: hosts, loading, error } = useApi('/api/v1/hosts')
  const { data: eventsData } = useApi('/api/v1/events?hours=24', 15000)

  if (loading) return <Spinner />
  if (error) return <ErrorBox message={error} />

  // Collect all problematic hosts
  const rows = []
  for (const host of hosts ?? []) {
    if (host.status === 'healthy') continue
    const problems = getProblem(host)
    for (const p of problems) {
      rows.push({ host, problem: p })
    }
  }

  rows.sort((a, b) =>
    (SEVERITY_ORDER[a.problem.severity] ?? 3) - (SEVERITY_ORDER[b.problem.severity] ?? 3)
  )

  return (
    <div className="space-y-8">
      {/* ── Sezione problemi attuali ── */}
      <div className="space-y-4">
        <div>
          <h1 className="text-2xl font-semibold text-gray-900">Problems</h1>
          <p className="text-sm text-gray-500 mt-1">
            Hosts and services that require attention
          </p>
        </div>

      {rows.length === 0 ? (
        <div className="card flex flex-col items-center py-16 text-center">
          <div className="w-12 h-12 rounded-full bg-green-50 flex items-center justify-center mb-4">
            <svg width="24" height="24" viewBox="0 0 24 24" fill="none" stroke="#22c55e" strokeWidth="2.5">
              <polyline points="20 6 9 17 4 12" />
            </svg>
          </div>
          <p className="font-semibold text-gray-900">All systems operational</p>
          <p className="text-sm text-gray-500 mt-1">No hosts with warnings or critical issues</p>
        </div>
      ) : (
        <div className="card p-0 overflow-hidden">
          <table className="w-full text-sm">
            <thead>
              <tr className="text-xs text-gray-500 border-b border-gray-100 bg-gray-50">
                <th className="text-left px-5 py-3 font-medium">Status</th>
                <th className="text-left px-5 py-3 font-medium">Host</th>
                <th className="text-left px-5 py-3 font-medium">Problem</th>
                <th className="text-left px-5 py-3 font-medium">Value</th>
                <th className="text-left px-5 py-3 font-medium hidden md:table-cell">Since</th>
              </tr>
            </thead>
            <tbody>
              {rows.map(({ host, problem }, i) => (
                <tr
                  key={`${host.hostname}-${problem.metric}-${i}`}
                  className={`hover:bg-gray-50 transition-colors ${
                    i < rows.length - 1 ? 'border-b border-gray-50' : ''
                  }`}
                >
                  <td className="px-5 py-3.5">
                    <StatusBadge status={problem.severity} />
                  </td>
                  <td className="px-5 py-3.5">
                    <Link
                      to={`/hosts/${host.hostname}`}
                      className="font-medium text-gray-900 hover:text-gray-600"
                    >
                      {host.hostname}
                    </Link>
                  </td>
                  <td className="px-5 py-3.5 text-gray-700">{problem.label}</td>
                  <td className="px-5 py-3.5 font-mono text-gray-700">
                    {problem.value != null
                      ? `${problem.value.toFixed(1)}${problem.unit}`
                      : '—'}
                  </td>
                  <td className="px-5 py-3.5 hidden md:table-cell text-gray-400 text-xs">
                    {formatRelative(host.last_seen)}
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      )}
      </div>

      {/* ── Event Log ── */}
      <div className="space-y-4">
        <div>
          <h2 className="text-lg font-semibold text-gray-900">Recent Events</h2>
          <p className="text-sm text-gray-500 mt-1">Cambi di stato nelle ultime 24 ore</p>
        </div>
        <div className="card">
          <EventLog events={eventsData?.events ?? []} />
        </div>
      </div>
    </div>
  )
}

function formatRelative(ts) {
  if (!ts) return '—'
  const diff = Math.floor((Date.now() - new Date(ts).getTime()) / 1000)
  if (diff < 60) return `${diff}s ago`
  if (diff < 3600) return `${Math.floor(diff / 60)}m ago`
  if (diff < 86400) return `${Math.floor(diff / 3600)}h ago`
  return `${Math.floor(diff / 86400)}d ago`
}
