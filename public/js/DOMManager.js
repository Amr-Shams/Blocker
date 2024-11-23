export class DOMManager {
    constructor() {
        this.elements = {};
        this.previousState = {};
        this.initializeDOMCache();
    }

    initializeDOMCache() {
        const elementIds = [
            'transaction-view',
            'nodes-view',
            'block-view',
            'balances-view',
            'log-view',
            'network-graph'
        ];
        
        elementIds.forEach(id => {
            this.elements[id] = document.getElementById(id);
        });
    }

    createElement(type, props = {}, ...children) {
        const element = document.createElement(type);
        
        Object.entries(props).forEach(([key, value]) => {
            if (key === 'className') {
                element.className = value;
            } else if (key === 'style') {
                Object.assign(element.style, value);
            } else if (key.startsWith('data-')) {
                element.setAttribute(key, value);
            } else {
                element[key] = value;
            }
        });

        children.flat().forEach(child => {
            if (typeof child === 'string') {
                element.appendChild(document.createTextNode(child));
            } else if (child instanceof Node) {
                element.appendChild(child);
            }
        });

        return element;
    }

    hasChanged(newObj, oldObj, path = '') {
        if (newObj === oldObj) return false;
        if (typeof newObj !== typeof oldObj) return true;
        if (typeof newObj !== 'object') return true;
        if (!newObj || !oldObj) return true;

        const keys = new Set([...Object.keys(newObj), ...Object.keys(oldObj)]);
        return Array.from(keys).some(key => {
            const newPath = path ? `${path}.${key}` : key;
            return this.hasChanged(newObj[key], oldObj[key], newPath);
        });
    }

    updateElement(parent, newNode, oldNode, index = 0) {
        if (!oldNode) {
            parent.appendChild(newNode);
        } else if (!newNode) {
            parent.removeChild(parent.childNodes[index]);
        } else if (this.hasChanged(newNode, oldNode)) {
            parent.replaceChild(newNode, parent.childNodes[index]);
        } else {
            const newLength = newNode.childNodes.length;
            const oldLength = oldNode.childNodes.length;
            for (let i = 0; i < newLength || i < oldLength; i++) {
                this.updateElement(
                    oldNode,
                    newNode.childNodes[i],
                    oldNode.childNodes[i],
                    i
                );
            }
        }
    }

    createNodeElement(node, stateName) {
        const peersLength = 
            (node.WalletPeers ? Object.keys(node.WalletPeers).length : 0) +
            (node.FullPeers ? Object.keys(node.FullPeers).length : 0);
        
        return this.createElement('div',
            { 
                className: 'mb-2 p-2 bg-gray-800 rounded',
                'data-node-id': node.Address 
            },
            this.createElement('div',
                { className: 'flex justify-between items-center' },
                this.createElement('span',
                    { className: 'text-xs text-gray-400' },
                    `Node ${node.Address}`
                ),
                this.createElement('span',
                    { className: 'register-value' },
                    stateName[node.Status]
                )
            ),
            this.createElement('div',
                { className: 'hex-display text-xs mt-1' },
                this.createElement('span',
                    { className: 'text-gray-500' },
                    'Hash: '
                ),
                this.createElement('span',
                    { className: 'text-green-400' },
                    node.Blockchain?.LastHash?.substring(0, 16) || ''
                )
            ),
            this.createElement('div',
                { className: 'text-xs mt-1' },
                this.createElement('span',
                    { className: 'text-gray-500' },
                    'Peers: '
                ),
                this.createElement('span',
                    { className: 'text-blue-400' },
                    peersLength.toString()
                )
            )
        );
    }

    updateTransactionHistory(transaction, maxHistory = 50) {
        const container = this.elements['transaction-view'];
        const transactionElement = this.createElement('div',
            { className: 'flex space-x-2 text-xs' },
            this.createElement('span', { className: 'text-green-400' }, transaction.from),
            this.createElement('span', { className: 'text-blue-400' }, '->'),
            this.createElement('span', { className: 'text-green-400' }, transaction.to),
            this.createElement('span', { className: 'text-blue-400' }, `${transaction.amount} BTC`)
        );
        
        container.insertBefore(transactionElement, container.firstChild);
        
        while (container.children.length > maxHistory) {
            container.removeChild(container.lastChild);
        }
    }

    updateNodeList(nodes, previousNodes, stateName) {
        if (!this.hasChanged(nodes, previousNodes)) return;

        const container = this.elements['nodes-view'];
        const fragment = document.createDocumentFragment();

        nodes.forEach((node, index) => {
            const oldNode = previousNodes[index];
            if (!oldNode || this.hasChanged(node, oldNode)) {
                fragment.appendChild(this.createNodeElement(node, stateName));
            } else {
                fragment.appendChild(container.children[index].cloneNode(true));
            }
        });

        container.innerHTML = '';
        container.appendChild(fragment);
    }

    updateBalances(newBalances, previousBalances) {
        if (!this.hasChanged(newBalances, previousBalances)) return;

        const container = this.elements['balances-view'];
        const fragment = document.createDocumentFragment();

        Object.entries(newBalances).forEach(([address, balance]) => {
            fragment.appendChild(
                this.createElement('div',
                    { className: 'flex space-x-2 text-xs' },
                    this.createElement('span',
                        { className: 'text-green-400' },
                        address
                    ),
                    this.createElement('span',
                        { className: 'text-blue-400' },
                        `${balance} BTC`
                    )
                )
            );
        });

        container.innerHTML = '';
        container.appendChild(fragment);
    }

    updateBlockView(blocks, type) {
        const blockView = this.elements['block-view'];
        const blockElement = this.createElement('div',
            { className: 'flex items-center space-x-2 mb-1' },
            this.createElement('span',
                { className: 'text-xs text-gray-500' },
                blocks.Height.toString(16).padStart(8, "0")
            ),
            this.createElement('span',
                { className: 'hex-display text-xs text-green-400' },
                blocks.LastHash.substring(0, 16)
            ),
            this.createElement('span',
                { className: 'text-xs text-blue-400' },
                new Date(blocks.LastTimeUpdate / 1e6).toLocaleString()
            ),
            this.createElement('span',
                { className: 'text-xs text-yellow-400' },
                `${type.toUpperCase()} Node`
            )
        );
    
        blockView.innerHTML = '';
        blockView.appendChild(blockElement);
    }

    logMessage(message, type = 'info') {
        const logView = this.elements['log-view'];
        const timestamp = new Date().toISOString().split('T')[1].slice(0, -1);
        const logElement = this.createElement('div',
            { 
                className: `text-xs ${type === 'error' ? 'text-red-400' : 'text-gray-400'}` 
            },
            `[${timestamp}] ${message}`
        );
        
        logView.insertBefore(logElement, logView.firstChild);
    }
}