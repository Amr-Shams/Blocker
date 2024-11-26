# Blocker

Blocker is a blockchain network debugger that allows you to start and manage full nodes, wallet nodes, and interact with the blockchain through a web interface.

## Features

- Start and manage full nodes and wallet nodes
- Sync blockchain and validate transactions
- Submit transactions and add new wallets
- Visualize network graph and blockchain status
- Interactive terminal for debugging

## Getting Started
https://github.com/user-attachments/assets/afa6d305-dff6-4758-9e76-5884996d9532

### Prerequisites

Before you begin, ensure you have the following:

- **Go 1.16 or later** (for building the application)
- **protobuf** and **protoc** (for generating gRPC files)

If you don't have `protobuf` installed, you can follow the installation instructions from the [official protobuf site](https://developers.google.com/protocol-buffers).

### Step 1: Generate Protobuf Files

Blocker uses gRPC for communication, so you need to generate the necessary protobuf files first. To do this, run the following command:

```sh
protoc --go_out=. --go_opt=paths=source_relative \
       --go-grpc_out=. --go-grpc_opt=paths=source_relative \
       --plugin=protoc-gen-go=/path-to-go/bin/protoc-gen-go \
       --plugin=protoc-gen-go-grpc=/path-to-go/bin/protoc-gen-go-grpc \
       server/blockchain_service.proto
```

This will generate the required `.proto` files for gRPC communication in the project directory.

### Step 2: Installation

1. **Clone the repository:**

   ```sh
   git clone https://github.com/Amr-Shams/Blocker.git
   cd Blocker
   ```

2. **Install Go dependencies:**

   ```sh
   go mod tidy
   ```

### Step 3: Running the Application

You can run the application either locally or using Docker Compose. Below are the instructions for both methods.

#### Method 1: Running the Application Locally

1. **Build and run the Go application:**

   ```sh
   go build -o blocker main.go
   ./blocker node -n 5001
   ```

   This will start the application as a full node on port `5001`.

2. **Open the web interface:**

   Navigate to `http://localhost:5001` in your web browser to interact with the blockchain.

#### Method 2: Running with Docker Compose

1. **Build and run the Docker containers using Docker Compose:**

   ```sh
   docker-compose up --build
   ```

   This will start multiple instances of the application, each running on ports `5001` to `5004`. Full nodes and wallet nodes will be available on these ports, and you can interact with the blockchain through them.

2. **Access the web interface:**

   Navigate to `http://localhost:5001` in your browser to view the blockchain status.

> **Note**: The services are running on ports `5001` to `5004`, and each port corresponds to a different node type (full or wallet).

---

### Ports and Node Types

- **Full Node**: A full node is a server that validates and relays blockchain transactions and blocks. Full nodes are essential for syncing the blockchain and maintaining its integrity. You need at least one full node running to sync and validate the blockchain.

  - Start a full node by running:

    ```sh
    air node -n 5001
    ```

    You can run additional full nodes on different ports (e.g., `5002`, `5003`, `5004`):

    ```sh
    air node -n 5002
    air node -n 5003
    air node -n 5004
    ```

    **Note**: Running multiple full nodes increases the robustness of the network and ensures decentralization.

- **Wallet Node**: A wallet node is used for managing cryptocurrency wallets and interacting with the blockchain. Wallet nodes connect to full nodes to send and receive transactions. Wallet nodes cannot validate or mine blocks on their own.

  - Start a wallet node by running:

    ```sh
    air wallet -n 5001
    ```

    You can run additional wallet nodes on other ports (e.g., `5002`, `5003`, `5004`):

    ```sh
    air wallet -n 5002
    air wallet -n 5003
    air wallet -n 5004
    ```

    **Difference Between Full Node and Wallet Node**:
    - A **full node** validates and maintains the blockchain by ensuring blocks and transactions are legitimate.
    - A **wallet node** manages wallets and sends transactions. It relies on full nodes to validate transactions and maintain blockchain state.

    **Important**: To send a transaction, at least one full node must be running to validate and relay the transaction to the blockchain.

---

### Available Commands

Here are the main commands you can use in Blocker:

- **Start a Full Node**:

  ```sh
  air node -n 5001
  ```

  Replace `5001` with another port number to run additional full nodes.

- **Start a Wallet Node**:

  ```sh
  air wallet -n 5001
  ```

  Similarly, you can replace `5001` with other ports for additional wallet nodes.

- **Send a Transaction**:

  ```sh
  ./blocker send --from <address> --to <address> --amount <amount>
  ```

  This command sends a transaction from one address to another with a specified amount.

- **Create a New Blockchain**:

  ```sh
  ./blocker create --wallet <address>
  ```

  This command initializes a new blockchain and associates it with the specified wallet address.

- **View Available Commands**:  
  Use the `help` command in the browser's terminal interface (xterm.js) for a full list of commands:

  ```sh
  help
  ```

---

### Technology Stack

- **Communication:**
  - RESTful API
  - WebSocket
  - gRPC (for blockchain node interaction)

- **Libraries and Tools:**
  - `xterm.js` (for terminal emulator)
  - `VivaGraphJS` (for graph visualization)
  - Docker (for running multiple node instances)
 
- this implementation is part of the [series](https://www.youtube.com/watch?v=mYlHT9bB6OE&list=PLJbE2Yu2zumC5QE39TQHBLYJDB2gfFE5Q) but expnaded into a more and more features for the sake of education
- I have not tested the presentation of the app (frontend) only on safari if there are any problem I have no idea how to solve them

