<template>
  <div class="container mt-4">
    <div class="row">
      <div class="col-md-8">
        <div class="node-box position-relative">
          <!-- SVG for edges -->
          <svg class="edges-layer" :width="boxWidth" :height="boxHeight">
            <line
              v-for="edge in edges"
              :key="edge.id"
              :x1="edge.startPos.x"
              :y1="edge.startPos.y"
              :x2="edge.endPos.x"
              :y2="edge.endPos.y"
              stroke="#6c757d"
              stroke-width="2"
            >
              <title>{{ edge.amount }} ETH</title>
            </line>
          </svg>
          
          <!-- Nodes -->
          <NodeComponent
            v-for="node in nodes"
            :key="node.id"
            :id="node.id"
            :address="node.address"
            :initialBalance="node.balance"
            :isSelected="isNodeSelected(node)"
            @node-click="handleNodeClick"
            ref="nodes"
          />
        </div>
        <div class="text-center mt-3">
          <button class="btn btn-primary" @click="addNode">
            Add Node
          </button>
        </div>
      </div>
      <div class="col-md-4">
        <TransactionForm 
          ref="transactionForm"
          @transaction-sent="handleTransaction"
        />
      </div>
    </div>
  </div>
</template>

<script>
import NodeComponent from './NodeComponent.vue'
import TransactionForm from './TransactionForm.vue'

export default {
  name: 'NodeContainer',
  components: {
    NodeComponent,
    TransactionForm
  },
  data() {
    return {
      nodes: [],
      edges: [],
      selectedNodeId: null,
      boxWidth: 0,
      boxHeight: 0
    }
  },
  methods: {
    generateRandomAddress() {
      return '0x' + Math.random().toString(16).substr(2, 40)
    },
    addNode() {
      const newNode = {
        id: Date.now().toString(),
        address: this.generateRandomAddress(),
        balance: +(Math.random() * 10).toFixed(2)
      }
      this.nodes.push(newNode)
    },
    handleNodeClick(nodeData) {
      if (!this.selectedNodeId) {
        // First node selected - set as 'from'
        this.selectedNodeId = nodeData.id
        this.$refs.transactionForm.setFromAddress(nodeData.address)
      } else if (this.selectedNodeId !== nodeData.id) {
        // Second node selected - set as 'to'
        this.$refs.transactionForm.setToAddress(nodeData.address)
      }
    },
    isNodeSelected(node) {
      return node.id === this.selectedNodeId
    },
    handleTransaction(transaction) {
      // Create new edge
      const fromNode = this.nodes.find(n => n.address === transaction.from)
      const toNode = this.nodes.find(n => n.address === transaction.to)
      
      if (fromNode && toNode) {
        const fromEl = this.$refs.nodes.find(n => n.id === fromNode.id).$el
        const toEl = this.$refs.nodes.find(n => n.id === toNode.id).$el
        
        const fromRect = fromEl.getBoundingClientRect()
        const toRect = toEl.getBoundingClientRect()
        const boxRect = fromEl.parentElement.getBoundingClientRect()
        
        const edge = {
          id: Date.now().toString(),
          amount: transaction.amount,
          startPos: {
            x: fromRect.left - boxRect.left + fromRect.width/2,
            y: fromRect.top - boxRect.top + fromRect.height/2
          },
          endPos: {
            x: toRect.left - boxRect.left + toRect.width/2,
            y: toRect.top - boxRect.top + toRect.height/2
          }
        }
        
        this.edges.push(edge)
      }
      
      // Reset selection
      this.selectedNodeId = null
    },
    updateBoxDimensions() {
      const box = this.$el.querySelector('.node-box')
      this.boxWidth = box.offsetWidth
      this.boxHeight = box.offsetHeight
    }
  },
  mounted() {
    this.updateBoxDimensions()
    window.addEventListener('resize', this.updateBoxDimensions)
  },
  beforeUnmount() {
    window.removeEventListener('resize', this.updateBoxDimensions)
  }
}
</script>

<style scoped>
.node-box {
  min-height: 200px;
  border: 2px dashed #dee2e6;
  border-radius: 8px;
  padding: 20px;
  display: flex;
  flex-wrap: wrap;
  gap: 20px;
  align-items: center;
  justify-content: center;
}

.edges-layer {
  position: absolute;
  top: 0;
  left: 0;
  pointer-events: none;
  z-index: 1;
}
</style>
