package main

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"flag"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"AIComputingNode/pkg/config"
	"AIComputingNode/pkg/conngater"
	"AIComputingNode/pkg/db"
	"AIComputingNode/pkg/libp2p/host"
	"AIComputingNode/pkg/libp2p/stream"
	"AIComputingNode/pkg/log"
	"AIComputingNode/pkg/model"
	ps "AIComputingNode/pkg/pubsub"
	"AIComputingNode/pkg/selfupdate"
	"AIComputingNode/pkg/serve"
	"AIComputingNode/pkg/timer"
	"AIComputingNode/pkg/types"

	"github.com/gin-gonic/gin"
	"github.com/go-co-op/gocron/v2"
	"github.com/libp2p/go-libp2p"
	libp2phost "github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/network"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/libp2p/go-libp2p/core/peerstore"
	"github.com/libp2p/go-libp2p/core/protocol"
	"github.com/libp2p/go-libp2p/core/routing"
	drouting "github.com/libp2p/go-libp2p/p2p/discovery/routing"
	dutil "github.com/libp2p/go-libp2p/p2p/discovery/util"
	rcmgr "github.com/libp2p/go-libp2p/p2p/host/resource-manager"
	"github.com/libp2p/go-libp2p/p2p/net/connmgr"
	"github.com/libp2p/go-libp2p/p2p/protocol/ping"

	dht "github.com/libp2p/go-libp2p-kad-dht"

	pubsub "github.com/libp2p/go-libp2p-pubsub"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var version string

const ProtocolVersion string = "aicn/0.0.1"

func main() {
	configPath := flag.String("config", "", "run using the configuration file")
	versionFlag := flag.Bool("version", false, "show version number and exit")
	initFlag := flag.String("init", "", "initialize configuration in input/worker mode")
	peerKeyPath := flag.String("peerkey", "", "parse or generate a key file based on the specified file path")
	pskFlag := flag.Bool("psk", false, "generate a random Pre-Shared Key")
	flag.Parse()

	if *versionFlag {
		fmt.Println(version)
		os.Exit(0)
	}

	if *initFlag == "client" || *initFlag == "server" || *initFlag == "input" || *initFlag == "worker" {
		config.Init(*initFlag)
		os.Exit(0)
	} else if *initFlag != "" {
		fmt.Println("only supports input or worker mode")
		os.Exit(0)
	}

	if *peerKeyPath != "" {
		config.PeerKeyParse(*peerKeyPath)
		os.Exit(0)
	}

	if *pskFlag {
		key := make([]byte, 32)
		_, err := rand.Read(key)
		if err != nil {
			fmt.Println("Generate Pre-Shared key:", err)
		} else {
			fmt.Println("Generate Pre-Shared key:", hex.EncodeToString(key))
		}
		os.Exit(0)
	}

	cfg, err := config.LoadConfig(*configPath)
	if err != nil {
		fmt.Println("Failed to load JSON configuration file:", err)
		os.Exit(1)
	}
	if err := cfg.Validate(); err != nil {
		fmt.Println("Invalid configuration item:", err)
		os.Exit(1)
	}

	if err := log.InitLogging(cfg.App.LogLevel, cfg.App.LogFile, cfg.App.LogOutput); err != nil {
		fmt.Println("Initialize the log module failed:", err)
		os.Exit(1)
	}

	log.Logger.Info("################################################################")
	log.Logger.Info("#                          START                               #")
	log.Logger.Info("################################################################")

	if err := model.InitModels(cfg.AIProjects); err != nil {
		log.Logger.Fatalf("Init models: %v", err)
	}

	if err := db.InitDb(db.InitOptions{
		Folder: cfg.App.Datastore,
		// the node is deployed on a public server && enable peers collect
		EnablePeersCollect: cfg.App.PeersCollect.Enabled,
	}); err != nil {
		log.Logger.Fatalf("Init database: %v", err)
	}

	PeersHistory, err := host.ConvertPeersFromStringMap(db.LoadPeerConnHistory())
	if err != nil {
		log.Logger.Fatalf("Load peer history: %v", err)
	}

	DefaultBootstrapPeers, err := host.ConvertPeersFromStringArray(cfg.Bootstrap)
	if err != nil {
		log.Logger.Fatalf("Parse bootstrap: %v", err)
	}
	if len(DefaultBootstrapPeers) < 1 {
		log.Logger.Fatal("Not enough bootstrap peers")
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	p2pCtx, p2pStopCancel := context.WithCancel(ctx)

	var kadDHT *dht.IpfsDHT
	var pingService *ping.PingService = nil

	privKey, _ := host.PrivKeyFromString(cfg.Identity.PrivKey)
	connGater := &conngater.ConnectionGater{}

	// https://github.com/ipfs/kubo/issues/9322
	// https://github.com/ipfs/kubo/pull/9351/files
	// https://github.com/ipfs/kubo/issues/9432
	rcmgr.MustRegisterWith(prometheus.DefaultRegisterer)
	strpt, err := rcmgr.NewStatsTraceReporter()
	if err != nil {
		log.Logger.Fatalf("NewStatsTraceReporter: %v", err)
	}
	rclimits := rcmgr.DefaultLimits
	libp2p.SetDefaultServiceLimits(&rclimits)
	resmgr, err := rcmgr.NewResourceManager(rcmgr.NewFixedLimiter(rclimits.AutoScale()), rcmgr.WithTraceReporter(strpt))
	if err != nil {
		log.Logger.Fatalf("NewResourceManager: %v", err)
	}

	opts := []libp2p.Option{
		libp2p.ListenAddrStrings(cfg.Addresses...),
		libp2p.Identity(privKey),
		libp2p.DefaultMuxers,
		libp2p.DefaultSecurity,
		libp2p.ProtocolVersion(ProtocolVersion),
		libp2p.UserAgent(version),
		libp2p.ConnectionGater(connGater),
		// libp2p.DefaultResourceManager,
		libp2p.ResourceManager(resmgr),
	}
	if cfg.Swarm.ConnMgr.Type == "basic" {
		gracePeriod, _ := time.ParseDuration(cfg.Swarm.ConnMgr.GracePeriod)
		connmgr, err := connmgr.NewConnManager(
			cfg.Swarm.ConnMgr.LowWater,
			cfg.Swarm.ConnMgr.HighWater,
			connmgr.WithGracePeriod(gracePeriod),
		)
		if err != nil {
			log.Logger.Fatalf("Create connection manager: %v", err)
		}
		opts = append(opts, libp2p.ConnectionManager(connmgr))
	}
	if cfg.App.PreSharedKey != "" {
		psk, err := hex.DecodeString(cfg.App.PreSharedKey)
		if err != nil {
			log.Logger.Fatalf("Decoding PSK: %v", err)
		}
		opts = append(opts, libp2p.PrivateNetwork(psk), libp2p.DefaultPrivateTransports)
	} else {
		opts = append(opts, libp2p.DefaultTransports)
	}
	if cfg.Routing.Type != "none" {
		opts = append(opts, libp2p.Routing(func(h libp2phost.Host) (routing.PeerRouting, error) {
			dhtOpts := []dht.Option{}
			if cfg.Routing.Type == "auto" {
				dhtOpts = append(dhtOpts, dht.Mode(dht.ModeAuto))
			} else if cfg.Routing.Type == "dhtclient" {
				dhtOpts = append(dhtOpts, dht.Mode(dht.ModeClient))
			} else if cfg.Routing.Type == "dhtserver" {
				dhtOpts = append(dhtOpts, dht.Mode(dht.ModeServer))
			}
			if cfg.Routing.ProtocolPrefix != "" {
				dhtOpts = append(dhtOpts, dht.ProtocolPrefix(protocol.ID(cfg.Routing.ProtocolPrefix)))
			}
			dhtOpts = append(dhtOpts, dht.BootstrapPeers(DefaultBootstrapPeers...))
			kadDHT, err = dht.New(p2pCtx, h, dhtOpts...)
			return kadDHT, err
		}))
	}
	if cfg.Swarm.RelayClient.Enabled {
		opts = append(opts, libp2p.EnableAutoRelayWithStaticRelays(DefaultBootstrapPeers))
	}
	if cfg.Swarm.RelayService.Enabled {
		opts = append(opts, libp2p.EnableRelayService())
	}
	if cfg.Swarm.RelayService.Enabled {
		opts = append(opts, libp2p.Ping(false))
	}
	if !cfg.Swarm.DisableNatPortMap {
		opts = append(opts, libp2p.NATPortMap())
	}
	if cfg.Swarm.EnableAutoNATService {
		opts = append(opts, libp2p.EnableNATService())
	}
	if cfg.Swarm.EnableHolePunching {
		opts = append(opts, libp2p.EnableHolePunching())
	}
	if dialTimeout, err := time.ParseDuration(cfg.Swarm.DialTimeout); err == nil && dialTimeout != 0 {
		opts = append(opts, libp2p.WithDialTimeout(dialTimeout))
	}
	h, err := libp2p.New(opts...)
	if err != nil {
		log.Logger.Fatalf("Create libp2p host: %v", err)
	}
	log.Logger.Info("Listen addresses:", h.Addrs())
	log.Logger.Info("Node id:", h.ID())

	// print the node's PeerInfo in multiaddr format
	peerInfo := peer.AddrInfo{
		ID:    h.ID(),
		Addrs: h.Addrs(),
	}
	addrs, err := peer.AddrInfoToP2pAddrs(&peerInfo)
	log.Logger.Info("libp2p node address:", addrs) // addrs[0]

	h.Network().Notify(&network.NotifyBundle{
		ConnectedF: func(n network.Network, c network.Conn) {
			log.Logger.Infof("OnConnected remote multi-addr %v %v", c.RemoteMultiaddr(), c.RemotePeer())
			db.PeerConnected(c.RemotePeer().String(), c.RemoteMultiaddr().String())
		},
		DisconnectedF: func(n network.Network, c network.Conn) {
			log.Logger.Infof("OnDisconnected remote multi-addr %v %v", c.RemoteMultiaddr(), c.RemotePeer())
			// db.PeerDisconnected(c.RemotePeer().String(), c.RemoteMultiaddr().String())
		},
	})

	// Topic publish channel
	publishChan := make(chan []byte, 1024)

	libp2pStream := stream.NewLibp2pStream(publishChan)
	h.SetStreamHandler(types.ChatProxyProtocol, libp2pStream.ChatProxyStreamHandler)

	pingCtx, pingStopCancel := context.WithCancel(ctx)
	if cfg.Swarm.RelayService.Enabled {
		pingService = &ping.PingService{Host: h}
		h.SetStreamHandler(ping.ID, pingService.PingHandler)
	}

	if cfg.Routing.Type == "none" {
		dhtOpts := []dht.Option{
			dht.Mode(dht.ModeClient),
		}
		if cfg.Routing.ProtocolPrefix != "" {
			dhtOpts = append(dhtOpts, dht.ProtocolPrefix(protocol.ID(cfg.Routing.ProtocolPrefix)))
		}
		kadDHT, err = dht.New(p2pCtx, h, dhtOpts...)
		if err != nil {
			log.Logger.Fatalf("Create Kademlia DHT: %v", err)
		}
	}

	psOpts := []pubsub.Option{
		// pubsub.WithDiscovery(routingDiscovery),
		pubsub.WithEventTracer(&ps.Tracer{}),
		// pubsub.WithRawTracer(&ps.RawTracer{}),
	}
	var gs *pubsub.PubSub

	if cfg.Pubsub.Router == "gossipsub" {
		psOpts = append(psOpts,
			pubsub.WithDirectPeers(DefaultBootstrapPeers),
			pubsub.WithDirectConnectTicks(30),
		)
		if cfg.Routing.Type == "dhtserver" || cfg.Swarm.RelayService.Enabled {
			psOpts = append(psOpts, pubsub.WithPeerExchange(true))
		}
		if cfg.Pubsub.FloodPublish {
			psOpts = append(psOpts, pubsub.WithFloodPublish(true))
		}
		gs, err = pubsub.NewGossipSub(p2pCtx, h, psOpts...)
	} else {
		gs, err = pubsub.NewFloodSub(p2pCtx, h, psOpts...)
	}
	if err != nil {
		log.Logger.Fatalf("New %s: %v", cfg.Pubsub.Router, err)
	}

	// Let's connect to the bootstrap nodes first. They will tell us about the
	// other nodes in the network.
	errs := make(chan error, len(DefaultBootstrapPeers))
	mapBootstrapIds := make(map[peer.ID]struct{})
	var wg sync.WaitGroup
	for _, peerinfo := range DefaultBootstrapPeers {
		mapBootstrapIds[peerinfo.ID] = struct{}{}

		wg.Add(1)
		go func(pi peer.AddrInfo) {
			defer wg.Done()

			h.Peerstore().AddAddrs(pi.ID, pi.Addrs, peerstore.PermanentAddrTTL)
			if err := h.Connect(p2pCtx, pi); err != nil {
				log.Logger.Warnf("Connect bootstrap node %v : %v", pi, err)
				errs <- err
				return
			}
			log.Logger.Info("Connection established with bootstrap node:", pi)
		}(peerinfo)
	}
	wg.Wait()

	// our failure condition is when no connection attempt succeeded.
	// So drain the errs channel, counting the results.
	close(errs)
	errCount := 0
	for err = range errs {
		if err != nil {
			errCount++
		}
	}
	if errCount == len(DefaultBootstrapPeers) {
		// log.Logger.Fatalf("Failed to bootstrap. %s", err)
		log.Logger.Warnf("Failed to bootstrap. %s", err)
	}

	if !cfg.Swarm.RelayService.Enabled {
		for _, peerinfo := range PeersHistory {
			if _, ok := mapBootstrapIds[peerinfo.ID]; ok {
				continue
			}

			wg.Add(1)
			go func(pi peer.AddrInfo) {
				defer wg.Done()
				if h.Network().Connectedness(pi.ID) == network.Connected {
					return
				}
				ispub := host.IsPublicNode(pi)
				if ispub {
					h.Peerstore().AddAddrs(pi.ID, pi.Addrs, peerstore.PermanentAddrTTL)
				}
				if err := h.Connect(p2pCtx, pi); err != nil {
					log.Logger.Warnf("Connect history node %v : %v", pi, err)
					db.PeerConnectFailed(pi.ID.String())
					return
				}
				log.Logger.Info("Connection established with history node:", pi)
			}(peerinfo)
		}
		wg.Wait()
	}

	if err := kadDHT.Bootstrap(p2pCtx); err != nil {
		log.Logger.Fatalf("Bootstrap the host: %v", err)
	}

	routingDiscovery := drouting.NewRoutingDiscovery(kadDHT)
	dutil.Advertise(p2pCtx, routingDiscovery, cfg.App.TopicName)

	topic, err := gs.Join(cfg.App.TopicName)
	if err != nil {
		log.Logger.Fatalf("Join PubSub: %v", err)
	}
	defer topic.Close()
	sub, err := topic.Subscribe()
	if err != nil {
		log.Logger.Fatalf("Subscribe PubSub: %v", err)
	}

	pubCtx, pubStopCancel := context.WithCancel(ctx)
	subCtx, subStopCancel := context.WithCancel(ctx)
	timerCtx, timerStopCancel := context.WithCancel(ctx)

	host.Hio = &host.HostInfo{
		Host:            h,
		UserAgent:       version,
		ProtocolVersion: ProtocolVersion,
		PrivKey:         privKey,
		PingService:     pingService,
		Dht:             kadDHT,
		RD:              routingDiscovery,
		Topic:           topic,
	}

	var activeHttpReqs int32 = 0
	// router := gin.Default()
	// router.Use(gin.Recovery())
	// router.Use(errorHandler)
	router := gin.New()
	router.HandleMethodNotAllowed = true
	// router.NoRoute(func(ctx *gin.Context) {
	// 	ctx.JSON(http.StatusNotFound, gin.H{"code": "PAGE_NOT_FOUND", "message": "Page not found"})
	// })
	// router.NoMethod(func(ctx *gin.Context) {
	// 	ctx.JSON(http.StatusMethodNotAllowed, gin.H{"code": "METHOD_NOT_ALLOWED", "message": "Method not allowed"})
	// })
	router.Use(
		log.GinzapWithConfig(
			&log.GinConfig{
				SkipPaths: []string{},
				Skip:      nil,
			},
			&activeHttpReqs,
		),
		log.GinzapRecovery(true),
	)
	// router.GET("/api/v0/id", serve.IdHandler)
	v0 := router.Group("/api/v0")
	{
		v0.GET("/id", serve.IdHandler)
		v0.GET("/peers", serve.PeersHandler)
		v0.POST("/peer", func(ctx *gin.Context) {
			serve.PeerHandler(ctx, publishChan)
		})
		v0.POST("/host/info", func(ctx *gin.Context) {
			serve.HostInfoHandler(ctx, publishChan)
		})
		v0.GET("/rendezvous/peers", serve.RendezvousPeersHandler)
		v0.GET("/swarm/peers", serve.SwarmPeersHandler)
		v0.GET("/swarm/addrs", serve.SwarmAddrsHandler)
		v0.POST("/swarm/connect", serve.SwarmConnectHandler)
		v0.POST("/swarm/disconnect", serve.SwarmDisconnectHandler)
		v0.GET("/pubsub/peers", serve.PubsubPeersHandler)

		v0.POST("/chat/completion", func(ctx *gin.Context) {
			serve.ChatCompletionHandler(ctx, publishChan)
		})
		v0.POST("/chat/completion/proxy", func(ctx *gin.Context) {
			serve.ChatCompletionProxyHandler(ctx, publishChan)
		})
		v0.POST("/image/gen", func(ctx *gin.Context) {
			serve.ImageGenHandler(ctx, publishChan)
		})
		v0.POST("/image/gen/proxy", func(ctx *gin.Context) {
			serve.ImageGenProxyHandler(ctx, publishChan)
		})
		v0.POST("/image/edit", func(ctx *gin.Context) {
			serve.ImageEditHandler(ctx, publishChan)
		})
		v0.POST("/image/edit/proxy", func(ctx *gin.Context) {
			serve.ImageEditProxyHandler(ctx, publishChan)
		})

		v0.POST("/ai/project/register", func(ctx *gin.Context) {
			serve.RegisterAIProjectHandler(ctx, *configPath, publishChan)
		})
		v0.POST("/ai/project/unregister", func(ctx *gin.Context) {
			serve.UnregisterAIProjectHandler(ctx, *configPath, publishChan)
		})
		v0.POST("/ai/project/peer", func(ctx *gin.Context) {
			serve.GetAIProjectOfNodeHandler(ctx, publishChan)
		})
		v0.GET("/ai/projects/list", serve.ListAIProjectsHandler)
		v0.GET("/ai/projects/models", serve.GetModelsOfAIProjectHandler)
		v0.GET("/ai/projects/peers", serve.GetPeersOfAIProjectHandler)
		v0.POST("/ai/model/register", func(ctx *gin.Context) {
			serve.RegisterAIModelHandler(ctx, *configPath, publishChan)
		})
		v0.POST("/ai/model/unregister", func(ctx *gin.Context) {
			serve.UnregisterAIModelHandler(ctx, *configPath, publishChan)
		})

		v0.GET("/bootstrap/list", serve.ListBootstrapHandler)
		v0.POST("/bootstrap/add", func(ctx *gin.Context) {
			serve.AddBootstrapHandler(ctx, *configPath)
		})
		v0.POST("/bootstrap/rm", func(ctx *gin.Context) {
			serve.RemoveBootstrapHandler(ctx, *configPath)
		})

		v0.GET("/debug/metrics/prometheus", gin.WrapH(promhttp.Handler()))
	}
	srv := &http.Server{
		Addr:    cfg.API.Addr,
		Handler: router,
		// ReadTimeout:  120 * time.Second,
		// WriteTimeout: 120 * time.Second,
		// IdleTimeout:  120 * time.Second,
	}

	heartbeatInterval, _ := time.ParseDuration(cfg.App.PeersCollect.HeartbeatInterval)
	scheduler, err := gocron.NewScheduler()
	if err != nil {
		log.Logger.Fatalf("NewScheduler failed: %v", err)
	}
	job1, err := scheduler.NewJob(
		gocron.DurationJob(heartbeatInterval),
		gocron.NewTask(
			func(pcn chan<- []byte) {
				timer.SendAIProjects(pcn)
				db.CleanExpiredPeerCollectInfo()
			},
			publishChan,
		),
	)
	if err != nil {
		log.Logger.Fatalf("Create scheduled ai projects job failed: %v", err)
	}
	log.Logger.Infof("Scheduled ai projects job: %v", job1.ID())
	if cfg.App.AutoUpgrade.Enabled {
		upgraderInterval, _ := time.ParseDuration(cfg.App.AutoUpgrade.TimeInterval)
		job2, err := scheduler.NewJob(
			gocron.DurationJob(upgraderInterval),
			gocron.NewTask(
				func(ctx context.Context, timeout time.Duration, cur_version string) {
					upgradeCtx, pgradeCancel := context.WithTimeout(ctx, timeout)
					defer pgradeCancel()
					selfupdate.UpdateGithubLatestRelease(upgradeCtx, cur_version, &activeHttpReqs)
				},
				timerCtx,
				upgraderInterval,
				version,
			),
		)
		if err != nil {
			log.Logger.Fatalf("Create scheduled selftupdate job failed: %v", err)
		}
		log.Logger.Infof("Scheduled selfupdate job: %v", job2.ID())
	}

	pst := ps.NewPubSub(topic, sub, publishChan)
	go pst.PublishToTopic(pubCtx)
	scheduler.Start()
	go pst.ReadFromTopic(subCtx)
	// serve.NewHttpServe(router, publishChan, *configPath)
	go func() {
		log.Logger.Info("HTTP server is running on http://", cfg.API.Addr)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Logger.Fatalf("Start HTTP Server: %v", err)
		}
		log.Logger.Info("HTTP server is stopped")
	}()
	host.Hio.StartPingService(pingCtx)

	log.Logger.Info("listening for connections")

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)
	// select {} // hang forever
	<-stop
	// Stop PingService
	pingStopCancel()
	// serve.StopHttpService()
	httpStopCtx, httpStopCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer httpStopCancel()
	if err := srv.Shutdown(httpStopCtx); err != nil {
		log.Logger.Fatalf("Shutdown HTTP Server: %v", err)
	} else {
		log.Logger.Info("HTTP server is shutdown gracefully")
	}
	subStopCancel()
	timerStopCancel()
	if err := scheduler.Shutdown(); err != nil {
		log.Logger.Errorf("Error shutdowning scheduler: %v", err)
	}
	pubStopCancel()

	p2pStopCancel()

	if err := kadDHT.Close(); err != nil {
		log.Logger.Errorf("Error closing kadDHT: %v", err)
	}
	if err := h.Close(); err != nil {
		log.Logger.Errorf("Error closing host: %v", err)
	}

	log.Logger.Info("################################################################")
	log.Logger.Info("#                          OVER                                #")
	log.Logger.Info("################################################################")
}
