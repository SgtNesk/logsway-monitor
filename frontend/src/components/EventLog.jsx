import { Link } from 'react-router-dom'

const STATUS_DOT = {
  ok:       'bg-healthy',
  warning:  'bg-warning',
  critical: 'bg-critical',
  unknown:  'bg-gray-400',
}

const TRANSITION_LABEL = {
  'ok→warning':        { text: 'Degraded',   cls: 'text-warning'  },
  'ok→critical':       { text: 'Critical',   cls: 'text-critical' },
  'warning→critical':  { text: 'Worsened',   cls: 'text-critical' },
  'warning→ok':        { text: 'Recovered',  cls: 'text-healthy'  },
  'critical→warning':  { text: 'Improving',  cls: 'text-warning'  },
  'critical→ok':       { text: 'Recovered',  cls: 'text-healthy'  },
  'unknown→ok':        { text: 'Online',     cls: 'text-healthy'  },
  'unknown→warning':   { text: 'Warning',    cls: 'text-warning'  },
  'unknown→critical':  { text: 'Critical',   cls: 'text-critical' },
  'critical→unknown':  { text: 'Lost',       cls: 'text-gray-500' },
  'ok→unknown':        { text: 'Lost',       cls: 'text-gray-500' },
}

export default function EventLog({ events = [] }) {
  if (events.length === 0) {
    return (
      <p className="text-sm text-gray-400 py-4 text-center">
        Nessun evento nelle ultime 24 ore.
      </p>
    )
  }

  return (
    <div className="divide-y divide-gray-50">
      {events.map((evt) => {
        const key = `${evt.from_status}→${evt.to_status}`
        const transition = TRANSITION_LABEL[key] ?? { text: `${evt.from_status}→${evt.to_status}`, cls: 'text-gray-600' }

        return (
          <div key={evt.id} className="flex items-center gap-3 py-2.5 px-1 hover:bg-gray-50 rounded">
            {/* Timestamp */}
            <span className="text-xs text-gray-400 shrink-0 w-28 tabular-nums">
              {formatTime(evt.timestamp)}
            </span>

            {/* from → to */}
            <div className="flex items-center gap-1 shrink-0">
              <span className={`w-2.5 h-2.5 rounded-full inline-block ${STATUS_DOT[evt.from_status] ?? 'bg-gray-300'}`} />
              <span className="text-gray-300 text-xs">→</span>
              <span className={`w-2.5 h-2.5 rounded-full inline-block ${STATUS_DOT[evt.to_status] ?? 'bg-gray-300'}`} />
            </div>

            {/* Transition label */}
            <span className={`text-xs font-medium shrink-0 w-20 ${transition.cls}`}>
              {transition.text}
            </span>

            {/* Host */}
            <Link
              to={`/hosts/${evt.hostname}`}
              className="text-sm font-medium text-gray-700 hover:text-gray-900 shrink-0 w-36 truncate"
            >
              {evt.hostname}
            </Link>

            {/* Service */}
            <Link
              to={`/host/${evt.hostname}/service/${evt.service}`}
              className="text-xs text-gray-500 hover:text-gray-700 bg-gray-100 px-2 py-0.5 rounded shrink-0"
            >
              {evt.service}
            </Link>

            {/* Value + message */}
            <span className="text-xs text-gray-400 truncate flex-1">
              {evt.value != null ? `${evt.value.toFixed(1)} · ` : ''}
              {evt.message}
            </span>
          </div>
        )
      })}
    </div>
  )
}

function formatTime(ts) {
  if (!ts) return '—'
  const d = new Date(ts)
  const today = new Date()
  const isToday = d.toDateString() === today.toDateString()
  if (isToday) {
    return d.toLocaleTimeString('it-IT', { hour: '2-digit', minute: '2-digit', second: '2-digit' })
  }
  return d.toLocaleDateString('it-IT', { day: '2-digit', month: '2-digit' }) +
    ' ' + d.toLocaleTimeString('it-IT', { hour: '2-digit', minute: '2-digit' })
}
