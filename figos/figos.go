package main

import (
	"bufio"
	"crypto/sha1"
	"encoding/base64"
	"flag"
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
	"path/filepath"
	"strings"
	"sync"
	"syscall"
	"time"
)

var (
	quit      chan bool
	wg        sync.WaitGroup
	file_path string
)

func RegisterZeroConf(port int, quit chan bool) {
	defer wg.Done()
	if server, err := zeroconf.Register("Figo", "_workstation._tcp", "local.", port, nil, nil); err != nil {
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

func main() {
	log.SetOutput(ioutil.Discard)
	port := 42423
	zeroconfQuit := make(chan bool)
	wg.Add(1)
	defer wg.Wait()
	flag.Parse()
	if t := flag.Arg(0); t == "" {
		fmt.Println("Give a file")
		os.Exit(1)
	} else {
		file_path, _ = filepath.Abs(t)
	}
	go RegisterZeroConf(port, zeroconfQuit)
	if listener, err := net.Listen("tcp", fmt.Sprintf(":%d", port)); err != nil {
		panic(err)
	} else {
		quit = make(chan bool)
		if fi, err := os.Stat(file_path); err != nil || !fi.Mode().IsRegular() {
			fmt.Println("Not valid file")
			os.Exit(1)
		}
		base_name := url.PathEscape(path.Base(file_path))
		hasher := sha1.New()
		hasher.Write([]byte(base_name))
		hasher.Write([]byte(time.Now().String()))
		uniq_id := strings.ToUpper(base64.URLEncoding.EncodeToString(hasher.Sum(nil))[:4])
		fmt.Println("Download using either of")
		fmt.Println("\tfigor", uniq_id)
		fmt.Println("\tfigor", base_name)
		http.HandleFunc(fmt.Sprintf("/%s", uniq_id), ServeHandler)
		http.HandleFunc(fmt.Sprintf("/%s", url.PathEscape(base_name)), ServeHandler)
		go http.Serve(listener, nil)
		sig := make(chan os.Signal, 1)
		signal.Notify(sig, os.Interrupt, syscall.SIGTERM)
		select {
		case <-sig:
		case <-quit:
		}
		listener.Close()
		zeroconfQuit <- true
		close(zeroconfQuit)
	}
}
