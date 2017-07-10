package core

import (
	"bufio"
	"context"
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

func LookupZeroconf(id string, entries chan *zeroconf.ServiceEntry, stop chan bool) {
	resolver, err := zeroconf.NewResolver(nil)
	if err != nil {
		log.Fatalln("Failed to initialize resolver:", err.Error())
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// err = resolver.Lookup(ctx, id, "_figo._http._tcp", "local.", entries)
	err = resolver.Lookup(ctx, id, "_figo._http._tcp", "local.", entries)
	if err != nil {
		panic(err)
	}

	<-stop

	<-ctx.Done()
	return
}

func Figor(name string, out string) {
	log.SetOutput(ioutil.Discard)
	fmt.Printf("searching for %s to save as %s\n", name, out)

	var entry *zeroconf.ServiceEntry

	{
		entries := make(chan *zeroconf.ServiceEntry)

		stopLookup := make(chan bool)
		go LookupZeroconf(GetHash(name), entries, stopLookup)

		{
			timer := time.NewTimer(time.Millisecond * 50)
			for i := 0; i < 2; i++ {
				loopDone := false
				select {
				case <-timer.C:
					fmt.Println("couldn't find anything!")
					fmt.Println("figos "+name, "or", "figos <filename> "+name)
				case entry = <-entries:
					fmt.Println(entry)
					timer.Stop()
					stopLookup <- true
					loopDone = true
				}
				if loopDone {
					break
				}
			}
		}
	}

	fmt.Println(entry)
	addr := entry.AddrIPv4[0]
	port := entry.Port

	urlPath := fmt.Sprintf("http://%v.%v.%v.%v:%d/%s", addr[0], addr[1], addr[2], addr[3], port, name)
	fmt.Println(urlPath)
	if resp, err := http.Get(urlPath); err != nil {
		panic(err)
	} else if resp.StatusCode != 200 {
		fmt.Println("Not correct file/code")
		os.Exit(1)
	} else {
		if out == "" {
			rgx := regexp.MustCompile("inline; filename=\"(.*)\"")
			out = rgx.FindStringSubmatch(resp.Header.Get("Content-Disposition"))[1]
		}
		if of, err := os.Create(out); err != nil {
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
