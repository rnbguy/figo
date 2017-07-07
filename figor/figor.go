package main

import (
	"bufio"
	"context"
	"flag"
	"fmt"
	"github.com/grandcat/zeroconf"
	"gopkg.in/cheggaaa/pb.v1"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"regexp"
	"time"
)

func LookupZeroconf() (addr []byte, port int) {
	resolver, err := zeroconf.NewResolver(nil)
	if err != nil {
		log.Fatalln("Failed to initialize resolver:", err.Error())
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*15)
	defer cancel()

	entries := make(chan *zeroconf.ServiceEntry)
	go func(results <-chan *zeroconf.ServiceEntry) {
		for entry := range results {
			port = entry.Port
			addr = entry.AddrIPv4[0]
			cancel()
		}
	}(entries)

	err = resolver.Lookup(ctx, "Figo", "_workstation._tcp", "local.", entries)
	if err != nil {
		panic(err)
	}

	<-ctx.Done()
	return
}

func main() {
	log.SetOutput(ioutil.Discard)
	addr, port := LookupZeroconf()
	flag.Parse()
	var file_name string
	if t := flag.Arg(0); t == "" {
		fmt.Println("Give a file")
		os.Exit(1)
	} else {
		file_name = t
	}
	urlPath := fmt.Sprintf("http://%v.%v.%v.%v:%d/%s", addr[0], addr[1], addr[2], addr[3], port, file_name)
	if resp, err := http.Get(urlPath); err != nil {
		panic(err)
	} else if resp.StatusCode != 200 {
		fmt.Println("Not correct file/code")
		os.Exit(1)
	} else {
		rgx := regexp.MustCompile("inline; filename=\"(.*)\"")
		filename := rgx.FindStringSubmatch(resp.Header.Get("Content-Disposition"))[1]
		if of, err := os.Create(filename); err != nil {
			panic(err)
		} else {
			var datalen int
			fmt.Sscanf(resp.Header.Get("Content-Length"), "%d", &datalen)

			resp_buf := bufio.NewReader(resp.Body)
			of_buf := bufio.NewWriter(of)

			bar := pb.New(datalen).SetUnits(pb.U_BYTES)
			bar.Start()
			bar_writer := io.MultiWriter(of_buf, bar)
			io.Copy(bar_writer, resp_buf)
			bar.Finish()
		}
	}
}
