import { NavLink, Link } from 'react-router-dom'

function ActivityIcon() {
  return (
    <svg className="logo-wave" width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2.5" strokeLinecap="round" strokeLinejoin="round">
      <polyline points="22 12 18 12 15 21 9 3 6 12 2 12" />
    </svg>
  )
}

export default function Layout({ children }) {
  const linkClass = ({ isActive }) =>
    `px-3 py-1.5 rounded-md text-sm font-medium transition-colors ${
      isActive
        ? 'bg-gray-100 text-gray-900'
        : 'text-gray-500 hover:text-gray-900 hover:bg-gray-50'
    }`

  return (
    <div className="min-h-screen flex flex-col">
      {/* Topbar */}
      <header className="bg-white border-b border-gray-200 sticky top-0 z-10">
        <div className="max-w-7xl mx-auto px-4 sm:px-6">
          <div className="flex items-center justify-between h-14">
            {/* Logo */}
            <Link to="/" className="flex items-center gap-2 group" title="Go to dashboard">
              <span className="text-healthy group-hover:scale-105 transition-transform">
                <ActivityIcon />
              </span>
              <span className="font-semibold text-gray-900 text-base tracking-tight group-hover:text-gray-700 transition-colors">
                LOGSWAY
              </span>
            </Link>

            {/* Nav */}
            <nav className="flex items-center gap-1">
              <NavLink to="/" end className={linkClass}>
                Dashboard
              </NavLink>
              <NavLink to="/matrix" className={linkClass}>
                Matrix
              </NavLink>
              <NavLink to="/nongreen" className={linkClass}>
                Problems
              </NavLink>
              <NavLink to="/hosts" className={linkClass}>
                Hosts
              </NavLink>
              <NavLink to="/settings" className={linkClass}>
                Settings
              </NavLink>
            </nav>

            {/* Live indicator */}
            <div className="flex items-center gap-1.5 text-xs text-gray-400">
              <span className="w-1.5 h-1.5 rounded-full bg-healthy animate-pulse" />
              Live
            </div>
          </div>
        </div>
      </header>

      {/* Content */}
      <main className="flex-1 max-w-7xl mx-auto w-full px-4 sm:px-6 py-8">
        {children}
      </main>

      <footer className="border-t border-gray-100 py-4 text-center text-xs text-gray-400">
        Logsway — Monitoring you understand in 5 minutes
      </footer>
    </div>
  )
}
