package main

import (
	"bufio"
	"context"
	"flag"
	"net/http"
	"os"
	"os/exec"
	"strings"
	"time"

	log "github.com/sirupsen/logrus"
)

func main() {
	var outputDir string
	var httpAddr string
	var help bool
	flag.StringVar(&outputDir, "output", "docs", "output directory")
	flag.StringVar(&httpAddr, "http", "localhost:6060", "godoc server listen on")
	flag.BoolVar(&help, "help", false, "help")
	flag.Parse()
	if help {
		flag.Usage()
		return
	}
	file, err := os.Open("go.mod")
	if err != nil {
		log.Errorln("open go.mod failed")
		return
	}
	defer file.Close()
	rd := bufio.NewReader(file)
	_, err = rd.ReadString(' ')
	if err != nil {
		log.Errorln("read go.mod failed")
		return
	}
	modName, err := rd.ReadString('\n')
	if err != nil {
		log.Errorln("read go.mod failed")
		return
	}
	modName = strings.TrimSpace(modName)
	log.Infoln("mod name: ", modName)
	ctx, cancel := context.WithCancel(context.Background())
	defer func() {
		cancel()
		<-ctx.Done()
		log.Infoln("godoc stopped")
	}()
	go func() {
		godocCmd := exec.CommandContext(ctx, "godoc", "-http", httpAddr)
		log.Infoln("run godoc:", godocCmd.String())
		godocCmd.Run()
	}()

	log.Infoln("ping godoc")
	for {
		resp, err := http.Get("http://" + httpAddr)
		if err == nil && resp.StatusCode == http.StatusOK {
			break
		}
		time.Sleep(300 * time.Millisecond)
	}

	log.Infoln("run wget")
	wgetCmd := exec.Command("wget", "-P", outputDir, "-r", "-np", "-N", "-E", "-p", "-k", "http://"+httpAddr+"/pkg/"+modName)
	wgetCmd.Run()
	log.Infoln("wget stopped")
}
