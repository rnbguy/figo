package core

import (
	"bufio"
	"fmt"
	"github.com/grandcat/zeroconf"
	"gopkg.in/cheggaaa/pb.v1"
	"io"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"path"
	"sync"
	"syscall"
	"time"
)

var (
	start     chan bool
	quit      chan bool
	wg        sync.WaitGroup
	file_path string
)

func RegisterZeroConf(hash string, port int, quit chan bool) {
	defer wg.Done()
	if server, err := zeroconf.Register(hash, "_figo._http._tcp", "local.", port, nil, nil); err != nil {
		panic(err)
	} else {
		<-quit
		server.Shutdown()
	}
}

func ServeHandler(w http.ResponseWriter, r *http.Request) {
	if file, err := os.Open(file_path); err != nil {
		panic(err)
	} else {
		start <- true
		w.Header().Set("Content-Type", "application/octet-stream")
		var datalen int
		if stat, err := file.Stat(); err != nil {
			panic(err)
		} else {
			datalen = int(stat.Size())
			w.Header().Add("Content-Length", fmt.Sprintf("%d", datalen))
			w.Header().Add("Content-Disposition", "inline; filename=\""+stat.Name()+"\"")
		}

		file_buf := bufio.NewReader(file)
		w_buf := bufio.NewWriter(w)

		bar := pb.New(datalen).SetUnits(pb.U_BYTES)
		bar.Start()
		bar_reader := bar.NewProxyReader(file_buf)
		io.Copy(w_buf, bar_reader)
		bar.Finish()
		quit <- true
	}
}

func Figos(name string, nick string) {
	log.SetOutput(ioutil.Discard)
	if nick == "" {
		nick = GetNick(4)
	}
	port := 44234
	// zeroconfQuit := make(chan bool)
	file_path = name
	allowed := []string{path.Base(file_path), nick}
	allowed_hash := GetHashes(allowed)
	zeroconfQuits := make([]chan bool, 2)
	defer wg.Wait()
	for i, hash := range allowed_hash {
		wg.Add(1)
		zeroconfQuits[i] = make(chan bool)
		go RegisterZeroConf(hash, port, zeroconfQuits[i])
	}
	if listener, err := net.Listen("tcp", fmt.Sprintf(":%d", port)); err != nil {
		panic(err)
	} else {
		quit = make(chan bool, 1)
		start = make(chan bool, 1)
		if fi, err := os.Stat(file_path); err != nil || !fi.Mode().IsRegular() {
			fmt.Println("Not valid file")
			os.Exit(1)
		}
		base_name := url.PathEscape(path.Base(file_path))
		http.HandleFunc(fmt.Sprintf("/%s", nick), ServeHandler)
		http.HandleFunc(fmt.Sprintf("/%s", url.PathEscape(base_name)), ServeHandler)
		go http.Serve(listener, nil)

		{
			timer := time.NewTimer(time.Millisecond * 50)
			select {
			case <-start:
			case <-timer.C:
				fmt.Println("Download using either of")
				fmt.Println("\tfigor", nick)
				fmt.Println("\tfigor", base_name)
			}
			timer.Stop()
		}

		sig := make(chan os.Signal, 1)
		signal.Notify(sig, os.Interrupt, syscall.SIGTERM)
		select {
		case <-sig:
		case <-quit:
		}
		listener.Close()
		for _, quit := range zeroconfQuits {
			quit <- true
		}
	}
}
