package main

import (
	"bufio"
	"encoding/json"
	"errors"
	"html/template"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strings"
)

type etcdresponse struct {
	Node etcdNode
}

type etcdNode struct {
	Nodes []node
}

type node struct {
	Value string
	Key   string
}

type server struct {
	Host    string
	Service string
}

func main() {
	var err error

	etcdEndpointsUrl := os.Getenv("ETCD_ENDPOINTS_URL")
	if etcdEndpointsUrl == "" {
		log.Fatal(errors.New("Sorry, you must specify the ETCD_ENDPOINTS_URL env var!"))
	}
	r, err := http.Get(etcdEndpointsUrl)
	if err != nil {
		log.Fatal(err)
	}
	defer r.Body.Close()
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		log.Fatal(err)
	}
	var n etcdresponse
	json.Unmarshal(body, &n)
	log.Println(n)
	locations := make([]server, 0)
	for _, endpoint := range n.Node.Nodes {
		keyparts := strings.Split(endpoint.Key, "/")
		log.Println(keyparts)
		s := server{
			Host:    endpoint.Value,
			Service: keyparts[len(keyparts)-1],
		}
		locations = append(locations, s)
	}
	tmpl, err := template.ParseFiles("nginx.template")
	if err != nil {
		log.Fatal(err)
	}
	// open output file
	fo, err := os.Create("/usr/share/nginx/autoproxy.conf")
	if err != nil {
		panic(err)
	}
	// close fo on exit and check for its returned error
	defer func() {
		if err := fo.Close(); err != nil {
			panic(err)
		}
	}()
	// make a write buffer
	w := bufio.NewWriter(fo)
	err = tmpl.Execute(w, locations)
	if err != nil {
		log.Fatal(err)
	}
	if err = w.Flush(); err != nil {
		panic(err)
	}
}
