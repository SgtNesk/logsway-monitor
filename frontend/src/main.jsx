import React from 'react'
import ReactDOM from 'react-dom/client'
import { BrowserRouter } from 'react-router-dom'
import App from './App'
import './index.css'

const savedTheme = localStorage.getItem('logsway-theme')
if (savedTheme === 'dark') {
  document.documentElement.classList.add('theme-dark')
} else {
  document.documentElement.classList.remove('theme-dark')
}

ReactDOM.createRoot(document.getElementById('root')).render(
  <React.StrictMode>
    <BrowserRouter>
      <App />
    </BrowserRouter>
  </React.StrictMode>
)
