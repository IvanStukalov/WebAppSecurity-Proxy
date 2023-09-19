package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
)

var transport = http.DefaultTransport

func logRequest(r *http.Request) {
	log.Println("-------Request------")
	log.Println(r.Method, r.RequestURI, r.Proto)
	log.Println("Host:", r.Host)
	log.Println("User-Agent:", r.UserAgent())
	log.Println("Accept:", r.Header.Get("Accept"))
	log.Println("Proxy-Connection:", r.Header.Get("Proxy-Connection"))
	fmt.Println()
}

func logResponse(response *http.Response) {
	log.Println("-------Response------")
	log.Println(response.Proto, response.Status)
	log.Println("Server: ", response.Header.Get("Server"))
	log.Println("Date: ", response.Header.Get("Date"))
	log.Println("Content-Type: ", response.Header.Get("Content-Type"))
	log.Println("Content-Length: ", response.Header.Get("Content-Length"))
	log.Println("Connection: ", response.Header.Get("Connection"))
	log.Println("Location: ", response.Header.Get("Location"))
	fmt.Println()

	body, err := io.ReadAll(response.Body)
	if err != nil {
		fmt.Println("cant read response body")
	} else {
		fmt.Println(string(body))
		fmt.Println()
	}
}

func copyReqHeaders(src *http.Request, dst *http.Request) {
	for name, values := range src.Header {
		for _, value := range values {
			dst.Header.Add(name, value)
		}
	}
}

func copyResHeaders(src *http.Response, w http.ResponseWriter) {
	for name, values := range src.Header {
		for _, value := range values {
			w.Header().Add(name, value)
		}
	}

	// set the status code from proxy response to src response
	w.WriteHeader(src.StatusCode)
}

// --------------------------------------------------------------------------------//
func handleRequest(w http.ResponseWriter, r *http.Request) {
	// 3. удалить заголовок Proxy-Connection
	r.Header.Del("Proxy-Connection")

	// 2. заменить путь на относительный
	r.RequestURI = r.URL.Path

	// log request
	logRequest(r)

	// create proxy request with the same method, url and body
	proxyReq, err := http.NewRequest(r.Method, r.URL.String(), r.Body)
	if err != nil {
		http.Error(w, "cant create proxy request", http.StatusInternalServerError)
		return
	}

	// copy headers from src request to proxy request
	copyReqHeaders(r, proxyReq)

	// send proxy request using transport
	response, err := transport.RoundTrip(proxyReq)
	if err != nil {
		http.Error(w, "error sending proxy request", http.StatusInternalServerError)
		return
	}
	defer response.Body.Close()

	// log response
	logResponse(response)

	// copy headers from proxy response to src response
	copyResHeaders(response, w)

	// copy body from proxy response to src response
	io.Copy(w, response.Body)
}

//------------------------------------------------------------------------//

func main() {
	// creating server with handleRequest func as a Handler
	server := http.Server{
		Addr:    ":8080",
		Handler: http.HandlerFunc(handleRequest),
	}

	// starting server
	log.Println("starting server on :8080")
	err := server.ListenAndServe()
	if err != nil {
		log.Fatalln("error starting proxy server: ", err)
	}
}
