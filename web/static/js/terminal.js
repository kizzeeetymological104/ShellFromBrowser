(function () {
    "use strict";

    var terminals = {};
    var activeTermId = null;
    var tabIndex = 0;

    function createTerminal(sessionId) {
        var id = "term-" + (tabIndex++);
        var term = new Terminal({
            cursorBlink: true,
            fontSize: 14,
            fontFamily: "'JetBrains Mono', 'Fira Code', monospace",
            theme: {
                background: "#1e1e2e",
                foreground: "#cdd6f4",
                cursor: "#f5e0dc",
            },
        });

        var fitAddon = new FitAddon.FitAddon();
        var webLinksAddon = new WebLinksAddon.WebLinksAddon();
        term.loadAddon(fitAddon);
        term.loadAddon(webLinksAddon);

        var protocol = window.location.protocol === "https:" ? "wss:" : "ws:";
        var wsUrl = protocol + "//" + window.location.host + "/ws?token=" +
            encodeURIComponent(SessionManager.getToken());
        if (sessionId) wsUrl += "&session=" + encodeURIComponent(sessionId);

        var ws = new WebSocket(wsUrl);
        ws.binaryType = "arraybuffer";

        ws.onopen = function () {
            var dims = JSON.stringify({ type: "resize", cols: term.cols, rows: term.rows });
            ws.send(dims);
            setStatus("Connected");
        };

        ws.onmessage = function (event) {
            if (event.data instanceof ArrayBuffer) {
                term.write(new Uint8Array(event.data));
            } else {
                try {
                    var msg = JSON.parse(event.data);
                    if (msg.type === "session") {
                        terminals[id].sessionId = msg.id;
                        SessionManager.setActive(msg.id);
                        updateSessionInfo();
                    }
                } catch (e) {
                    term.write(event.data);
                }
            }
        };

        ws.onclose = function () { setStatus("Disconnected"); };

        term.onData(function (data) {
            if (ws.readyState === WebSocket.OPEN) ws.send(data);
        });

        term.onResize(function (size) {
            if (ws.readyState === WebSocket.OPEN) {
                ws.send(JSON.stringify({ type: "resize", cols: size.cols, rows: size.rows }));
            }
        });

        terminals[id] = { term: term, fitAddon: fitAddon, ws: ws, sessionId: sessionId || null };
        addTab(id);
        switchTab(id);

        return id;
    }

    function addTab(id) {
        var tabs = document.getElementById("tabs");
        var tab = document.createElement("div");
        tab.className = "tab";
        tab.dataset.id = id;
        tab.textContent = "Session " + (Object.keys(terminals).length);

        var closeBtn = document.createElement("span");
        closeBtn.className = "tab-close";
        closeBtn.textContent = "×";
        closeBtn.onclick = function (e) {
            e.stopPropagation();
            closeTab(id);
        };

        tab.appendChild(closeBtn);
        tab.onclick = function () { switchTab(id); };
        tabs.appendChild(tab);
    }

    function switchTab(id) {
        if (activeTermId && terminals[activeTermId]) {
            terminals[activeTermId].term.element.style.display = "none";
        }
        activeTermId = id;
        var t = terminals[id];
        var container = document.getElementById("terminal");

        if (!t.term.element) {
            t.term.open(container);
        } else {
            t.term.element.style.display = "";
            container.appendChild(t.term.element);
        }
        t.fitAddon.fit();
        t.term.focus();

        document.querySelectorAll(".tab").forEach(function (el) {
            el.classList.toggle("active", el.dataset.id === id);
        });

        if (t.sessionId) SessionManager.setActive(t.sessionId);
        updateSessionInfo();
    }

    function closeTab(id) {
        var t = terminals[id];
        if (t) {
            t.ws.close();
            t.term.dispose();
            delete terminals[id];
        }
        var tabEl = document.querySelector('.tab[data-id="' + id + '"]');
        if (tabEl) tabEl.remove();

        var remaining = Object.keys(terminals);
        if (remaining.length > 0) {
            switchTab(remaining[remaining.length - 1]);
        }
    }

    function setStatus(text) {
        document.getElementById("connection-status").textContent = text;
    }

    function updateSessionInfo() {
        var info = document.getElementById("session-info");
        var count = Object.keys(terminals).length;
        info.textContent = count + " session" + (count !== 1 ? "s" : "");
    }

    function init() {
        var token = SessionManager.getToken();
        if (!token) {
            document.getElementById("login-screen").style.display = "flex";
        } else {
            startApp();
        }

        document.getElementById("login-form").onsubmit = async function (e) {
            e.preventDefault();
            var user = document.getElementById("username").value;
            var pass = document.getElementById("password").value;
            try {
                await SessionManager.login(user, pass);
                document.getElementById("login-screen").style.display = "none";
                startApp();
            } catch (err) {
                document.getElementById("login-error").textContent = err.message;
            }
        };
    }

    function startApp() {
        document.getElementById("app").style.display = "flex";
        createTerminal();

        document.getElementById("new-tab").onclick = function () {
            createTerminal();
        };

        window.addEventListener("resize", function () {
            if (activeTermId && terminals[activeTermId]) {
                terminals[activeTermId].fitAddon.fit();
            }
        });
    }

    // Check if auth is required
    fetch("/api/login", { method: "POST", headers: { "Content-Type": "application/json" }, body: "{}" })
        .then(function (r) { return r.json(); })
        .then(function (data) {
            if (data.token === "no-auth") {
                SessionManager.login("", "").then(function () { init(); });
            } else {
                init();
            }
        })
        .catch(function () { init(); });
})();
