package main

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strconv"
)

type Proxy struct {
	target *url.URL
	proxy  *httputil.ReverseProxy
}

func (p *Proxy) handle(w http.ResponseWriter, r *http.Request) {
	p.proxy.ServeHTTP(w, r)

}

func NewProxy(target string) *Proxy {

	url, err := url.Parse(target)
	if err != nil {
		panic(err)
	}

	rp := httputil.NewSingleHostReverseProxy(url)
	originalDirector := rp.Director

	rp.Director = func(req *http.Request) {
		originalDirector(req)
		modifyRequest(req)
	}

	rp.ModifyResponse = modifyResponse()

	return &Proxy{target: url, proxy: rp}
}

func modifyRequest(req *http.Request) {
	err := req.ParseForm()
	if err != nil {
		panic(err)
	}

	//To add query in the http request
	form, err := url.ParseQuery(req.URL.RawQuery)
	if err != nil {
		panic(err)
	}

	form.Add("user_id", "user_id_value")
	req.URL.RawQuery = form.Encode()

}

func modifyResponse() func(resp *http.Response) (err error) {
	return func(resp *http.Response) (err error) {
		//delete headers from the response that are not required for the client
		//resp.Header.Del(XAuthState)
		log.Println("in the modify response")

		b, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			panic(err)
		}

		log.Println(string(b))

		err = resp.Body.Close()
		if err != nil {
			panic(err)
		}

		b, err = json.Marshal("resp from reverse proxy")
		if err != nil {
			panic(err)
		}

		body := ioutil.NopCloser(bytes.NewReader(b))
		resp.Body = body
		resp.ContentLength = int64(len(b))
		resp.Header.Set("Content-Length", strconv.Itoa(len(b)))

		return
	}
}

func main() {
	port := ":9091"

	redirectURL := "http://localhost:9090"
	proxy := NewProxy(redirectURL)

	http.HandleFunc("/user/greet", proxy.handle)
	http.HandleFunc("/user/welcome", proxy.handle)

	log.Printf("reverse proxy server running on port no : %v", port)
	log.Fatal(http.ListenAndServe(port, nil))
}
