// USAGE
//
//   kite-http [--header name=value] [--post --data DATA [--debug]] <url>
//
// Sets a default content-type of "application/x-www-form-urlencoded" for POSTs.
//
// Writes the http response (including header's protocol and status) to stdout.
//
// Use the --debug flag to write DATA to stderr.

package main

import (
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
	"time"
)

type headers map[string]string

var myHeaders headers

func (i *headers) String() string {
	return fmt.Sprint(*i)
}

func (i *headers) Set(value string) error {
	header := strings.SplitN(value, "=", 2)
	myHeaders[header[0]] = header[1]
	return nil
}

func main() {
	myHeaders = make(map[string]string)
	flag.Var(&myHeaders, "header", "HTTP header as name=value")

	debug := flag.Bool("debug", false, "write --data arg to stderr")
	isPost := flag.Bool("post", false, "POST request")
	data := flag.String("data", "", "data to send to the HTTP server")
	timeout := flag.Duration("timeout", time.Second, "timeout for receiving a response from the HTTP server")
	flag.Parse()

	url := flag.Arg(0)

	if url == "" {
		log.Fatal("missing url")
	}

	if *debug {
		io.WriteString(os.Stderr, *data+"\n")
	}

	var req *http.Request
	var resp *http.Response
	var err error

	client := &http.Client{
		Timeout: *timeout,
	}

	if *isPost {
		req, err = http.NewRequest("POST", url, strings.NewReader(*data))
	} else {
		req, err = http.NewRequest("GET", url, nil)
	}

	for k, v := range myHeaders {
		req.Header.Set(k, v)
	}

	if *isPost {
		_, ok := myHeaders["Content-Type"]
		if !ok {
			req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		}
	}

	resp, err = client.Do(req)

	if err != nil {
		log.Fatal(err)
	}
	defer resp.Body.Close()

	// We want to write to stdout the dechunked response.  However
	// httputil.DumpResponse() writes the chunked response.
	//
	// So we write the status part of the header, which is the only
	// part we care about, ourselves and then the body.

	fmt.Print(resp.Proto + " " + resp.Status + "\r\n")
	fmt.Print("\r\n")
	io.Copy(os.Stdout, resp.Body)
}
