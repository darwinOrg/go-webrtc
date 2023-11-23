package dgwrtc

import (
	"fmt"
	dglogger "github.com/darwinOrg/go-logger"
	"github.com/pion/turn/v3"
	"log"
	"net"
	"strings"
	"time"
)

type TurnServerConfig struct {
	PublicHost            string
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
	config *TurnServerConfig
	server *turn.Server
}

type UserCredentials struct {
	Realm    string
	Username string `json:"username"`
	Password string `json:"password"`
	Uris     []string
}

func (s *TurnServer) GenerateLongTermCredentials() (*UserCredentials, error) {
	username, password, err := turn.GenerateLongTermCredentials(s.config.AuthSecret, s.config.AuthLongTermDuration)
	if err != nil {
		return nil, err
	}

	return &UserCredentials{
		Realm:    s.config.Realm,
		Username: username,
		Password: password,
		Uris: []string{
			"turn:" + fmt.Sprintf("%s:%d", s.config.PublicHost, s.config.Port),
		},
	}, nil
}

func (s *TurnServer) Close() error {
	return s.server.Close()
}

func NewTurnServer(config *TurnServerConfig) *TurnServer {
	if len(config.PublicHost) == 0 {
		dglogger.ProdFatal("PublicHost is required")
		return nil
	}
	if len(config.AuthSecret) == 0 {
		dglogger.ProdFatal("AuthSecret is required")
		return nil
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
			fillListenerConfig(&sc, config)
		} else {
			fillPacketConnConfig(&sc, config)
		}
	} else {
		if isTcp {
			fillListenerConfigs(&sc, config)
		} else {
			fillPacketConnConfigs(&sc, config)
		}
	}

	server, err := turn.NewServer(sc)
	if err != nil {
		dglogger.ProdFatal(err)
		return nil
	}

	return &TurnServer{
		config: config,
		server: server,
	}
}

func fillListenerConfig(sc *turn.ServerConfig, config *TurnServerConfig) {
	// Create a TCP listener to pass into pion/turn
	// pion/turn itself doesn't allocate any TCP listeners, but lets the user pass them in
	// this allows us to add logging, storage or modify inbound/outbound traffic
	tcpListener, err := config.ListenerBuilder(config.Network, config.Port)
	if err != nil {
		dglogger.ProdFatalf("Failed to create TURN server listener: %s", err)
		return
	}

	// ListenerConfig is a list of Listeners and the configuration around them
	sc.ListenerConfigs = []turn.ListenerConfig{
		{
			Listener:              tcpListener,
			RelayAddressGenerator: config.RelayAddressGenerator,
		},
	}
}

func fillPacketConnConfig(sc *turn.ServerConfig, config *TurnServerConfig) {
	// Create a UDP listener to pass into pion/turn
	// pion/turn itself doesn't allocate any UDP sockets, but lets the user pass them in
	// this allows us to add logging, storage or modify inbound/outbound traffic
	udpListener, err := config.PacketConnBuilder(config.Network, config.Port)
	if err != nil {
		dglogger.ProdFatalf("Failed to create TURN server listener: %s", err)
		return
	}

	// PacketConnConfigs is a list of UDP Listeners and the configuration around them
	sc.PacketConnConfigs = []turn.PacketConnConfig{
		{
			PacketConn:            udpListener,
			RelayAddressGenerator: config.RelayAddressGenerator,
		},
	}
}

func fillListenerConfigs(sc *turn.ServerConfig, config *TurnServerConfig) {
	listenerConfigs := make([]turn.ListenerConfig, config.ThreadNum)
	for i := 0; i < config.ThreadNum; i++ {
		conn, listErr := config.ListenerBuilder(config.Network, config.Port)
		if listErr != nil {
			dglogger.ProdFatalf("Failed to allocate TCP listener at %s:%d", config.Network, config.Port)
			return
		}

		listenerConfigs[i] = turn.ListenerConfig{
			Listener:              conn,
			RelayAddressGenerator: config.RelayAddressGenerator,
		}

		log.Printf("Turn Server %d listening on %s\n", i, conn.Addr().String())
	}
	sc.ListenerConfigs = listenerConfigs
}

func fillPacketConnConfigs(sc *turn.ServerConfig, config *TurnServerConfig) {
	packetConnConfigs := make([]turn.PacketConnConfig, config.ThreadNum)
	for i := 0; i < config.ThreadNum; i++ {
		conn, listErr := config.PacketConnBuilder(config.Network, config.Port)
		if listErr != nil {
			dglogger.ProdFatalf("Failed to allocate UDP listener at %s:%d", config.Network, config.Port)
			return
		}

		packetConnConfigs[i] = turn.PacketConnConfig{
			PacketConn:            conn,
			RelayAddressGenerator: config.RelayAddressGenerator,
		}

		log.Printf("Turn Server %d listening on %s\n", i, conn.LocalAddr().String())
	}
	sc.PacketConnConfigs = packetConnConfigs
}
