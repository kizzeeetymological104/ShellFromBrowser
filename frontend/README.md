# ShellFromBrowser Frontend

Modern React + TypeScript frontend for ShellFromBrowser web terminal.

## Tech Stack

- **React 18** - UI framework
- **TypeScript** - Type safety
- **Vite** - Build tool & dev server
- **xterm.js** - Terminal emulator
- **WebSocket** - Real-time communication with backend

## Features (Phase 1)

- ✅ Login UI (mock authentication for MVP)
- ✅ xterm.js terminal emulator with fit addon
- ✅ WebSocket connection with reconnect logic
- ✅ Terminal resize support
- ✅ Status indicator (connecting/connected/disconnected/error)
- ✅ Clean, modern UI with dark theme

## Development

### Prerequisites

- Node.js v20.18+ (check: `node --version`)
- npm v10+ (check: `npm --version`)

### Setup

```bash
# Install dependencies
npm install

# Start dev server (port 3000)
npm run dev

# Build for production
npm run build

# Preview production build
npm run preview

# Lint code
npm run lint
```

### Dev Server

The dev server runs on `http://localhost:3000` with HMR (Hot Module Replacement).

WebSocket proxy is configured to forward:
- `/api` → `http://localhost:8080` (REST API)
- `/ws` → `ws://localhost:8080` (WebSocket)

**To test end-to-end:**
1. Start backend: `cd ../backend && go run ./cmd/gateway`
2. Start frontend: `npm run dev`
3. Open browser: `http://localhost:3000`

## Project Structure

```
src/
├── components/
│   ├── Login.tsx        # Authentication UI
│   ├── Login.css
│   ├── Terminal.tsx     # xterm.js terminal component
│   └── Terminal.css
├── App.tsx              # Main app component
├── App.css
├── main.tsx             # Entry point
└── index.css            # Global styles
```

## Phase 1 MVP Limitations

**Mock Authentication:**
- Login UI accepts any username/password
- No backend validation yet
- Session token is client-side only

**WebSocket:**
- Connection to `ws://localhost:8080/ws`
- Backend must be running (echo mode stub OK for Phase 1 Week 1-2)

## Roadmap

### Phase 1 Week 3 (Current)
- [x] Setup Vite + React + TypeScript
- [x] xterm.js integration
- [x] WebSocket client with reconnect
- [x] Login UI (mock)
- [x] Status indicators
- [ ] Connect to real backend JWT auth
- [ ] Test with real Docker container PTY

### Phase 1 Week 4
- [ ] Session management (multiple tabs)
- [ ] Theme switcher (light/dark)
- [ ] Command history
- [ ] Session recording UI

### Phase 1 Week 5
- [ ] MFA integration
- [ ] GDPR compliance UI (export/delete data)
- [ ] RBAC UI (admin panel)

### Phase 2
- [ ] Multi-tab support (concurrent sessions)
- [ ] Customizable themes
- [ ] Persistent command history
- [ ] Session recording playback

### Phase 3
- [ ] OAuth2 SSO integration
- [ ] Advanced RBAC UI
- [ ] Full GDPR compliance

## Security Notes

**CSP (Content Security Policy):**
- Vite dev server has no CSP by default
- Production build should include strict CSP headers (HAProxy config)

**XSS Protection:**
- React escapes by default (no dangerouslySetInnerHTML used)
- xterm.js sanitizes ANSI escape sequences

**CORS:**
- Dev server proxies API requests to avoid CORS issues
- Production uses HAProxy for same-origin enforcement

## Performance

**Bundle Size (Phase 1):**
- React + ReactDOM: ~130KB gzipped
- xterm.js + addons: ~80KB gzipped
- App code: ~20KB gzipped
- **Total: ~230KB gzipped**

**Lighthouse Scores (Target):**
- Performance: 90+
- Accessibility: 90+
- Best Practices: 95+
- SEO: 90+

## Testing

```bash
# Unit tests (TODO Phase 1 Week 4)
npm run test

# E2E tests (TODO Phase 1 Week 5)
npm run test:e2e
```

## License

MIT (or project license)
