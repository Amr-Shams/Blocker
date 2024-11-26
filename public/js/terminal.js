import { syntaxHighlight, getRandomColor } from './helpers.js';

export default class TerminalDebugger {
    constructor() {
        this.term = new Terminal({
            cursorBlink: true,
            rows: 5,
            theme: {
                background: '#1e1e1e',
                padding: '0',
            },
            allowTransparency: true,
            scrollback: 1000,
            fontSize: 14,
            fontFamily: 'Menlo, Monaco, "Courier New", monospace',
            rendererType: 'canvas',
            convertEol: true,
            mouseEvents: true,
            clipboard: true,
        });
        this.socketNumber = window.location.port;
        this.socket = new WebSocket(`ws://localhost:${this.socketNumber}`);
        this.historyIndex = -1;
        this.commandHistory = [];
        this.inputBuffer = "";
        this.isProcessingKey = false;

        this.initialize();

    }

    initialize() {
        const terminal = document.getElementById("terminal");
        terminal.innerHTML = '';
        terminal.style.display = 'flex';

        this.term.open(terminal);
        this.term.element.style.padding = '0';
        this.term.element.style.margin = '0';

        if (window.FitAddon) {
            this.fitAddon = new window.FitAddon.FitAddon();
            this.term.loadAddon(this.fitAddon);
            this.fitAddon.fit();
            window.addEventListener('resize', () => {
                this.fitAddon.fit();
            });
        }
        this.term.onKey(({ key, domEvent }) => {
            if (this.isProcessingKey) return;
            this.isProcessingKey = true;
            this.handleInputData(key, domEvent);
            setTimeout(() => {
                this.isProcessingKey = false;
            }, 10);
        });

        this.socket.onmessage = (event) => this.handleIncomingMessage(event);
        this.socket.onopen = () => this.handleOnOpen();
        this.socket.onclose = () => this.handleOnClose();
        this.socket.onerror = (error) => this.handleOnError(error);
    }

    getPrompt() {
        const color = getRandomColor();
        return `${color}${this.socketNumber}$ \x1b[0m`;
    }

    handleInputData(key, domEvent) {
        if (domEvent && domEvent.key) {
            switch (domEvent.key) {
                case 'Enter':
                    if (this.inputBuffer === '') {
                        this.term.write('\r\n' + this.getPrompt());
                        return;
                    }else if(this.inputBuffer === 'clear'){
                        this.term.clear();
                        this.term.write('\r\n');
                        this.commandHistory.push(this.inputBuffer);
                        this.historyIndex = this.commandHistory.length;
                        this.term.write(this.getPrompt());
                        this.inputBuffer = '';
                        return;
                    }
                    const message = JSON.stringify({ type: 'command', payload: this.inputBuffer });
                    this.socket.send(message);
                    this.commandHistory.push(this.inputBuffer);
                    this.historyIndex = this.commandHistory.length;
                    this.inputBuffer = '';
                    this.term.write('\r\n');
                    return;

                case 'Backspace':
                    if (this.inputBuffer.length > 0) {
                        this.inputBuffer = this.inputBuffer.slice(0, -1);
                        this.term.write('\b \b');
                    }
                    return;

                case 'ArrowUp':
                    if (this.historyIndex > 0) {
                        this.historyIndex--;
                        this.inputBuffer = this.commandHistory[this.historyIndex];
                        this.term.write('\r\x1b[K' + this.getPrompt() + this.inputBuffer);
                    }
                    return;

                case 'ArrowDown':
                    if (this.historyIndex < this.commandHistory.length - 1) {
                        this.historyIndex++;
                        this.inputBuffer = this.commandHistory[this.historyIndex];
                        this.term.write('\r\x1b[K' + this.getPrompt() + this.inputBuffer);
                    } else {
                        this.historyIndex = this.commandHistory.length;
                        this.inputBuffer = '';
                        this.term.write('\r\x1b[K' + this.getPrompt());
                    }
                    return;

            }
        }
        if (key >= String.fromCharCode(0x20) && key <= String.fromCharCode(0x7E)) {
            this.inputBuffer += key;
            this.term.write(key);
        }
    }

    handleIncomingMessage(event) {
        try {
            const jsonResponse = JSON.parse(event.data);
            const formattedResponse = JSON.stringify(jsonResponse, null, 2);
            const highlightedResponse = syntaxHighlight(formattedResponse);
            this.term.write('\r');
            highlightedResponse.split('\n').forEach(line => {
                this.term.write(line + '\r\n');
            });
        } catch (e) {
            this.term.write(event.data + '\r\n');
        }
        this.term.write(this.getPrompt());
    }

    handleOnOpen() {
        this.term.write('Connected to server\r\n');
        this.term.write(this.getPrompt());
    }

    handleOnClose() {
        this.term.write('\r\nConnection closed remotely\r\n');
        setTimeout(() => {
            this.term.write('Reconnecting...\r\n');
            this.socket = new WebSocket(`ws://localhost:${this.socketNumber}`);
            this.socket.onmessage = (event) => this.handleIncomingMessage(event);
            this.socket.onopen = () => this.handleOnOpen();
            this.socket.onclose = () => this.handleOnClose();
            this.socket.onerror = (error) => this.handleOnError(error);
        }, 1000);
    }
    handleOnError(error) {
        this.term.write('\r\n' + error.message + '\r\n');
        this.term.write(this.getPrompt());
    }
}
