package main

import (
	"context"
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"strings"

	"github.com/brunotm/tact"
	_ "github.com/brunotm/tact/collector/aix"
	_ "github.com/brunotm/tact/collector/linux"
	_ "github.com/brunotm/tact/collector/oracle"
	"github.com/brunotm/tact/log"
	"github.com/brunotm/tact/scheduler"
)

var (
	sched      = flag.Bool("sched", false, "Start scheduler")
	cron       = flag.String("cron", "0 */1 * * * *", "Cron like scheduling expression: 0 */1 * * * *")
	user       = flag.String("u", "", "user")
	password   = flag.String("p", "", "password")
	key        = flag.String("k", "", "ssh/sftp key file path")
	hostName   = flag.String("n", "", "hostname")
	netAddr    = flag.String("a", "", "network address")
	logFiles   = flag.String("l", "", "log files, format name:path,name:path")
	dbUser     = flag.String("dbuser", "", "log files, format name:path,name:path")
	dbPassword = flag.String("dbpass", "", "log files, format name:path,name:path")
	dbPort     = flag.String("dbport", "", "log files, format name:path,name:path")
	collector  = flag.String("c", "", "Collector or group to run")
	logLevel   = flag.String("log", "info", "Log level")
	dataPath   = flag.String("datapath", "./statedb", "Path for state data")
)

func main() {
	// grmon.Start()
	flag.Parse()
	var err error

	switch strings.ToLower(*logLevel) {
	case "debug":
		log.SetDebug()
	case "info":
		log.SetInfo()
	case "warn":
		log.SetWarn()
	case "error":
		log.SetError()
	default:
		log.Error("invalid log level")
		os.Exit(1)
	}

	tact.Init(*dataPath)

	node := &tact.Node{}
	node.HostName = *hostName
	if *netAddr == "" {
		node.NetAddr = *hostName
	} else {
		node.NetAddr = *netAddr
	}

	var sshKey []byte
	if *key != "" {
		sshKey, err = ioutil.ReadFile(*key)
		if err != nil {
			panic(err)
		}
	}

	node.SSHUser = *user
	node.SSHKey = sshKey
	node.SSHPassword = *password
	node.DBUser = *dbUser
	node.DBPassword = *dbPassword
	node.DBPort = *dbPort
	node.LogFiles = make(map[string]string) // map[string]string{"messages": "/var/log/messages"}

	if *logFiles != "" {
		for _, entry := range strings.Split(*logFiles, ",") {
			els := strings.Split(entry, ":")
			node.LogFiles[els[0]] = els[1]
		}
	}

	wchan := make(chan []byte)
	go func() {
		for e := range wchan {
			fmt.Println(string(e))
		}
	}()

	if collector == nil {
		panic("no colector specified")
	}

	var coll *tact.Collector
	var collGroup []*tact.Collector
	if len(strings.Split(*collector, "/")) > 3 {
		coll = tact.Registry.Get(*collector)
	} else {
		collGroup = tact.Registry.GetGroup(*collector)
	}

	if *sched {
		sched := scheduler.New(100, 60*time.Second, wchan)

		if collGroup != nil {
			for _, c := range collGroup {
				if err = sched.AddJob(*cron, c, node, 290*time.Second); err != nil {
					panic(err)
				}
			}
		}

		if coll != nil {
			if err = sched.AddJob(*cron, coll, node, 290*time.Second); err != nil {
				panic(err)
			}
		}

		sched.Start()
		quit := make(chan os.Signal, 1)
		signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
		<-quit
		sched.Stop()

	} else {
		var wg sync.WaitGroup

		if collGroup != nil {
			for _, c := range collGroup {
				sess, err := tact.NewContext(context.Background(), c.Name, node, tact.Store, 290*time.Second)
				if err != nil {
					panic(err)
				}
				wg.Add(1)
				go func(c *tact.Collector) {
					c.Start(sess, wchan)
					wg.Done()
				}(c)
			}
		}

		if coll != nil {
			sess, err := tact.NewContext(context.Background(), *collector, node, tact.Store, 290*time.Second)
			if err != nil {
				panic(err)
			}
			wg.Add(1)
			func() {
				coll.Start(sess, wchan)
				wg.Done()
			}()
		}

		wg.Wait()
	}

	log.Info("Shutting down")
	tact.Close()
}
