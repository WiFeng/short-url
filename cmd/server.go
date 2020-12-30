package main

import (
	"database/sql"
	"flag"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"text/tabwriter"

	"github.com/go-redis/redis"
	"github.com/oklog/oklog/pkg/group"
	"github.com/opentracing/opentracing-go"
	jaegerconfig "github.com/uber/jaeger-client-go/config"
	"github.com/uber/jaeger-lib/metrics/prometheus"

	"github.com/WiFeng/short-url/pkg/core/config"
	"github.com/WiFeng/short-url/pkg/core/log"
	"github.com/WiFeng/short-url/pkg/endpoint"
	"github.com/WiFeng/short-url/pkg/service"
	"github.com/WiFeng/short-url/pkg/transport"
)

func main() {
	// Define our flags.
	fs := flag.NewFlagSet("short-url", flag.ExitOnError)
	var (
		// httpAddr    = fs.String("http-addr", ":8081", "HTTP listen address")
		environment = fs.String("env", "development", "Runing environment")
	)
	fs.Usage = usageFor(fs, os.Args[0]+" [flags]")
	fs.Parse(os.Args[1:])

	var conf = &config.Config{}
	var confFile string
	{
		confFile = "./conf/config.toml"
		if *environment != "" {
			confFile = fmt.Sprintf("./conf/config_%s.toml", *environment)
		}
		if _, err := config.DecodeFile(confFile, &conf); err != nil {
			fmt.Println("config.DecodeFile error.", err)
			os.Exit(1)
		}
	}

	// Create a single logger, which we'll use and give to other components.
	var err error
	var logger log.Logger
	{
		logger, err = log.NewLogger(conf)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
		log.SetDefaultLogger(logger)
		defer logger.Sync()
	}

	var tracer opentracing.Tracer
	var tracerCloser io.Closer
	{
		serviceName := conf.Server.Name
		metricsFactory := prometheus.New()
		tracer, tracerCloser, err = jaegerconfig.Configuration{
			ServiceName: serviceName,
		}.NewTracer(
			jaegerconfig.Metrics(metricsFactory),
		)

		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		defer tracerCloser.Close()
		opentracing.InitGlobalTracer(tracer)
	}

	// Create a db client
	var db *sql.DB
	{

	}

	// Create a redis client
	var redisCli *redis.Client
	{
		addr := fmt.Sprintf("%s:%d", conf.Redis.Host, conf.Redis.Port)
		pass := conf.Redis.Auth
		db := conf.Redis.Db
		redisCli = redis.NewClient(&redis.Options{
			Addr:     addr, // use default Addr
			Password: pass, // no password set
			DB:       db,   // use default DB
		})

		if _, err := redisCli.Ping().Result(); err != nil {
			logger.Fatalw("redis ping error", "err", err)
			os.Exit(1)
		}
	}

	// Build the layers of the service "onion" from the inside out. First, the
	// business logic service; then, the set of endpoints that wrap the service;
	// and finally, a series of concrete transport adapters. The adapters, like
	// the HTTP handler or the gRPC server, are the bridge between Go kit and
	// the interfaces that the transports expect. Note that we're not binding
	// them to ports or anything yet; we'll do that next.
	var (
		service     = service.New(conf, db, redisCli, logger)
		endpoints   = endpoint.New(service, logger)
		httpHandler = transport.NewHTTPHandler(endpoints, logger)
	)

	var g group.Group
	{
		// httpAddr is configurable
		httpAddr := &conf.Server.HTTP.Addr

		// The HTTP listener mounts the Go kit HTTP handler we created.
		httpListener, err := net.Listen("tcp", *httpAddr)
		if err != nil {
			logger.Fatalw("listen error", "transport", "HTTP", "during", "Listen", "err", err)
			os.Exit(1)
		}
		g.Add(func() error {
			logger.Infow("serve start", "transport", "HTTP", "addr", *httpAddr, "config", confFile)
			return http.Serve(httpListener, httpHandler)
		}, func(error) {
			httpListener.Close()
		})
	}

	{
		// This function just sits and waits for ctrl-C.
		cancelInterrupt := make(chan struct{})
		g.Add(func() error {
			c := make(chan os.Signal, 1)
			signal.Notify(c, syscall.SIGINT, syscall.SIGTERM)
			select {
			case sig := <-c:
				return fmt.Errorf("received signal %s", sig)
			case <-cancelInterrupt:
				return nil
			}
		}, func(error) {
			close(cancelInterrupt)
		})
	}
	logger.Info("serve exit. ", g.Run())
}

func usageFor(fs *flag.FlagSet, short string) func() {
	return func() {
		fmt.Fprintf(os.Stderr, "USAGE\n")
		fmt.Fprintf(os.Stderr, "  %s\n", short)
		fmt.Fprintf(os.Stderr, "\n")
		fmt.Fprintf(os.Stderr, "FLAGS\n")
		w := tabwriter.NewWriter(os.Stderr, 0, 2, 2, ' ', 0)
		fs.VisitAll(func(f *flag.Flag) {
			fmt.Fprintf(w, "\t-%s %s\t%s\n", f.Name, f.DefValue, f.Usage)
		})
		w.Flush()
		fmt.Fprintf(os.Stderr, "\n")
	}
}
