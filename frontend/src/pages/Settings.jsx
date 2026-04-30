import { useState } from 'react'

const DEFAULT_SETTINGS = {
  serverUrl: 'http://localhost:8080',
  agentInterval: 30,
  cpuWarn: 70,
  cpuCrit: 85,
  ramWarn: 75,
  ramCrit: 90,
  diskWarn: 80,
  diskCrit: 90,
  retention: 7,
}

export default function SettingsPage() {
  const [s, setS] = useState(() => {
    try {
      const saved = localStorage.getItem('logsway-settings')
      return saved ? { ...DEFAULT_SETTINGS, ...JSON.parse(saved) } : DEFAULT_SETTINGS
    } catch {
      return DEFAULT_SETTINGS
    }
  })
  const [saved, setSaved] = useState(false)

  function update(key, value) {
    setS((prev) => ({ ...prev, [key]: value }))
    setSaved(false)
  }

  function save() {
    localStorage.setItem('logsway-settings', JSON.stringify(s))
    setSaved(true)
    setTimeout(() => setSaved(false), 2000)
  }

  return (
    <div className="space-y-6 max-w-2xl">
      <div>
        <h1 className="text-2xl font-semibold text-gray-900">Settings</h1>
        <p className="text-sm text-gray-500 mt-1">Configure thresholds and connection details</p>
      </div>

      {/* Connection */}
      <Section title="Server">
        <Field label="API Server URL" description="URL of the Logsway server">
          <input
            type="url"
            value={s.serverUrl}
            onChange={(e) => update('serverUrl', e.target.value)}
            className="w-full border border-gray-200 rounded-md px-3 py-1.5 text-sm focus:outline-none focus:ring-2 focus:ring-gray-300"
            placeholder="http://localhost:8080"
          />
        </Field>
      </Section>

      {/* Agent config snippet */}
      <Section title="Agent Config Example">
        <pre className="bg-gray-50 border border-gray-200 rounded-lg p-4 text-xs text-gray-700 overflow-auto">
{`server:
  url: "${s.serverUrl}"
  timeout: 10

agent:
  hostname: "my-server"
  interval: ${s.agentInterval}
  tags:
    - production`}
        </pre>
      </Section>

      {/* Thresholds */}
      <Section title="Alert Thresholds">
        <div className="grid grid-cols-2 gap-4">
          <ThresholdField label="CPU Warning (%)" value={s.cpuWarn} onChange={(v) => update('cpuWarn', v)} />
          <ThresholdField label="CPU Critical (%)" value={s.cpuCrit} onChange={(v) => update('cpuCrit', v)} />
          <ThresholdField label="RAM Warning (%)" value={s.ramWarn} onChange={(v) => update('ramWarn', v)} />
          <ThresholdField label="RAM Critical (%)" value={s.ramCrit} onChange={(v) => update('ramCrit', v)} />
          <ThresholdField label="Disk Warning (%)" value={s.diskWarn} onChange={(v) => update('diskWarn', v)} />
          <ThresholdField label="Disk Critical (%)" value={s.diskCrit} onChange={(v) => update('diskCrit', v)} />
        </div>
        <p className="text-xs text-gray-400 mt-3">
          Note: threshold changes apply to the UI display only. Server-side thresholds are configured via environment variables.
        </p>
      </Section>

      {/* Retention */}
      <Section title="Data Retention">
        <Field label="Metric retention (days)" description="How long to keep historical metrics">
          <input
            type="number"
            min={1}
            max={90}
            value={s.retention}
            onChange={(e) => update('retention', Number(e.target.value))}
            className="w-24 border border-gray-200 rounded-md px-3 py-1.5 text-sm focus:outline-none focus:ring-2 focus:ring-gray-300"
          />
        </Field>
      </Section>

      {/* Save */}
      <div className="flex items-center gap-3">
        <button onClick={save} className="btn-primary">
          Save settings
        </button>
        {saved && (
          <span className="text-sm text-healthy font-medium">Saved ✓</span>
        )}
      </div>
    </div>
  )
}

function Section({ title, children }) {
  return (
    <div className="card space-y-4">
      <h2 className="text-sm font-semibold text-gray-900 pb-2 border-b border-gray-100">{title}</h2>
      {children}
    </div>
  )
}

function Field({ label, description, children }) {
  return (
    <div>
      <label className="block text-sm font-medium text-gray-700 mb-1">{label}</label>
      {description && <p className="text-xs text-gray-400 mb-2">{description}</p>}
      {children}
    </div>
  )
}

function ThresholdField({ label, value, onChange }) {
  return (
    <div>
      <label className="block text-xs text-gray-500 mb-1">{label}</label>
      <input
        type="number"
        min={0}
        max={100}
        value={value}
        onChange={(e) => onChange(Number(e.target.value))}
        className="w-full border border-gray-200 rounded-md px-3 py-1.5 text-sm focus:outline-none focus:ring-2 focus:ring-gray-300"
      />
    </div>
  )
}
