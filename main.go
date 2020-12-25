package main

import (
	"context"
	"crypto/tls"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"regexp"
	"strconv"
	"sync"
	"time"
)

var dialer = &net.Dialer{
	Timeout:   30 * time.Second,
	KeepAlive: 30 * time.Second,
	DualStack: true,
}

var path string
var uri string
var fingerprint string
var thread int

func main() {
	stopChan := make(chan os.Signal, 1)
	signal.Notify(stopChan, os.Interrupt)
	go listenForInterrupt(stopChan)

	flag.StringVar(&path, "path", "ips.txt", "File path")
	flag.StringVar(&uri, "uri", "https://example.org/", "Uri to search")
	flag.StringVar(&fingerprint, "fp", "example", "Fingerprint regex")
	flag.IntVar(&thread, "thread", 200, "Tread count")

	flag.Parse()

	parsedIps := ParseIPAddresses(path)
	dataLength := len(parsedIps)

	log.Println("Total IP Address: " + strconv.Itoa(dataLength))

	var wg sync.WaitGroup
	wg.Add(dataLength)

	ch := make(chan bool, thread)

	done := 0.0
	last := 0.0

	for i, xi := range parsedIps {
		ch <- true
		calc := done / float64(dataLength)
		if calc-last > 10 {
			log.Println(calc, "p.")
			last = calc
		}

		go func(i int, xi string, chb chan bool, wg *sync.WaitGroup) {
			defer wg.Done()
			doRequest(uri, xi, fingerprint)
			done++

			<-chb
		}(i, xi, ch, &wg)
	}

	wg.Wait()

	log.Println("Done!!")
}

func listenForInterrupt(stopScan chan os.Signal) {
	<-stopScan
	fmt.Println("halting scan")
	os.Exit(1)
}

func doRequest(addr string, origin string, fp string) {
	client := &http.Client{
		Timeout: time.Second * 10,
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
			DialContext: func(ctx context.Context, network, addr string) (net.Conn, error) {
				//TODO: Check better solutions for dialcontext like timeouts.
				uri, ferr := url.Parse(addr)

				opaq := "80"
				if ferr != nil {
					log.Println(ferr, addr)
				} else {
					opaq = uri.Opaque
				}

				addr = origin + ":" + opaq

				return dialer.DialContext(ctx, network, addr)
			},
		},
	}

	res, herr := client.Get(addr)

	if herr == nil {
		defer res.Body.Close()
	}

	if res != nil && res.StatusCode < 400 {

		if fp != "" {
			bodyBytes, berr := ioutil.ReadAll(res.Body)

			if berr != nil {
				log.Println(berr, strconv.Itoa(res.StatusCode)+" "+origin)
				return
			}

			re := regexp.MustCompile(fp)

			if re.Match(bodyBytes) {
				log.Println("FINGERPRINT MATCH: " + strconv.Itoa(res.StatusCode) + " " + origin)
				return
			}
		}

		//log.Println("Status Code is: " + strconv.Itoa(res.StatusCode) + " " + origin)
	}
}
