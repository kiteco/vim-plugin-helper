// USAGE
//
//   kite-http [--post --data DATA [--debug]] <url>
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

func main() {
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

	var resp *http.Response
	var err error

	client := &http.Client{
		Timeout: *timeout,
	}

	if *isPost {
		resp, err = client.Post(url, "application/x-www-form-urlencoded", strings.NewReader(*data)) // match curl's content-type
	} else {
		resp, err = client.Get(url)
	}

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
