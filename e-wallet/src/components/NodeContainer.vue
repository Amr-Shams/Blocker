<template>
  <div class="container-dark p-4">
    <div class="dashboard-layout">
      <!-- Left side - Original Wallet Section -->
      <div class="wallet-section">
        <div class="table-wrapper">
          <table class="table table-dark">
            <thead>
              <tr>
                <th v-for="key in nodeKeys" :key="key">{{ key }}</th>
              </tr>
            </thead>
            <tbody>
              <tr
                v-for="index in totalRows"
                :key="index"
                class="table-row"
                :class="{ filled: index <= nodes.length }"
              >
                <template v-if="nodes[index - 1]">
                  <td v-for="key in nodeKeys" :key="key">
                    {{ nodes[index - 1][key] }}
                  </td>
                </template>
                <template v-else>
                  <td v-for="key in nodeKeys" :key="key" class="empty-cell">
                    &nbsp;
                  </td>
                </template>
              </tr>
            </tbody>
          </table>
        </div>
        <div v-if="syncInProgress" class="sync-label">Syncing...</div>
        <button
          class="btn btn-primary-dark"
          @click="addNode"
          :disabled="nodes.length >= totalRows"
        >
          {{ buttonLabel }}
        </button>
        <hr />
        <div class="transfer-form">
          <form @submit.prevent="transferAmount">
            <div style="display: flex; gap: 1rem">
              <div>
                <label for="fromWallet">From Wallet:</label>
                <input
                  id="fromWallet"
                  v-model="fromWallet"
                  type="text"
                  required
                  placeholder="0axc1234..."
                />
              </div>
              <div>
                <label for="toWallet">To Wallet:</label>
                <input
                  id="toWallet"
                  v-model="toWallet"
                  type="text"
                  required
                  placeholder="0axc1234..."
                />
              </div>
              <div>
                <label for="amount">Amount:</label>
                <input id="amount" v-model.number="amount" required />
              </div>
            </div>
            <button type="submit" :disabled="transactionInProgress">
              <span v-if="transactionInProgress">Processing...</span>
              <span v-else>Transfer</span>
            </button>
          </form>
        </div>
      </div>

      <!-- Right side - New Tables -->
      <div class="right-panel">
        <!-- Recent Transactions Table -->
        <div class="table-wrapper transaction-history">
          <table class="table table-dark">
            <thead>
              <tr>
                <th v-for="key in historyKeys" :key="key">
                  {{ key }}
                </th>
              </tr>
            </thead>
            <tbody>
              <tr v-for="tx in recentTransactions" :key="tx.hash">
                <td>{{ formatTime(tx.timestamp) }}</td>
                <td>{{ truncateAddress(tx.from) }}</td>
                <td>{{ truncateAddress(tx.to) }}</td>
                <td
                  :class="{
                    credit: tx.type === 'credit',
                    debit: tx.type === 'debit',
                  }"
                >
                  {{ tx.type === "credit" ? "+" : "-" }}{{ tx.amount }}
                </td>
                <td>{{ truncateHash(tx.hash) }}</td>
                <td>
                  <span :class="{ valid: tx.valid, invalid: !tx.valid }">
                    {{ tx.valid ? "Valid" : "Invalid" }}
                  </span>
                </td>
              </tr>
            </tbody>
          </table>
        </div>
        <hr />
        <div class="table-wrapper wallet-status">
          <table class="table table-dark">
            <thead>
              <tr>
                <th>Wallet Address</th>
                <th>Last Hash</th>
                <th>Last Update</th>
              </tr>
            </thead>
            <tbody>
              <tr v-for="wallet in walletStatus" :key="wallet.address">
                <td>{{ truncateAddress(wallet.address) }}</td>
                <td>{{ truncateHash(wallet.lastHash) }}</td>
                <td>{{ formatTime(wallet.lastUpdate) }}</td>
              </tr>
            </tbody>
          </table>
        </div>
      </div>
    </div>
  </div>
</template>

<script>
export default {
  name: "NodeContainer",
  data() {
    return {
      nodes: [],
      totalRows: 10,
      buttonLabel: "Add Wallet",
      buttonTimeout: null,
      fromWallet: "",
      toWallet: "",
      amount: 0,
      transactionInProgress: false,
      recentTransactions: [],
      walletStatus: [],
    };
  },
  computed: {
    nodeKeys() {
      if (this.nodes.length === 0) return ["Address", "Type", "Status"];
      return Object.keys(this.nodes[0]).filter((key) => key !== "id");
    },
    historyKeys() {
      return ["Time", "From", "To", "Amount", "Hash", "Status"];
    },
  },
  methods: {
    addNode() {
      if (this.nodes.length < this.totalRows) {
        const newNode = {
          id: Date.now().toString(),
          Address: `0x${Math.random().toString(16).substr(2, 40)}`,
          Type: `Type${Math.floor(Math.random() * 5) + 1}`, // Random type for demonstration
          Status: "Idle",
        };
        this.nodes.push(newNode);
        this.updateWalletStatus(newNode.Address);
        this.buttonLabel = `${this.nodes.length}/${this.totalRows}`;

        if (this.buttonTimeout) clearTimeout(this.buttonTimeout);
        this.buttonTimeout = setTimeout(() => {
          this.buttonLabel = "Add Wallet";
        }, 2000);

        this.syncInProgress = true;
        setTimeout(() => {
          this.syncInProgress = false;
        }, 3000); // Sync label visible for 3 seconds
      }
    },
    async transferAmount() {
      this.transactionInProgress = true;
      const fromWalletIndex = this.nodes.findIndex(
        (node) => node.Address === this.fromWallet
      );
      const toWalletIndex = this.nodes.findIndex(
        (node) => node.Address === this.toWallet
      );

      if (fromWalletIndex === -1 || toWalletIndex === -1) {
        alert("Invalid wallet address");
        this.transactionInProgress = false;
        return;
      }

      if (this.nodes[fromWalletIndex].Balance < this.amount) {
        alert("Insufficient balance");
        this.transactionInProgress = false;
        return;
      }

      // Process transaction
      this.nodes[fromWalletIndex].Balance -= this.amount;
      this.nodes[toWalletIndex].Balance += this.amount;

      // Create transaction records
      const txHash = `0x${Math.random().toString(16).substr(2, 64)}`;
      const timestamp = new Date();

      // Add debit transaction
      this.recentTransactions.unshift({
        timestamp,
        from: this.fromWallet,
        to: this.toWallet,
        amount: this.amount,
        hash: txHash,
        valid: true,
        type: "debit",
      });

      // Add credit transaction
      this.recentTransactions.unshift({
        timestamp,
        from: this.fromWallet,
        to: this.toWallet,
        amount: this.amount,
        hash: txHash,
        valid: true,
        type: "credit",
      });

      // Update wallet status for both wallets involved in the transaction
      this.updateWalletStatus(this.fromWallet);
      this.updateWalletStatus(this.toWallet);

      // Update the latest hash for all wallets
      this.nodes.forEach((node) => {
        this.updateWalletStatus(node.Address);
      });

      // Limit recent transactions to last 10
      if (this.recentTransactions.length > 10) {
        this.recentTransactions = this.recentTransactions.slice(0, 10);
      }

      this.fromWallet = "";
      this.toWallet = "";
      this.amount = 0;
      await new Promise((resolve) => setTimeout(resolve, 2000));
      this.transactionInProgress = false;
    },
    updateWalletStatus(address) {
      const status = {
        address,
        lastHash: `0x${Math.random().toString(16).substr(2, 64)}`,
        lastUpdate: new Date(),
      };

      const existingIndex = this.walletStatus.findIndex(
        (w) => w.address === address
      );
      if (existingIndex >= 0) {
        this.walletStatus[existingIndex] = status;
      } else {
        this.walletStatus.push(status);
      }
    },
    formatTime(timestamp) {
      return new Date(timestamp).toLocaleString();
    },
    truncateAddress(address) {
      return `${address.substr(0, 6)}...${address.substr(-4)}`;
    },
    truncateHash(hash) {
      return `${hash.substr(0, 6)}...${hash.substr(-4)}`;
    },
  },
  beforeUnmount() {
    if (this.buttonTimeout) {
      clearTimeout(this.buttonTimeout);
    }
  },
};
</script>

<style scoped>
.container-dark {
  background-color: #243642;
  color: #a9b1d6;
  min-height: 100vh;
  display: flex;
  align-items: flex-start;
  font-family: "Courier New", Courier, monospace;
}

.dashboard-layout {
  display: flex;
  gap: 2rem;
  width: 100%;
}

.wallet-section {
  width: 50%;
}

.right-panel {
  width: 50%;
  display: flex;
  flex-direction: column;
  gap: 2rem;
}

.table-wrapper {
  background-color: transparent;
  border-radius: 4px;
  overflow: hidden;
  margin-bottom: 1rem;
}

.transaction-history {
  height: 30vh;
  overflow-y: auto;
  background-color: transparent;
}

.wallet-status {
  height: 54vh;
  overflow-y: auto;
  background-color: transparent;
}

.table-dark {
  width: 100%;
  background-color: transparent;
  color: #a9b1d6;
  table-layout: fixed;
}
.vertical-row {
  float: left;
  height: 100%;
  width: 2px; /* edit this if you want */
  background-color: yellow;
}
.table-dark th {
  background-color: transparent;
  padding: 12px;
  text-align: left;
  color: #7aa2f7;
  position: sticky;
  top: 0;
  z-index: 1;
}

.table-row {
  border-bottom: 1px solid #414868;
  transition: background-color 0.3s ease;
}

.table-dark th,
.table-dark td {
  border: 1px solid #a9b1d6;
  color: yellow;
  padding: 8px;
}

.credit {
  color: #9ece6a !important;
}

.debit {
  color: #f7768e !important;
}

.valid {
  color: #9ece6a;
}

.invalid {
  color: #f7768e;
}

h3 {
  color: #7aa2f7;
  padding: 1rem;
  margin: 0;
}

.btn-primary-dark {
  background-color: transparent;
  border: 1px solid yellow;
  color: yellow;
  padding: 8px 16px;
  border-radius: 4px;
  transition: all 0.3s ease;
  width: 100%;
  margin-top: 10px;
}

.transfer-form {
  display: flex;
  color: yellow;
  margin-top: 1rem;
}

.transfer-form form {
  display: flex;
  flex-direction: column;
  width: 100%;
}

.transfer-form form div {
  margin-bottom: 10px;
}

.transfer-form input {
  background-color: transparent;
  color: yellow;
  border: 1px solid yellow;
  padding: 4px 8px;
  border-radius: 4px;
}

.transfer-form button {
  background-color: transparent;
  border: 1px solid yellow;
  color: yellow;
  padding: 8px 16px;
  border-radius: 4px;
  transition: all 0.3s ease;
  width: 100%;
  margin-top: 10px;
}

.transfer-form button:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}
table td {
  background-color: transparent;
}
</style>
