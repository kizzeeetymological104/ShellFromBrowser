(function () {
    "use strict";

    const term = new Terminal({
        cursorBlink: true,
        fontSize: 14,
        fontFamily: "'JetBrains Mono', 'Fira Code', monospace",
        theme: {
            background: "#1e1e2e",
            foreground: "#cdd6f4",
            cursor: "#f5e0dc",
        },
    });

    const fitAddon = new FitAddon.FitAddon();
    const webLinksAddon = new WebLinksAddon.WebLinksAddon();

    term.loadAddon(fitAddon);
    term.loadAddon(webLinksAddon);
    term.open(document.getElementById("terminal"));
    fitAddon.fit();

    window.addEventListener("resize", () => fitAddon.fit());

    const protocol = window.location.protocol === "https:" ? "wss:" : "ws:";
    const wsUrl = `${protocol}//${window.location.host}/ws`;
    let ws = null;

    function connect() {
        ws = new WebSocket(wsUrl);
        ws.binaryType = "arraybuffer";

        ws.onopen = function () {
            term.writeln("\x1b[32mConnected to ShellFromBrowser\x1b[0m");
            const dims = { type: "resize", cols: term.cols, rows: term.rows };
            ws.send(JSON.stringify(dims));
        };

        ws.onmessage = function (event) {
            if (event.data instanceof ArrayBuffer) {
                term.write(new Uint8Array(event.data));
            } else {
                term.write(event.data);
            }
        };

        ws.onclose = function () {
            term.writeln("\r\n\x1b[31mDisconnected\x1b[0m");
        };

        ws.onerror = function () {
            term.writeln("\r\n\x1b[31mConnection error\x1b[0m");
        };
    }

    term.onData(function (data) {
        if (ws && ws.readyState === WebSocket.OPEN) {
            ws.send(data);
        }
    });

    term.onResize(function (size) {
        if (ws && ws.readyState === WebSocket.OPEN) {
            const dims = { type: "resize", cols: size.cols, rows: size.rows };
            ws.send(JSON.stringify(dims));
        }
    });

    connect();
})();
