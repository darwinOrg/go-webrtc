package dgwrtc

import (
	"github.com/pion/turn/v3"
	"log"
	"net"
	"strings"
	"time"
)

type TurnServerConfig struct {
	PublicIP              string
	Network               string
	Port                  int
	ThreadNum             int
	ListenerBuilder       func(network string, port int) (net.Listener, error)
	PacketConnBuilder     func(network string, port int) (net.PacketConn, error)
	RelayAddressGenerator turn.RelayAddressGenerator
	Realm                 string
	AuthSecret            string
	AuthLongTermDuration  time.Duration
}

type TurnServer struct {
	authSecret           string
	authLongTermDuration time.Duration
	server               *turn.Server
}

type UserCredentials struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

func (s *TurnServer) GenerateLongTermCredentials() (*UserCredentials, error) {
	username, password, err := turn.GenerateLongTermCredentials(s.authSecret, s.authLongTermDuration)
	if err != nil {
		return nil, err
	}

	return &UserCredentials{
		Username: username,
		Password: password,
	}, nil
}

func (s *TurnServer) Close() error {
	return s.server.Close()
}

func NewTurnServer(config *TurnServerConfig) *TurnServer {
	if len(config.PublicIP) == 0 {
		log.Fatalf("PublicIP is required")
	}
	if len(config.AuthSecret) == 0 {
		log.Fatal("AuthSecret is required")
	}

	sc := turn.ServerConfig{
		Realm: config.Realm,
		// Set AuthHandler callback
		// This is called every time a user tries to authenticate with the TURN server
		// Return the key for that user, or false when no user is found
		AuthHandler: turn.NewLongTermAuthHandler(config.AuthSecret, nil),
	}

	isTcp := strings.HasPrefix(config.Network, "tcp")

	if config.ThreadNum < 2 {
		if isTcp {
			fillListenerConfig(sc, config)
		} else {
			fillPacketConnConfig(sc, config)
		}
	} else {
		if isTcp {
			fillListenerConfigs(sc, config)
		} else {
			fillPacketConnConfigs(sc, config)
		}
	}

	server, err := turn.NewServer(sc)
	if err != nil {
		log.Panic(err)
	}

	return &TurnServer{
		authSecret:           config.AuthSecret,
		authLongTermDuration: config.AuthLongTermDuration,
		server:               server,
	}
}

func fillListenerConfig(sc turn.ServerConfig, config *TurnServerConfig) {
	// Create a TCP listener to pass into pion/turn
	// pion/turn itself doesn't allocate any TCP listeners, but lets the user pass them in
	// this allows us to add logging, storage or modify inbound/outbound traffic
	tcpListener, err := config.ListenerBuilder(config.Network, config.Port)
	if err != nil {
		log.Panicf("Failed to create TURN server listener: %s", err)
	}

	// ListenerConfig is a list of Listeners and the configuration around them
	sc.ListenerConfigs = []turn.ListenerConfig{
		{
			Listener:              tcpListener,
			RelayAddressGenerator: config.RelayAddressGenerator,
		},
	}
}

func fillPacketConnConfig(sc turn.ServerConfig, config *TurnServerConfig) {
	// Create a UDP listener to pass into pion/turn
	// pion/turn itself doesn't allocate any UDP sockets, but lets the user pass them in
	// this allows us to add logging, storage or modify inbound/outbound traffic
	udpListener, err := config.PacketConnBuilder(config.Network, config.Port)
	if err != nil {
		log.Panicf("Failed to create TURN server listener: %s", err)
	}

	// PacketConnConfigs is a list of UDP Listeners and the configuration around them
	sc.PacketConnConfigs = []turn.PacketConnConfig{
		{
			PacketConn:            udpListener,
			RelayAddressGenerator: config.RelayAddressGenerator,
		},
	}
}

func fillListenerConfigs(sc turn.ServerConfig, config *TurnServerConfig) {
	listenerConfigs := make([]turn.ListenerConfig, config.ThreadNum)
	for i := 0; i < config.ThreadNum; i++ {
		conn, listErr := config.ListenerBuilder(config.Network, config.Port)
		if listErr != nil {
			log.Fatalf("Failed to allocate TCP listener at %s:%s", config.Network, config.Port)
		}

		listenerConfigs[i] = turn.ListenerConfig{
			Listener:              conn,
			RelayAddressGenerator: config.RelayAddressGenerator,
		}

		log.Printf("signalingServer %d listening on %s\n", i, conn.Addr().String())
	}
	sc.ListenerConfigs = listenerConfigs
}

func fillPacketConnConfigs(sc turn.ServerConfig, config *TurnServerConfig) {
	packetConnConfigs := make([]turn.PacketConnConfig, config.ThreadNum)
	for i := 0; i < config.ThreadNum; i++ {
		conn, listErr := config.PacketConnBuilder(config.Network, config.Port)
		if listErr != nil {
			log.Fatalf("Failed to allocate UDP listener at %s:%s", config.Network, config.Port)
		}

		packetConnConfigs[i] = turn.PacketConnConfig{
			PacketConn:            conn,
			RelayAddressGenerator: config.RelayAddressGenerator,
		}

		log.Printf("signalingServer %d listening on %s\n", i, conn.LocalAddr().String())
	}
	sc.PacketConnConfigs = packetConnConfigs
}
