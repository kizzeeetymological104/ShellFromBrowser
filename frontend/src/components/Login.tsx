import { useState } from 'react'
import './Login.css'

interface LoginProps {
  onLogin: (token: string) => void
}

function Login({ onLogin }: LoginProps) {
  const [username, setUsername] = useState('')
  const [password, setPassword] = useState('')
  const [error, setError] = useState('')
  const [loading, setLoading] = useState(false)

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault()
    setError('')
    setLoading(true)

    try {
      // TODO Phase 1 Week 3: Replace with real API call to /api/auth/login
      // For now, mock authentication
      if (username && password) {
        // Simulate API delay
        await new Promise((resolve) => setTimeout(resolve, 500))

        // Mock JWT token (real one will come from backend)
        const mockToken = `mock-jwt-${Date.now()}`

        // Set cookie (real one will be HttpOnly from backend)
        document.cookie = `session_token=${mockToken}; Path=/; SameSite=Strict`

        onLogin(mockToken)
      } else {
        setError('Please enter username and password')
      }
    } catch (err) {
      setError('Authentication failed. Please try again.')
    } finally {
      setLoading(false)
    }
  }

  return (
    <div className="login-container">
      <div className="login-box">
        <h1>ShellFromBrowser</h1>
        <p className="subtitle">Secure Web Terminal</p>

        <form onSubmit={handleSubmit}>
          <div className="form-group">
            <label htmlFor="username">Username</label>
            <input
              id="username"
              type="text"
              value={username}
              onChange={(e) => setUsername(e.target.value)}
              placeholder="Enter your username"
              disabled={loading}
              required
            />
          </div>

          <div className="form-group">
            <label htmlFor="password">Password</label>
            <input
              id="password"
              type="password"
              value={password}
              onChange={(e) => setPassword(e.target.value)}
              placeholder="Enter your password"
              disabled={loading}
              required
            />
          </div>

          {error && <div className="error-message">{error}</div>}

          <button type="submit" disabled={loading} className="login-btn">
            {loading ? 'Logging in...' : 'Login'}
          </button>
        </form>

        <div className="warning">
          <strong>Phase 1 MVP:</strong> Mock authentication active. Real JWT auth with bcrypt backend coming soon.
        </div>
      </div>
    </div>
  )
}

export default Login
