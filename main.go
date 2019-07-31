package main

import (
	"flag"
	"net/http"
	"os"
	"os/signal"

	"github.com/diegohce/httprecorder/recorder"
	"github.com/diegohce/logger"
)

var (
	log *logger.Logger
	srv *http.Server
)

func mode() bool {
	var recordMode bool
	var replayMode bool

	flag.BoolVar(&recordMode, "record", false, "Executes httprecorder in record mode")
	flag.BoolVar(&replayMode, "replay", false, "Executes httprecorder in replay mode")

	flag.Parse()

	if recordMode == replayMode {
		log.Error().Println("Must be one of -record or -replay")
		os.Exit(1)
	}

	return recordMode
}

func main() {

	log = logger.New("httprecorder::")

	recordMode := mode()

	bindAddr := os.Getenv("HTTPRECORDER_BINDADDR")
	if bindAddr == "" {
		bindAddr = ":8080"
	}

	if err := loadConfig(); err != nil {
		log.Error().Fatalln(err, "loading config file")
	}

	var proxies *http.ServeMux
	if recordMode {
		proxies = createRecordingProxies()

	} else {
		proxies = createReplayProxies()

	}

	if recordMode {
		log.Info().Println("Starting httprecorder on", bindAddr)

	} else {
		log.Info().Println("Starting httprecorder in replay mode on", bindAddr)
	}

	srv = &http.Server{Addr: bindAddr, Handler: proxies}

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt)
	go func() {
		_ = <-sigChan
		if recordMode {
			log.Info().Println("Dumping recorder data to", httprConfig.Filename)
			recorder.RRRecorder.Dump(httprConfig.Filename)
		}
		os.Exit(0)
	}()

	log.Error().Println(srv.ListenAndServe())
}
