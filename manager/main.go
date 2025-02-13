// A server to help launch browser remotely
package main

import (
	"flag"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"os"

	"github.com/go-rod/rod/lib/launcher"
	"github.com/go-rod/rod/lib/launcher/flags"
	"github.com/go-rod/rod/lib/utils"
)

var (
	addr              = flag.String("address", ":7317", "the address to listen to")
	quiet             = flag.Bool("quiet", false, "silence the log")
	proxy             = flag.String("proxy", "", "the address of proxy server")
	proxyGen          = flag.String("proxyGen", "", "the address of proxy generator")
	allowAllPath      = flag.Bool("allow-all", false, "allow all path set by the client")
	optimizeDataUsage = flag.Bool("optimize-data-usage", true, "allow to optimize data usage")
)

func fetchProxy(url string) (string, error) {
	resp, err := http.Get(url)
	if err != nil {
		return "", fmt.Errorf("failed to fetch proxy: %v", err)
	}
	defer func() { _ = resp.Body.Close() }()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response: %v", err)
	}

	return string(body), nil
}

func main() {
	flag.Parse()

	m := launcher.NewManager()

	if !*quiet {
		m.Logger = log.New(os.Stdout, "", 0)
	}

	if *allowAllPath {
		m.BeforeLaunch = func(l *launcher.Launcher, _ http.ResponseWriter, _ *http.Request) {
			if *optimizeDataUsage {
				l.Set("disable-features", "OptimizationGuideModelDownloading,OptimizationHintsFetching,OptimizationTargetPrediction,OptimizationHints")
			}

			if *proxy != "" {
				l.Set(flags.ProxyServer, *proxy)
			}

			if *proxyGen != "" {
				proxyAddr, err := fetchProxy(*proxyGen)
				if err != nil {
					return
				}
				l.Set(flags.ProxyServer, proxyAddr)
			}
		}
	}

	l, err := net.Listen("tcp", *addr)
	if err != nil {
		utils.E(err)
	}

	if !*quiet {
		fmt.Println("[rod-manager] listening on:", l.Addr().String())
	}

	srv := &http.Server{Handler: m}
	utils.E(srv.Serve(l))
}
