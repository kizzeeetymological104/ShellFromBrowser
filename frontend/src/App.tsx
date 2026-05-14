import { useState } from 'react'
import Terminal from './components/Terminal'
import Login from './components/Login'
import './App.css'

function App() {
  const [isAuthenticated, setIsAuthenticated] = useState(false)
  const [sessionToken, setSessionToken] = useState<string | null>(null)

  const handleLogin = (token: string) => {
    setSessionToken(token)
    setIsAuthenticated(true)
  }

  const handleLogout = () => {
    setSessionToken(null)
    setIsAuthenticated(false)
    // Clear session cookie
    document.cookie = 'session_token=; Path=/; Expires=Thu, 01 Jan 1970 00:00:01 GMT;'
  }

  return (
    <div className="app">
      {!isAuthenticated ? (
        <Login onLogin={handleLogin} />
      ) : (
        <div className="terminal-container">
          <div className="header">
            <h1>ShellFromBrowser</h1>
            <button onClick={handleLogout} className="logout-btn">
              Logout
            </button>
          </div>
          <Terminal sessionToken={sessionToken} />
        </div>
      )}
    </div>
  )
}

export default App
