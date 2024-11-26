import BlockchainDebugger from "./BlockchainDebugger.js";
import  TerminalDebugger  from "./terminal.js";



// Initialize debugger when DOM is loaded
document.addEventListener("DOMContentLoaded", () => {
  window.debugger = new BlockchainDebugger();
  window.terminal = new TerminalDebugger();
  window.debugger.initialize();
  window.terminal.initialize();
});




