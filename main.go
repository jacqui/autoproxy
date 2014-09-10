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
	log.Println("Parsed nginx template:", tmpl)
	// open output file
	fo, err := os.Create("/etc/nginx/nginx.conf.tmp")
	if err != nil {
		return err
	}
	log.Println("Created temporary nginx config file:", fo)
	// close fo on exit and check for its returned error
	defer fo.Close()

	// make a write buffer
	w := bufio.NewWriter(fo)
	log.Println("Made a write buffer", w)

	err = tmpl.Execute(w, endpoints)
	log.Println("executed template variables")
	if err != nil {
		return err
	}
	if err = w.Flush(); err != nil {
		return err
	}
	log.Println("done, returning")
	return nil
}

func main() {
	var err error
	client := etcd.NewClient([]string{"http://172.17.42.1:4001"})
	resp, err := client.Get("endpoints", false, false)
	if err != nil {
		log.Fatal(err)
	}

	locations := findEndpoints(resp.Node.Nodes)
	err = writeNginxConfig(locations)
	if err != nil {
		log.Fatal(err)
	}

	watchChan := make(chan *etcd.Response, 10)
	stopChan := make(chan bool, 1)

	go client.Watch("/endpoints", 0, false, watchChan, stopChan)
	log.Println("Waiting for an update...")

	for {
		select {
		case <- stopChan:
			log.Println("Stop channel hit")
		case r := <- watchChan:
			log.Println("Got updated endpoints:")
			locations = findEndpoints(r.Node.Nodes)
			err = writeNginxConfig(locations)
			if err != nil {
				log.Fatal(err)
			}
		}
	}
}
