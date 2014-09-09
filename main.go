package main

import (
	"bufio"
	"html/template"
	"log"
	"os"
	"strings"

	"github.com/coreos/go-etcd/etcd"
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

func findEndpoints(nodes etcd.Nodes) []server {
	endpoints := make([]server, 0)

	log.Println("Current endpoints:")
	for _, n := range nodes {
		log.Printf("%s: %s\n", n.Key, n.Value)
		keyparts := strings.Split(n.Key, "/")
		log.Println(keyparts)

		s := server{
			Host:    n.Value,
			Service: keyparts[len(keyparts)-1],
		}
		endpoints = append(endpoints, s)
	}
	return endpoints
}

func writeNginxConfig(endpoints []server) error {
	tmpl, err := template.ParseFiles("nginx.template")
	if err != nil {
		return err
	}
	// open output file
	fo, err := os.Create("/etc/nginx/nginx.conf.tmp")
	if err != nil {
		return err
	}
	// close fo on exit and check for its returned error
	defer fo.Close()

	// make a write buffer
	w := bufio.NewWriter(fo)
	err = tmpl.Execute(w, endpoints)
	if err != nil {
		return err
	}
	if err = w.Flush(); err != nil {
		return err
	}
	return nil
}

func main() {
	var err error
	client := etcd.NewClient([]string{"http://172.17.42.1:4001"})
	resp, err := client.Get("endpoints", true, true)
	if err != nil {
		log.Fatal(err)
	}

	locations := findEndpoints(resp.Node.Nodes)
	err = writeNginxConfig(locations)
	if err != nil {
		log.Fatal(err)
	}

	watchChan := make(chan *etcd.Response)
	go client.Watch("/endpoints", 0, false, watchChan, nil)
	log.Println("Waiting for an update...")

	r := <-watchChan

	log.Printf("Got updated endpoints:")

	locations = findEndpoints(r.Node.Nodes)

	err = writeNginxConfig(locations)
	if err != nil {
		log.Fatal(err)
	}
}
