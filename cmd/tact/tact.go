package main

import (
	"context"
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"os/signal"
	"syscall"
	"time"

	"strings"

	log "github.com/Sirupsen/logrus"
	"github.com/brunotm/tact"
	_ "github.com/brunotm/tact/collector/aix"
	_ "github.com/brunotm/tact/collector/linux"
	_ "github.com/brunotm/tact/collector/oracle"
	"github.com/brunotm/tact/scheduler"
)

var (
	sched      = flag.Bool("sched", false, "Start scheduler")
	user       = flag.String("u", "", "user")
	password   = flag.String("p", "", "password")
	key        = flag.String("k", "", "ssh/sftp key file path")
	hostName   = flag.String("n", "", "hostname")
	netAddr    = flag.String("a", "", "network address")
	logFiles   = flag.String("l", "", "log files, format name:path,name:path")
	dbUser     = flag.String("dbuser", "", "log files, format name:path,name:path")
	dbPassword = flag.String("dbpass", "", "log files, format name:path,name:path")
	dbPort     = flag.String("dbport", "", "log files, format name:path,name:path")
	collector  = flag.String("c", "", "Collector to run")
)

func main() {
	flag.Parse()

	log.SetFormatter(&log.JSONFormatter{})

	node := &tact.Node{}

	node.HostName = *hostName
	if *netAddr == "" {
		node.NetAddr = *hostName
	} else {
		node.NetAddr = *netAddr
	}

	var err error
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

	coll := tact.Registry.Get(*collector)
	if *sched {
		sched := scheduler.New(1, 60, wchan)
		if err = sched.AddJob("0 */1 * * * *", coll, node, 290*time.Second); err != nil {
			panic(err)
		}
		sched.Start()
		defer sched.Stop()
		quit := make(chan os.Signal, 1)
		signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
		<-quit

	} else {
		sess := tact.NewSession(context.Background(), *collector, node, 290*time.Second)
		coll.Start(sess, wchan)
		return
	}

	fmt.Println("#### Shutting down")
	tact.Close()
}
