import { useApi } from '../hooks/useApi'
import { Spinner, ErrorBox } from '../components/ui'
import { Link } from 'react-router-dom'

const SERVICES = ['cpu', 'memory', 'disk', 'load', 'network', 'ping']

const STATUS_COLOR = {
  ok:      'bg-healthy',
  warning: 'bg-warning',
  critical:'bg-critical',
  unknown: 'bg-gray-300',
}

const STATUS_RING = {
  ok:      'ring-green-400',
  warning: 'ring-yellow-400',
  critical:'ring-red-400',
  unknown: 'ring-gray-300',
}

function StatusDot({ status, hostname, service }) {
  return (
    <td className="px-2 py-2 text-center">
      <Link
        to={`/host/${hostname}/service/${service}`}
        title={`${hostname} / ${service}: ${status}`}
        className={`
          inline-block w-2.5 h-2.5 rounded-full status-orb
          ${STATUS_COLOR[status] ?? 'bg-gray-300'}
          ring-2 ring-offset-1 ${STATUS_RING[status] ?? 'ring-gray-300'}
          hover:scale-125 transition-transform
        `}
      />
    </td>
  )
}

const WORST_LABEL = {
  ok:       { text: 'All OK',    cls: 'text-healthy' },
  warning:  { text: 'Warning',   cls: 'text-warning'  },
  critical: { text: 'Critical',  cls: 'text-critical' },
  unknown:  { text: 'No data',   cls: 'text-gray-400' },
}

export default function MatrixPage() {
  const { data, loading, error } = useApi('/api/v1/matrix', 15000)

  if (loading) return <Spinner />
  if (error)   return <ErrorBox message={error} />

  const { hosts = [], services = SERVICES } = data ?? {}

  return (
    <div className="space-y-6">
      <div>
        <h1 className="text-2xl font-semibold text-gray-900">Matrix</h1>
        <p className="text-sm text-gray-500 mt-1">
          Each dot represents the state of a service. Click for details.
        </p>
      </div>

      {hosts.length === 0 ? (
        <div className="card text-center py-16 text-gray-400">
          No connected hosts yet.
        </div>
      ) : (
        <div className="card p-0 overflow-x-auto">
          <table className="w-full text-sm border-collapse">
            <thead>
              <tr className="bg-gray-50 border-b border-gray-100">
                <th className="text-left px-4 py-3 font-medium text-gray-500 text-xs uppercase tracking-wide min-w-[160px]">
                  Host
                </th>
                {services.map(svc => (
                  <th key={svc} className="px-2 py-3 font-medium text-gray-500 text-xs uppercase tracking-wide text-center min-w-[60px]">
                    {svc}
                  </th>
                ))}
                <th className="px-4 py-3 font-medium text-gray-500 text-xs uppercase tracking-wide text-left">
                  Worst
                </th>
              </tr>
            </thead>
            <tbody>
              {hosts.map((host, i) => {
                const worst = WORST_LABEL[host.worst_status] ?? WORST_LABEL.unknown
                return (
                  <tr
                    key={host.hostname}
                    className={`hover:bg-gray-50 transition-colors ${
                      i < hosts.length - 1 ? 'border-b border-gray-50' : ''
                    }`}
                  >
                    <td className="px-4 py-2.5">
                      <Link
                        to={`/hosts/${host.hostname}`}
                        className="font-medium text-gray-900 hover:text-gray-600 text-sm"
                      >
                        {host.hostname}
                      </Link>
                    </td>
                    {services.map(svc => (
                      <StatusDot
                        key={svc}
                        status={host.services?.[svc]?.status ?? 'unknown'}
                        hostname={host.hostname}
                        service={svc}
                      />
                    ))}
                    <td className={`px-4 py-2.5 text-xs font-medium ${worst.cls}`}>
                      {worst.text}
                    </td>
                  </tr>
                )
              })}
            </tbody>
          </table>
        </div>
      )}

      {/* Legenda */}
      <div className="flex flex-wrap gap-4 text-xs text-gray-500">
        {[
          ['bg-healthy', 'OK'],
          ['bg-warning',  'Warning'],
          ['bg-critical', 'Critical'],
          ['bg-gray-300', 'No data'],
        ].map(([cls, label]) => (
          <span key={label} className="flex items-center gap-1.5">
            <span className={`w-3 h-3 rounded-full ${cls} inline-block`} />
            {label}
          </span>
        ))}
        <span className="text-gray-400">· Clicca su un pallino per il dettaglio</span>
      </div>
    </div>
  )
}
