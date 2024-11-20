
// API Configuration
const API_CONFIG = {
    baseUrl: `http://${API_PORT}`,
    endpoints: {
        pairs: "/pairs",
        info: "/info",
        status: "/status",
        send: "/send",
        wallets: "/wallets",
        addWallet: "/AddWallet",
        mempool: "/mempool",
        balances: "/balance",
    },
};
const stateName = {
    0: "Idle",
    1: "Initialized",
    2: "Running",
    3: "Connected",
    4: "Syncing",
};
class BlockchainDebugger {
    constructor() {
        this.state = {
            nodes: [],
            wallets: [],
            mempool: [],
            blocks: "",
            status: "",
            type: "full",
            selectedNode: null,
            selectedWallet: null,
            balances: {},
            Address: 0,
        };
        this.mempoolHistory = [];
        this.transactionHistory = [];
        this.renderedGraph = null;
    }

    async initialize() {
        await this.refreshData();
        this.setupEventListeners();
        this.startPolling();
    }

    async refreshData() {
        try {
            const [info, balances] = await Promise.all([
                this.fetchEndpoint("info"),
                this.fetchEndpoint("balances"),
            ]);

            const newState = {
                status: stateName[info.Status],
                nodes: [
                    ...Object.values(info.WalletPeers || {}),
                    ...Object.values(info.FullPeers || {}),
                ],
                wallets: Object.values(info.Wallets.Wallets || {}),
                blocks: info.Blockchain,
                type: info.Type,
                mempool: Object.values(info.Mempool || {}),
                balances: balances,
                address: info.Address.split(":")[1],
            };
            this.updateDisplay(newState);
            this.logMessage("Data refreshed successfully");
        } catch (error) {
            this.logError("Data refresh failed:", error);
        }
    }

    updateDisplay(newState) {
        if (JSON.stringify(this.state.mempool) !== JSON.stringify(newState.mempool)) {
            this.state.mempool = newState.mempool;
            this.updateMempool();
            this.logMessage("Mempool updated");
        }
        if (this.state.status !== newState.status) {
            this.state.status = newState.status;
            this.logMessage("Status updated");
        }
        if (JSON.stringify(this.state.wallets) !== JSON.stringify(newState.wallets)) {
            this.state.wallets = newState.wallets;
            this.logMessage("Wallets updated");
        }
        if (JSON.stringify(this.state.blocks) !== JSON.stringify(newState.blocks)) {
            this.state.blocks = newState.blocks;
            this.updateBlockchain();
            this.logMessage("Blockchain updated");
        }
        if (JSON.stringify(this.state.balances) !== JSON.stringify(newState.balances)) {
            this.state.balances = newState.balances;
            this.updateBalances();
            this.logMessage("Balances updated");
        }
        if (JSON.stringify(this.state.nodes) !== JSON.stringify(newState.nodes) || this.state.Address !== newState.address) {
            this.state.nodes = newState.nodes;
            this.state.Address = newState.address;
            this.updateNetworkGraph();
            this.updateNodeStatus();
            this.logMessage("Network graph updated");
        }
    }
    async fetchEndpoint(endpoint) {
        const response = await fetch(
            `${API_CONFIG.baseUrl}${API_CONFIG.endpoints[endpoint]}`
        );
        if (!response.ok)
            throw new Error(`API Error: ${response.statusText}`);
        return response.json();
    }



    updateMempool() {
        const mempoolView = document.getElementById("mempool-view");
        const transactions = this.state.mempool
            .map((tx) => {
                return `
                        <div class="flex space-x-2 text-xs">
                            <span class="text-gray-500">0x${tx.hash.substring(
                    0,
                    8
                )}</span>
                            <span class="text-green-400">${tx.from.substring(
                    0,
                    8
                )}</span>
                            <span class="text-yellow-400">→</span>
                            <span class="text-green-400">${tx.to.substring(
                    0,
                    8
                )}</span>
                            <span class="text-blue-400">${tx.amount} BTC</span>
                        </div>
                    `;
            })
            .join("");
        mempoolView.innerHTML =
            transactions ||
            '<div class="text-gray-500">No pending transactions</div>';
    }

    updateNodeStatus() {
        const nodesView = document.getElementById("nodes-view");
        const nodesHtml = this.state.nodes
            .map((node, index) => {
                const walletPeersLength = node.WalletPeers
                    ? Object.keys(node.WalletPeers).length
                    : 0;
                const fullPeersLength = node.FullPeers
                    ? Object.keys(node.FullPeers).length
                    : 0;
                const peersLength = walletPeersLength + fullPeersLength;
                const lasHash = node.Blockchain
                    ? node.Blockchain.LastHash
                        ? node.Blockchain.LastHash.substring(0, 16)
                        : ""
                    : "";

                return `
        <div class="mb-2 p-2 bg-gray-800 rounded">
          <div class="flex justify-between items-center">
            <span class="text-xs text-gray-400">Node ${node.Address}</span>
            <span class="register-value">${stateName[node.Status]}</span>
          </div>
          <div class="hex-display text-xs mt-1">
            <span class="text-gray-500">Hash:</span>
            <span class="text-green-400">${lasHash}</span>
          </div>
          <div class="text-xs mt-1">
            <span class="text-gray-500">Peers:</span>
            <span class="text-blue-400">${peersLength}</span>
          </div>
        </div>
      `;
            })
            .join("");
        nodesView.innerHTML = nodesHtml;
    }

    updateBlockchain() {
        const blockView = document.getElementById("block-view");
        const blocksHtml = `
                        <div class="flex items-center space-x-2 mb-1">
                          <span class="text-xs text-gray-500">${this.state.blocks.Height.toString(
            16
        ).padStart(8, "0")}</span>
                          <span class="hex-display text-xs text-green-400">${this.state.blocks.LastHash.substring(
            0,
            16
        )}</span>
                          <span class="text-xs text-blue-400">${new Date(
            this.state.blocks.LastTimeUpdate / 1e6
        ).toLocaleString()} </span>
                          <span class="text-xs text-yellow-400">${this.state.type.toUpperCase()} Node</span>
  </div>


  `;

        blockView.innerHTML = blocksHtml;
    }
    updateBalances() {
        const balancesView = document.getElementById("balances-view");
        const balancesHtml = Object.entries(this.state.balances)
            .map(([address, balance]) => {
                return `
                <div class="flex space-x-2 text-xs">
                  <span class="text-green-400">${address}</span>
                  <span class="text-blue-400">${balance} BTC</span>
                </div>
              `;
            })
            .join("");
        balancesView.innerHTML =
            balancesHtml ||
            '<div class="text-gray-500">No balances available</div>';
    }

    updateNetworkGraph() {
        // Dispose of the previous graph if it exists
        this.renderedGraph?.dispose();

        // Initialize a new graph
        var graph = Viva.Graph.graph();

        // Add the main node (address) to the graph
        graph.addNode(this.state.Address, {
            label: `Node ${this.state.Address}`,
            data: this.state.nodes,
        });

        this.state.nodes.forEach((node) => {
            const address = node.Address.split(":")[1];
            graph.addNode(address, {
                label: `Node ${address}`,
            });
            graph.addLink(this.state.Address, address);
        });
        for (let i = 0; i < this.state.nodes.length; i++) {
            for (let j = i + 1; j < this.state.nodes.length; j++) {
                graph.addLink(
                    this.state.nodes[i].Address.split(":")[1],
                    this.state.nodes[j].Address.split(":")[1]
                );
            }
        }
        var renderer = Viva.Graph.View.renderer(graph, {
            container: document.getElementById("network-graph"),
        });
        this.renderedGraph = renderer;
        this.renderedGraph.run();
    }

    logMessage(message, type = "info") {
        const logView = document.getElementById("log-view");
        const timestamp = new Date().toISOString().split("T")[1].slice(0, -1);
        const logEntry = `
                    <div class="text-xs ${type === "error" ? "text-red-400" : "text-gray-400"
            }">
                        [${timestamp}] ${message}
                    </div>
                `;
        logView.innerHTML = logEntry + logView.innerHTML;
    }

    logError(message, error) {
        console.error(message, error);
        this.logMessage(`${message} ${error.message}`, "error");
    }

        // startPolling() {
        //     setInterval(() => this.refreshData(), 2000);
        // }

    setupEventListeners() {
        document
            .getElementById("transaction-form")
            .addEventListener("submit", (e) => this.handleTransaction(e));
        document
            .getElementById("add-wallet-btn")
            .addEventListener("click", () => this.handleAddWallet());
    }
}

// Initialize debugger when DOM is loaded
document.addEventListener("DOMContentLoaded", () => {
    window.debugger = new BlockchainDebugger();
    window.debugger.initialize();
});
