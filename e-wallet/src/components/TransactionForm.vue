<template>
  <div class="transaction-form p-3 border rounded">
    <h5>Send Transaction</h5>
    <div class="mb-3">
      <label class="form-label">From Address</label>
      <input 
        type="text" 
        class="form-control" 
        v-model="fromAddress" 
       
      >
    </div>
    <div class="mb-3">
      <label class="form-label">To Address</label>
      <input 
        type="text" 
        class="form-control" 
        v-model="toAddress" 
      >
    </div>
    <div class="mb-3">
      <label class="form-label">Amount (ETH)</label>
      <input 
        type="number" 
        class="form-control" 
        v-model="amount"
        min="0"
        step="0.01"
      >
    </div>
    <button 
      class="btn btn-primary" 
      @click="sendTransaction"
      :disabled="!isValid"
    >
      Send
    </button>
  </div>
</template>

<script>
export default {
  name: 'TransactionForm',
  data() {
    return {
      fromAddress: '',
      toAddress: '',
      amount: 0
    }
  },
  computed: {
    isValid() {
      return this.fromAddress && 
             this.toAddress && 
             this.amount > 0 && 
             this.fromAddress !== this.toAddress
    }
  },
  methods: {
    setFromAddress(address) {
      this.fromAddress = address
    },
    setToAddress(address) {
      this.toAddress = address
    },
    sendTransaction() {
      if (this.isValid) {
        this.$emit('transaction-sent', {
          from: this.fromAddress,
          to: this.toAddress,
          amount: this.amount
        })
        // Reset form
        this.amount = 0
      }
    }
  }
}
</script>
