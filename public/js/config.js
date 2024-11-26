const API_CONFIG = {
    baseUrl: `http://${API_PORT}`,
    endpoints: {
      pairs: "/pairs",
      info: "/info",
      status: "/status",
      send: "/send",
      wallets: "/wallets",
      addWallet: "/add-wallet",
      mempool: "/mempool",
      balances: "/balance"
    },
  };

  export default API_CONFIG;