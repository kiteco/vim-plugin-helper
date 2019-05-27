// USAGE
//
//   kite-http [--header name=value] [--debug] [-] <url>
//
// Use a hyphen to read data from stdin and POST it.
//
// Sets a default content-type of "application/x-www-form-urlencoded" for POSTs.
//
// Writes the http response (including header's protocol and status) to stdout.
//
// Use the --debug flag to write data from stdin to stderr.

package main

import (
	"flag"
	"fmt"
	"io"
	"io/ioutil"
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

	debug := flag.Bool("debug", false, "write data from stdin to stderr")
	timeout := flag.Duration("timeout", time.Second, "timeout for receiving a response from the HTTP server")
	flag.Parse()

	var stdin = false
	var url string

	switch flag.NArg() {
	case 0:
		log.Fatal("missing url")
	case 1:
		url = flag.Arg(0)
	case 2:
		if flag.Arg(0) == "-" {
			stdin = true
		} else {
			log.Fatal("unrecognised argument: " + flag.Arg(0))
		}
		url = flag.Arg(1)
	}

	var data []byte
	var err error

	if stdin {
		data, err = ioutil.ReadAll(os.Stdin)
		if err != nil {
			log.Fatal(err)
		}
	}

	if *debug {
		io.WriteString(os.Stderr, string(data)+"\n")
	}

	var req *http.Request
	var resp *http.Response

	client := &http.Client{
		Timeout: *timeout,
	}

	if len(string(data)) > 0 {
		req, err = http.NewRequest("POST", url, strings.NewReader(string(data)))
	} else {
		req, err = http.NewRequest("GET", url, nil)
	}

	for k, v := range myHeaders {
		req.Header.Set(k, v)
	}

	if len(string(data)) > 0 {
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
