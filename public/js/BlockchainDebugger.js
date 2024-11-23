// BlockchainDebugger.js
import API_CONFIG from "./config.js";
import { stateName } from "./helpers.js";
import { DOMManager } from "./DOMManager.js";
export default class BlockchainDebugger {
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
            TSXhistory: [],
        };
        this.domManager = new DOMManager();
        this.networkGraph = null;
        this.pollInterval = 5000;
        this.timer = null;
    }

    /**
     * Initialize the blockchain debugger
     */
    async initialize() {
        try {
            await this.refreshData();
            this.setupEventListeners();
            this.startPolling();
            this.domManager.logMessage("Blockchain debugger initialized successfully");
        } catch (error) {
            this.domManager.logMessage("Initialization failed: " + error.message, "error");
        }
    }
    /**
     * Set up all event listeners
     */
    setupEventListeners() {
        const handlers = {
            'transaction-form': {
                event: 'submit',
                handler: (e) => this.handleTransaction(e)
            },
            'add-wallet-btn': {
                event: 'click',
                handler: () => this.handleAddWallet()
            },
            'step-btn': {
                event: 'click',
                handler: () => this.handleStep()
            },
            'pause-btn': {
                event: 'click',
                handler: () => this.handlePause()
            },
            'continue-btn': {
                event: 'click',
                handler: () => this.handleContinue()
            }
        };

        Object.entries(handlers).forEach(([id, { event, handler }]) => {
            const element = document.getElementById(id);
            if (element) {
                element.addEventListener(event, handler);
            } else {
                this.domManager.logMessage(`Warning: Element ${id} not found`, "error");
            }
        });
    }

    /**
     * Fetch data from an endpoint
     */
    async fetchEndpoint(endpoint) {
        try {
            const response = await fetch(
                `${API_CONFIG.baseUrl}${API_CONFIG.endpoints[endpoint]}`
            );
            
            if (!response.ok) {
                throw new Error(`API Error: ${response.statusText}`);
            }
            
            return await response.json();
        } catch (error) {
            this.domManager.logMessage(`API Error: ${error.message}`, "error");
            throw error;
        }
    }

    /**
     * Refresh blockchain data
     */
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
                balances: balances,
                address: info.Address.split(":")[1],
                TSXhistory: info.TSXhistory? info.TSXhistory : [],
            };
            console.log(newState.blocks);
            this.updateState(newState);
            this.domManager.logMessage("Data refreshed successfully");
        } catch (error) {
            this.domManager.logMessage("Failed to refresh data: " + error.message, "error");
        }
    }

    /**
     * Update state and trigger UI updates
     */
    updateState(newState) {
        const previousState = { ...this.state };
        Object.assign(this.state, newState);
        if (newState.nodes !== previousState.nodes) {
            this.domManager.updateNodeList(newState.nodes, previousState.nodes, stateName);
            this.updateNetworkGraph();
        }

        if (newState.balances !== previousState.balances) {
            this.domManager.updateBalances(newState.balances, previousState.balances);
        }
        
        if (newState.blocks !== previousState.blocks) {
            this.domManager.updateBlockView(newState.blocks, newState.type);
        }

        if (newState.status !== previousState.status) {
            this.domManager.logMessage(`Status changed to: ${newState.status}`);
        }
        if (JSON.stringify(newState.TSXhistory) !== JSON.stringify(previousState.TSXhistory)) {
            newState.TSXhistory.forEach((tsx) => {
                this.domManager.updateTransactionHistory(tsx);
            });
        }
    }

    /**
     * Handle transaction submission
     */
    async handleTransaction(event) {
        event.preventDefault();
        
        try {
            const formData = new FormData(event.target);
            const transaction = {
                from: formData.get('from'),
                to: formData.get('to'),
                amount: formData.get('amount')
            };

            const response = await fetch(
                `${API_CONFIG.baseUrl}${API_CONFIG.endpoints.send}`,
                {
                    method: "POST",
                    body: formData
                }
            );

            if (!response.ok) {
                throw new Error(`Transaction failed: ${response.statusText}`);
            }

            this.domManager.updateTransactionHistory(transaction);
            this.domManager.logMessage("Transaction sent successfully");
            await this.refreshData();
            
        } catch (error) {
            this.domManager.logMessage(`Transaction error: ${error.message}`, "error");
        }
    }

    /**
     * Handle adding a new wallet
     */
    async handleAddWallet() {
        try {
            const response = await fetch(
                `${API_CONFIG.baseUrl}${API_CONFIG.endpoints.addWallet}`,
                { method: "POST" }
            );

            if (!response.ok) {
                throw new Error(`Failed to add wallet: ${response.statusText}`);
            }

            this.domManager.logMessage("Wallet added successfully");
            await this.refreshData();
            
        } catch (error) {
            this.domManager.logMessage(`Failed to add wallet: ${error.message}`, "error");
        }
    }

    /**
     * Update network visualization
     */
    updateNetworkGraph() {
        if (this.networkGraph) {
            this.networkGraph.dispose();
        }

        const graph = Viva.Graph.graph();

        // Add main node
        graph.addNode(this.state.Address, {
            label: `Node ${this.state.Address}`,
            data: this.state.nodes,
        });

        // Add peer nodes and connections
        this.state.nodes.forEach((node) => {
            const address = node.Address.split(":")[1];
            graph.addNode(address, {
                label: `Node ${address}`,
            });
            graph.addLink(this.state.Address, address);
        });

        // Add peer-to-peer connections
        for (let i = 0; i < this.state.nodes.length; i++) {
            for (let j = i + 1; j < this.state.nodes.length; j++) {
                graph.addLink(
                    this.state.nodes[i].Address.split(":")[1],
                    this.state.nodes[j].Address.split(":")[1]
                );
            }
        }

        // Render the graph
        this.networkGraph = Viva.Graph.View.renderer(graph, {
            container: document.getElementById("network-graph"),
        });
        
        this.networkGraph.run();
    }

    /**
     * Handle stepping to next node
     */
    handleStep() {
        const currentPort = parseInt(API_CONFIG.baseUrl.split(":")[2]);
        const newPort = 5001 + ((currentPort - 5000) % 4);
        API_CONFIG.baseUrl = `http://localhost:${newPort}`;
        this.updateURL();
    }

    /**
     * Update URL after stepping
     */
    updateURL() {
        const currentUrl = window.location.href;
        const baseUrl = currentUrl.split(":").slice(0, 2).join(":");
        window.location.replace(`${baseUrl}:${API_CONFIG.baseUrl.split(":")[2]}`);
    }

    /**
     * Handle pausing updates
     */
    handlePause() {
        if (this.timer) {
            clearInterval(this.timer);
            this.timer = null;
            this.domManager.logMessage("Updates paused");
        }
    }

    /**
     * Handle continuing updates
     */
    handleContinue() {
        if (!this.timer) {
            this.startPolling();
            this.domManager.logMessage("Updates resumed");
        }
    }

    /**
     * Start polling for updates
     */
    startPolling() {
        this.timer = setInterval(() => this.refreshData(), this.pollInterval);
    }

    /**
     * Clean up resources
     */
    dispose() {
        this.handlePause();
        if (this.networkGraph) {
            this.networkGraph.dispose();
        }
        // Clean up any other resources or event listeners

    }
}