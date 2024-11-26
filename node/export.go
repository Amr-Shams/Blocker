package node

import (
	"fmt"
	"log"
	"net"

	"github.com/Amr-Shams/Blocker/util"
	"github.com/soheilhy/cmux"
	"github.com/spf13/cobra"
	"golang.org/x/sync/errgroup"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func StartWalletNodeCommand() *cobra.Command {
	return &cobra.Command{
		Use:     "wallet",
		Short:   "Start a wallet node",
		Long:    "a wallet node that can send and receive transactions",
		Example: "wallet print",
		Args:    cobra.ExactArgs(0),
		Run: func(cmd *cobra.Command, args []string) {
			n := NewWalletNode()
			bestPeer := n.DiscoverPeers()
			fmt.Printf("Node is connected to %d peers\n", len(n.AllPeersAddress()))
			if bestPeer != nil {
				fmt.Printf("Best peer is %s\n", bestPeer.GetAddress())
				fmt.Printf("Height is %d\n", bestPeer.Blockchain.GetHeight())
			}
			go func() {
				if err := SyncBlockchain(n, bestPeer); err != nil {
					log.Printf("Error syncing blockchain: %v", err)
				}
			}()

			// Set up server
			port := n.GetPort()
			lis, err := net.Listen("tcp", fmt.Sprintf(":%s", port))
			if err != nil {
				log.Fatalf("Failed to listen on port %s: %v", port, err)
			}

			conn, err := grpc.NewClient(n.GetAddress(), grpc.WithTransportCredentials(insecure.NewCredentials()))
			if err != nil {
				log.Fatalf("Failed to create gRPC connection: %v", err)
			}
			defer conn.Close()

			n.SetStatus(StateConnected)
			defer func() {
				n.SetStatus(StateIdle)
			}()

			// Initialize cmux
			m := cmux.New(lis)

			// Use specific matchers for gRPC, HTTP, and WebSocket
			wsL := m.Match(cmux.HTTP1HeaderField("Upgrade", "websocket"))
			grpcL := m.MatchWithWriters(
				cmux.HTTP2MatchHeaderFieldSendSettings("content-type", "application/grpc"),
			)
			httpL := m.Match(cmux.HTTP1Fast())

			// Create error group for managing goroutines
			g := new(errgroup.Group)

			// Start gRPC server
			g.Go(func() error {
				log.Printf("Starting gRPC server on port %s\n", port)
				if err := grpcServer(grpcL, n); err != nil {
					return fmt.Errorf("gRPC server error: %v", err)
				}
				return nil
			})

			// Start HTTP server
			g.Go(func() error {
				log.Printf("Starting HTTP server on port %s\n", port)
				if err := httpServer(httpL, n); err != nil {
					return fmt.Errorf("HTTP server error: %v", err)
				}
				return nil
			})

			// Start WebSocket server
			g.Go(func() error {
				log.Printf("Starting WebSocket server on port %s\n", port)
				if err := wsServer(wsL, n); err != nil {
					return fmt.Errorf("WebSocket server error: %v", err)
				}
				return nil
			})

			// Start cmux
			g.Go(func() error {
				log.Printf("Starting multiplexer on port %s\n", port)
				return m.Serve()
			})

			// Wait for all servers and handle errors
			if err := g.Wait(); err != nil {
				log.Fatalf("Server error: %v", err)
			}
		},
	}
}

func StartFullNodeCommand() *cobra.Command {
	return &cobra.Command{
		Use:     "node",
		Short:   "Start a full node",
		Long:    "A full node that can mine blocks and validate transactions",
		Example: "node -n portNumber",
		Args:    cobra.ExactArgs(0),
		Run: func(cmd *cobra.Command, args []string) {
			n := NewFullNode()
			b := n.DiscoverPeers()
			fmt.Printf("Node is connected to %d peers\n", len(n.AllPeersAddress()))

			conn, err := grpc.NewClient(n.GetAddress(), grpc.WithTransportCredentials(insecure.NewCredentials()))
			util.Handle(err)
			defer conn.Close()

			n.SetStatus(StateConnected)
			defer func() {
				n.SetStatus(StateIdle)
			}()
			go func() {
				if err := SyncBlockchain(n, b); err != nil {
					log.Printf("Error syncing blockchain: %v", err)
				}
			}()
			port := n.GetPort()
			lis, err := net.Listen("tcp", fmt.Sprintf(":%s", port))
			if err != nil {
				log.Fatalf("Failed to listen on port %s: %v", port, err)
			}

			m := cmux.New(lis)
			wsL := m.Match(cmux.HTTP1HeaderField("Upgrade", "websocket"))
			grpcL := m.MatchWithWriters(
				cmux.HTTP2MatchHeaderFieldSendSettings("content-type", "application/grpc"),
			)
			httpL := m.Match(cmux.HTTP1Fast())

			g := new(errgroup.Group)

			g.Go(func() error {
				log.Printf("Starting gRPC server on port %s\n", port)
				if err := grpcServer(grpcL, n); err != nil {
					return fmt.Errorf("gRPC server error: %v", err)
				}
				return nil
			})

			g.Go(func() error {
				log.Printf("Starting HTTP server on port %s\n", port)
				if err := httpServer(httpL, n); err != nil {
					return fmt.Errorf("HTTP server error: %v", err)
				}
				return nil
			})

			g.Go(func() error {
				log.Printf("Starting WebSocket server on port %s\n", port)
				if err := wsServer(wsL, n); err != nil {
					return fmt.Errorf("WebSocket server error: %v", err)
				}
				return nil
			})

			g.Go(func() error {
				log.Printf("Starting multiplexer on port %s\n", port)
				return m.Serve()
			})

			if err := g.Wait(); err != nil {
				log.Fatalf("Server error: %v", err)
			}
		},
	}
}
