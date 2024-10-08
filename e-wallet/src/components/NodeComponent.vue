<template>
  <div class="node-cell" @click="handleClick">
    <component
      :is="componentType"
      :value="formattedValue"
      :class="{ selected: isSelected }"
    />
  </div>
</template>

<script>
export default {
  name: "NodeComponent",
  props: {
    nodeData: {
      type: Object,
      required: true,
    },
    property: {
      type: String,
      required: true,
    },
    isSelected: {
      type: Boolean,
      default: false,
    },
  },
  computed: {
    componentType() {
      // Determine the appropriate component type based on the property
      switch (this.property) {
        case "address":
          return "address-display";
        case "balance":
          return "balance-display";
        default:
          return "default-display";
      }
    },
    formattedValue() {
      const value = this.nodeData[this.property];
      switch (this.property) {
        case "address":
          return this.shortenAddress(value);
        case "balance":
          return `${value} ETH`;
        default:
          return value;
      }
    },
  },
  methods: {
    shortenAddress(address) {
      return `${address.substring(0, 6)}...${address.substring(
        address.length - 4
      )}`;
    },
    handleClick() {
      this.$emit("node-click", this.nodeData);
    },
  },
};
</script>

<style scoped>
.node-cell {
  cursor: pointer;
  transition: all 0.2s ease;
}

.node-cell:hover {
  opacity: 0.8;
}

.selected {
  background-color: rgba(122, 162, 247, 0.1);
  border-radius: 4px;
}

.address-display {
  font-family: monospace;
  color: #9ece6a;
}

.balance-display {
  color: #7aa2f7;
  font-weight: bold;
}

.default-display {
  color: #a9b1d6;
}
</style>
