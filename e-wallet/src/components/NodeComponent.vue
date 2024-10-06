<template>
  <div class="node-wrapper">
    <div class="node">
      <div class="address">{{ shortenedAddress }}</div>
      <div class="balance">{{ balance }} ETH</div>
    </div>
  </div>
</template>

<script>
export default {
  name: 'NodeComponent',
  props: {
    address: {
      type: String,
      required: true
    },
    initialBalance: {
      type: Number,
      default: 0
    }
  },
  data() {
    return {
      balance: this.initialBalance
    }
  },
  computed: {
    shortenedAddress() {
      return `${this.address.substring(0, 6)}...`
    }
  },
  mounted() {
    // Simulate balance updates
    setInterval(() => {
      this.balance = +(Math.random() * 10).toFixed(2)
    }, 5000)
  }
}
</script>

<style scoped>
.node-wrapper {
  display: inline-block;
  margin: 10px;
}

.node {
  width: 120px;
  height: 120px;
  background: #f8f9fa;
  border: 2px solid #dee2e6;
  border-radius: 50%;
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  cursor: pointer;
  transition: all 0.3s ease;
}

.node:hover {
  transform: scale(1.05);
  box-shadow: 0 4px 12px rgba(0, 0, 0, 0.1);
}

.address {
  font-size: 14px;
  font-family: monospace;
  margin-bottom: 8px;
}

.balance {
  font-weight: bold;
  color: #198754;
}
</style>
