import { Routes, Route } from 'react-router-dom'
import Layout from './components/Layout'
import DashboardPage from './pages/Dashboard'
import ProblemsPage from './pages/Problems'
import HostsPage from './pages/Hosts'
import HostDetailPage from './pages/HostDetail'
import SettingsPage from './pages/Settings'

export default function App() {
  return (
    <Layout>
      <Routes>
        <Route path="/" element={<DashboardPage />} />
        <Route path="/nongreen" element={<ProblemsPage />} />
        <Route path="/problems" element={<ProblemsPage />} />
        <Route path="/hosts" element={<HostsPage />} />
        <Route path="/hosts/:hostname" element={<HostDetailPage />} />
        <Route path="/settings" element={<SettingsPage />} />
      </Routes>
    </Layout>
  )
}
