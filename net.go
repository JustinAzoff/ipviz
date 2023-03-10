package main

import (
	"bufio"
	"bytes"
	"encoding/binary"
	"fmt"
	"log"
	"net"

	"github.com/buger/jsonparser"
)

type IPRecord struct {
	ip   string
	orig bool
}

func ip2Long(ip string) uint32 {
	var long uint32
	binary.Read(bytes.NewBuffer(net.ParseIP(ip).To4()), binary.BigEndian, &long)
	return long
}

func int2ip(nn uint32) net.IP {
	ip := make(net.IP, 4)
	binary.BigEndian.PutUint32(ip, nn)
	return ip
}

const bufferSize int = 128 * 1024

func listen(addr string, port int, lineChan chan IPRecord) {
	bind := fmt.Sprintf("%s:%d", addr, port)
	log.Printf("Listening on %s", bind)
	l, err := net.Listen("tcp", bind)
	if err != nil {
		log.Fatalf("Error listening: %v", err)
	}
	defer l.Close()
	for {
		conn, err := l.Accept()
		if err != nil {
			log.Fatalf("Error accepting: %v", err)
		}
		log.Printf("New connection from %s", conn.RemoteAddr())
		go handleLog(conn, lineChan)
	}
}

func handleLog(conn net.Conn, lineChan chan IPRecord) {
	defer conn.Close()
	scanner := bufio.NewScanner(conn)
	buf := make([]byte, 0, bufferSize)
	scanner.Buffer(buf, 1024*1024)
	for scanner.Scan() {
		value, err := jsonparser.GetString(scanner.Bytes(), "id.orig_h")
		if err != nil {
			log.Printf("Error json %v", err)
			continue
		}
		lineChan <- IPRecord{ip: value, orig: true}
		value, err = jsonparser.GetString(scanner.Bytes(), "id.resp_h")
		if err != nil {
			log.Printf("Error json %v", err)
			continue
		}
		lineChan <- IPRecord{ip: value, orig: false}
	}
	if err := scanner.Err(); err != nil {
		log.Printf("Error reading from connection: %v", err)
	}
	log.Printf("Connection from %s closed", conn.RemoteAddr())
}
