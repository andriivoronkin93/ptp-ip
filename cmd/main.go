package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"

	"github.com/malc0mn/ptp-ip/ip"
)

const (
	ok                  = 0
	errGeneral          = 1
	errInvalidArgs      = 2
	errOpenConfig       = 102
	errCreateClient     = 104
	errResponderConnect = 105
)

var (
	version   = "0.0.0"
	buildTime = "unknown"
	exe       string
	quit      = make(chan struct{}) // Should this be global or do we need to pass it along to all who need it?
)

func main() {
	exe = filepath.Base(os.Args[0])

	initFlags()

	if noArgs := len(os.Args) < 2; noArgs || showHelp {
		printUsage()
		exit := ok
		if noArgs {
			exit = errGeneral
		}
		os.Exit(exit)
	}

	if showVersion {
		// fmt.Printf("%s version %s built on %s\n", exe, version, buildTime)
		os.Exit(ok)
	}

	if file != "" {
		loadConfig()
	}

	checkPorts()

	if cmd != "" && (interactive || server) || (interactive && server) {
		fmt.Fprintln(os.Stderr, "Too many arguments: either run in server mode OR interactive mode OR execute a single command; not all at once!")
		os.Exit(errInvalidArgs)
	}

	// TODO: finish this implementation so CTRL+C will also abort client.Dial() etc. properly.
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		sig := <-sigs
		log.Printf("Received signal %s, shutting down...\n", sig)
		close(quit)
	}()

	client, err := ip.NewClient(conf.vendor, conf.host, uint16(conf.port), conf.fname, conf.guid, verbosity)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error creating PTP/IP client - %s\n", err)
		os.Exit(errCreateClient)
	}
	defer client.Close()

	if conf.cport != 0 {
		client.SetCommandDataPort(uint16(conf.cport))
	}
	if conf.eport != 0 {
		client.SetEventPort(uint16(conf.eport))
	}
	if conf.sport != 0 {
		client.SetStreamerPort(uint16(conf.sport))
	}

	// fmt.Printf("Created new client with name '%s' and GUID '%s'.\n", client.InitiatorFriendlyName(), client.InitiatorGUIDAsString())
	// fmt.Printf("Attempting to connect to %s\n", client.CommandDataAddress())
	err = client.Dial()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error connecting to responder - %s\n", err)
		os.Exit(errResponderConnect)
	}

	if cmd != "" {
		executeCommand(cmd, bufio.NewWriter(os.Stdout), client, "cli")
	}

	if server || interactive {
		if interactive {
			go iShell(client)
		}

		if server {
			go launchServer(client)
		}

		mainThread()

		<-quit
		fmt.Println("Bye bye!")
	}

	os.Exit(ok)
}
