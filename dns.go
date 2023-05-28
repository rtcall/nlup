package main

import (
	"bufio"
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"net"
	"os"
	"strings"
	"time"
)

const (
	ARecord = 0x0001
	INClass = 0x0001
)

const DNSPort = 53
const Timeout = 2

type dnsConn struct {
	net.Conn
}

type header struct {
	Id, Flags uint16
	Qdcount   uint16
	Ancount   uint16
	Nscount   uint16
	Arcount   uint16
}

type answer struct {
	Ntype   byte
	Ptr     uint16
	Qtype   uint16
	Class   uint16
	Ttl     uint32
	Addrlen uint16
	Addr    uint32
}

func findNsAddr() (string, error) {
	conf := []string{
		"/etc/resolv.conf",
	}

	for _, i := range conf {
		f, err := os.Open(i)

		if err != nil {
			return "", err
		}

		defer f.Close()
		scanner := bufio.NewScanner(f)

		for scanner.Scan() {
			s := scanner.Text()

			if s[0] == '#' {
				continue
			}

			args := strings.Split(s, " ")
			if args[0] == "nameserver" && len(args) > 1 {
				return args[1], nil
			}
		}
	}

	return "", errors.New("nameserver not found")
}

func read(buf *bytes.Buffer, data any) error {
	return binary.Read(buf, binary.BigEndian, data)
}

func write(buf *bytes.Buffer, data any) error {
	return binary.Write(buf, binary.BigEndian, data)
}

func writeName(buf *bytes.Buffer, name string) (n int) {
	for _, i := range strings.Split(name, ".") {
		buf.WriteByte(byte(len(i)))
		n++
		buf.WriteString(i)
		n += len(i)
	}

	buf.WriteByte(0)
	return
}

func dial(server string) (dnsConn, error) {
	c, err := net.Dial("udp", fmt.Sprintf("%s:%d", server, DNSPort))

	if err == nil {
		c.SetDeadline(time.Now().Add(Timeout * time.Second))
	}

	return dnsConn{c}, err
}

func (c dnsConn) sendQuery(name string) (answer, error) {
	buf := &bytes.Buffer{}
	ah := &header{}
	a := &answer{}

	h := header{
		0x4c3f,
		1 << 8, // recursion
		0x0001, // single question
		0x0000,
		0x0000,
		0x0000,
	}

	write(buf, h)
	nc := writeName(buf, name)
	write(buf, []uint16{ARecord, INClass})
	if _, err := c.Write(buf.Bytes()); err != nil {
		return *a, err
	}

	b := make([]byte, 64)
	if _, err := c.Read(b); err != nil {
		return *a, err
	}

	abuf := bytes.NewBuffer(b)
	read(abuf, ah)
	abuf.Next(nc + 4)
	read(abuf, a)

	if a.Addrlen != 4 {
		return *a, errors.New("invalid name")
	}

	return *a, nil
}
