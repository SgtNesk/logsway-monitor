import { useApi } from '../hooks/useApi'
import { StatusBadge, Spinner, ErrorBox } from '../components/ui'
import { Link } from 'react-router-dom'

function StatCard({ label, value, color }) {
  return (
    <div className="card">
      <p className="text-sm text-gray-500">{label}</p>
      <p className={`text-3xl font-semibold mt-1 ${color}`}>{value ?? '—'}</p>
    </div>
  )
}

function HostGrid({ hosts }) {
  if (!hosts?.length) return null
  return (
    <div className="card">
      <h2 className="text-sm font-medium text-gray-500 mb-4">Host Grid</h2>
      <div className="flex flex-wrap gap-3">
        {hosts.map((h) => (
          <Link
            key={h.hostname}
            to={`/hosts/${h.hostname}`}
            title={`${h.hostname} — ${h.status}`}
          >
            <div
              className={`w-7 h-7 rounded border-2 transition-transform hover:scale-110 ${
                h.status === 'healthy'
                  ? 'bg-healthy border-green-300'
                  : h.status === 'warning'
                  ? 'bg-warning border-yellow-300'
                  : h.status === 'critical'
                  ? 'bg-critical border-red-300'
                  : 'bg-offline border-gray-300'
              }`}
            />
          </Link>
        ))}
      </div>
    </div>
  )
}

export default function DashboardPage() {
  const { data: stats, loading: sl, error: se } = useApi('/api/v1/stats')
  const { data: hosts, loading: hl, error: he } = useApi('/api/v1/hosts')

  if (sl || hl) return <Spinner />
  if (se || he) return <ErrorBox message={se || he} />

  return (
    <div className="space-y-6">
      {/* Header */}
      <div>
        <h1 className="text-2xl font-semibold text-gray-900">Dashboard</h1>
        <p className="text-sm text-gray-500 mt-1">
          Infrastructure overview — auto-refreshes every 10s
        </p>
      </div>

      {/* Stats */}
      <div className="grid grid-cols-2 sm:grid-cols-4 gap-4">
        <StatCard label="Total Hosts" value={stats?.total_hosts} color="text-gray-900" />
        <StatCard label="Healthy" value={stats?.healthy_hosts} color="text-healthy" />
        <StatCard label="Warning" value={stats?.warning_hosts} color="text-warning" />
        <StatCard label="Critical / Offline" value={(stats?.critical_hosts ?? 0) + (stats?.offline_hosts ?? 0)} color="text-critical" />
      </div>

      {/* Grid view */}
      <HostGrid hosts={hosts} />

      {/* Host list */}
      <div className="card p-0 overflow-hidden">
        <div className="px-5 py-4 border-b border-gray-100">
          <h2 className="text-sm font-medium text-gray-900">All Hosts</h2>
        </div>
        {!hosts?.length ? (
          <div className="px-5 py-10 text-center text-sm text-gray-400">
            No hosts yet. Start an agent to begin collecting metrics.
          </div>
        ) : (
          <table className="w-full text-sm">
            <thead>
              <tr className="text-xs text-gray-500 border-b border-gray-100">
                <th className="text-left px-5 py-3 font-medium">Host</th>
                <th className="text-left px-5 py-3 font-medium">Status</th>
                <th className="text-left px-5 py-3 font-medium hidden sm:table-cell">CPU</th>
                <th className="text-left px-5 py-3 font-medium hidden sm:table-cell">RAM</th>
                <th className="text-left px-5 py-3 font-medium hidden md:table-cell">Disk</th>
                <th className="text-left px-5 py-3 font-medium hidden md:table-cell">Last Seen</th>
              </tr>
            </thead>
            <tbody>
              {hosts.map((h, i) => (
                <tr
                  key={h.hostname}
                  className={`hover:bg-gray-50 transition-colors ${i < hosts.length - 1 ? 'border-b border-gray-50' : ''}`}
                >
                  <td className="px-5 py-3.5">
                    <Link
                      to={`/hosts/${h.hostname}`}
                      className="font-medium text-gray-900 hover:text-gray-600"
                    >
                      {h.hostname}
                    </Link>
                    {h.tags?.length > 0 && (
                      <span className="ml-2 text-xs text-gray-400">
                        {h.tags.join(', ')}
                      </span>
                    )}
                  </td>
                  <td className="px-5 py-3.5">
                    <StatusBadge status={h.status} />
                  </td>
                  <td className="px-5 py-3.5 hidden sm:table-cell text-gray-600">
                    {h.last_metrics?.cpu_percent != null
                      ? `${h.last_metrics.cpu_percent.toFixed(1)}%`
                      : '—'}
                  </td>
                  <td className="px-5 py-3.5 hidden sm:table-cell text-gray-600">
                    {h.last_metrics?.ram_percent != null
                      ? `${h.last_metrics.ram_percent.toFixed(1)}%`
                      : '—'}
                  </td>
                  <td className="px-5 py-3.5 hidden md:table-cell text-gray-600">
                    {h.last_metrics?.disk_percent != null
                      ? `${h.last_metrics.disk_percent.toFixed(1)}%`
                      : '—'}
                  </td>
                  <td className="px-5 py-3.5 hidden md:table-cell text-gray-400 text-xs">
                    {formatRelative(h.last_seen)}
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        )}
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
