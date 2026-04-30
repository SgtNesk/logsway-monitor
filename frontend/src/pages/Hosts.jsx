import { useApi } from '../hooks/useApi'
import { StatusBadge, Spinner, ErrorBox } from '../components/ui'
import { Link } from 'react-router-dom'

export default function HostsPage() {
  const { data: hosts, loading, error } = useApi('/api/v1/hosts')

  if (loading) return <Spinner />
  if (error) return <ErrorBox message={error} />

  const grouped = groupByTag(hosts ?? [])

  return (
    <div className="space-y-6">
      <div>
        <h1 className="text-2xl font-semibold text-gray-900">Hosts</h1>
        <p className="text-sm text-gray-500 mt-1">
          {hosts?.length ?? 0} host{hosts?.length !== 1 ? 's' : ''} registered
        </p>
      </div>

      {!hosts?.length ? (
        <div className="card text-center py-16 text-gray-400 text-sm">
          No hosts registered yet. Start an agent to begin.
        </div>
      ) : (
        Object.entries(grouped).map(([tag, tagHosts]) => (
          <div key={tag} className="space-y-3">
            <h2 className="text-xs font-semibold uppercase tracking-wider text-gray-400">
              {tag}
            </h2>
            <div className="grid gap-3 sm:grid-cols-2 lg:grid-cols-3">
              {tagHosts.map((h) => (
                <HostCard key={h.hostname} host={h} />
              ))}
            </div>
          </div>
        ))
      )}
    </div>
  )
}

function HostCard({ host }) {
  const m = host.last_metrics ?? {}
  return (
    <Link to={`/hosts/${host.hostname}`} className="block">
      <div className="card hover:shadow-md transition-shadow cursor-pointer">
        <div className="flex items-start justify-between mb-3">
          <div>
            <p className="font-medium text-gray-900">{host.hostname}</p>
            {host.tags?.length > 0 && (
              <p className="text-xs text-gray-400 mt-0.5">{host.tags.join(' · ')}</p>
            )}
          </div>
          <StatusBadge status={host.status} />
        </div>
        <div className="grid grid-cols-3 gap-3 text-xs">
          <Metric label="CPU" value={m.cpu_percent} />
          <Metric label="RAM" value={m.ram_percent} />
          <Metric label="Disk" value={m.disk_percent} />
        </div>
      </div>
    </Link>
  )
}

function Metric({ label, value }) {
  return (
    <div className="text-center bg-gray-50 rounded-md py-2">
      <p className="text-gray-400 text-xs">{label}</p>
      <p className="font-semibold text-gray-700 mt-0.5">
        {value != null ? `${value.toFixed(0)}%` : '—'}
      </p>
    </div>
  )
}

function groupByTag(hosts) {
  const result = {}
  for (const h of hosts) {
    const tag = h.tags?.[0] ?? 'untagged'
    if (!result[tag]) result[tag] = []
    result[tag].push(h)
  }
  return result
}
