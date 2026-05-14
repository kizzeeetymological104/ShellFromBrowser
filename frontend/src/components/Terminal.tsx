import { useEffect, useRef, useState } from 'react'
import { Terminal as XTerm } from '@xterm/xterm'
import { FitAddon } from '@xterm/addon-fit'
import { WebLinksAddon } from '@xterm/addon-web-links'
import '@xterm/xterm/css/xterm.css'
import './Terminal.css'

interface TerminalProps {
  sessionToken: string | null
}

function Terminal({ sessionToken }: TerminalProps) {
  const terminalRef = useRef<HTMLDivElement>(null)
  const xtermRef = useRef<XTerm | null>(null)
  const wsRef = useRef<WebSocket | null>(null)
  const fitAddonRef = useRef<FitAddon | null>(null)

  const [status, setStatus] = useState<'connecting' | 'connected' | 'disconnected' | 'error'>('disconnected')
  const [errorMessage, setErrorMessage] = useState<string>('')

  useEffect(() => {
    if (!terminalRef.current || !sessionToken) return

    // Initialize xterm.js
    const xterm = new XTerm({
      cursorBlink: true,
      fontSize: 14,
      fontFamily: 'Menlo, Monaco, "Courier New", monospace',
      theme: {
        background: '#1e1e1e',
        foreground: '#d4d4d4',
        cursor: '#d4d4d4',
        selectionBackground: 'rgba(255, 255, 255, 0.3)',
      },
      rows: 30,
      cols: 100,
    })

    const fitAddon = new FitAddon()
    const webLinksAddon = new WebLinksAddon()

    xterm.loadAddon(fitAddon)
    xterm.loadAddon(webLinksAddon)
    xterm.open(terminalRef.current)

    fitAddon.fit()
    xtermRef.current = xterm
    fitAddonRef.current = fitAddon

    // WebSocket connection
    const connectWebSocket = () => {
      setStatus('connecting')
      xterm.writeln('Connecting to shell...')

      // TODO Phase 1 Week 3: Use real WebSocket URL with session token
      // For now, mock connection
      const ws = new WebSocket('ws://localhost:8080/ws')

      ws.onopen = () => {
        setStatus('connected')
        xterm.writeln('\x1b[32mConnected!\x1b[0m')
        xterm.writeln('Phase 1 MVP: Mock mode - echo server active')
        xterm.writeln('Type commands to see them echoed back.')
        xterm.writeln('')
        xterm.write('$ ')
      }

      ws.onmessage = (event) => {
        xterm.write(event.data)
      }

      ws.onerror = (error) => {
        console.error('WebSocket error:', error)
        setStatus('error')
        setErrorMessage('WebSocket connection failed')
        xterm.writeln('\x1b[31mError: Connection failed\x1b[0m')
      }

      ws.onclose = () => {
        setStatus('disconnected')
        xterm.writeln('\x1b[33mConnection closed\x1b[0m')
      }

      wsRef.current = ws

      // Send data from terminal to WebSocket
      xterm.onData((data) => {
        if (ws.readyState === WebSocket.OPEN) {
          ws.send(data)
        }
      })

      // Handle terminal resize
      xterm.onResize(({ rows, cols }) => {
        if (ws.readyState === WebSocket.OPEN) {
          // TODO: Send resize event to backend
          ws.send(JSON.stringify({ type: 'resize', rows, cols }))
        }
      })
    }

    connectWebSocket()

    // Handle window resize
    const handleResize = () => {
      if (fitAddonRef.current) {
        fitAddonRef.current.fit()
      }
    }

    window.addEventListener('resize', handleResize)

    // Cleanup
    return () => {
      window.removeEventListener('resize', handleResize)
      if (wsRef.current) {
        wsRef.current.close()
      }
      if (xtermRef.current) {
        xtermRef.current.dispose()
      }
    }
  }, [sessionToken])

  const handleReconnect = () => {
    if (wsRef.current) {
      wsRef.current.close()
    }
    // Trigger re-render to reconnect
    setStatus('connecting')
  }

  return (
    <div className="terminal-wrapper">
      <div className="status-bar">
        <div className="status-indicator">
          <span className={`status-dot status-${status}`}></span>
          <span className="status-text">
            {status === 'connecting' && 'Connecting...'}
            {status === 'connected' && 'Connected'}
            {status === 'disconnected' && 'Disconnected'}
            {status === 'error' && `Error: ${errorMessage}`}
          </span>
        </div>
        {(status === 'disconnected' || status === 'error') && (
          <button onClick={handleReconnect} className="reconnect-btn">
            Reconnect
          </button>
        )}
      </div>
      <div ref={terminalRef} className="terminal" />
    </div>
  )
}

export default Terminal
