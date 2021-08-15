package server

import (
	"bufio"
	"fmt"
	"io"
	"net"

	"github.com/Rorical/McSpeed/parse"
	"github.com/Rorical/McSpeed/proxy"
)

func handler(conn net.Conn) {
	reader := bufio.NewReader(conn)

	defer conn.Close()
	for {
		reader, err := parse.MsgBody(reader)
		if err != nil {
			if err == io.ErrUnexpectedEOF || err == io.EOF {
				return
			}
			panic(err)
		}

		packageId, err := parse.ReadPackId(reader)
		fmt.Println("Package id:", packageId)
		switch packageId {
		case 0:
			clienthand := &HandshakeClient{}
			err = parse.UnPack(reader, clienthand)
			if err != nil {
				if err == io.ErrUnexpectedEOF || err == io.EOF {
					return
				}
				panic(err)
			}
			switch clienthand.State {
			case 1:
				serverhand := &HandshakeServer{
					Json: `{"version":{"name":"1.8.9","protocol":47},"players":{"max":200000,"online":128279,"sample":[]},"description":"Hypixel加速服务器","favicon":"data:image/gif;base64:0;"}`,
				}
				serverhandpacked, err := parse.Pack(serverhand)
				if err != nil {
					panic(err)
				}
				response, err := parse.ConstructPack(serverhandpacked, 0)
				if err != nil {
					panic(err)
				}
				conn.Write(response)
			case 2:
				server, err := net.Dial("tcp", "172.65.211.101:25565")
				if err != nil {
					panic(err)
				}
				clienthand.Address = "mc.hypixel.net"
				serverhandpacked, err := parse.Pack(clienthand)
				if err != nil {
					panic(err)
				}
				response, err := parse.ConstructPack(serverhandpacked, 0)
				if err != nil {
					panic(err)
				}
				server.Write(response)
				pro := proxy.New(conn, server)
				pro.Start()
				return
			}

		case 1:
			clientping := &PingClient{}
			err = parse.UnPack(reader, clientping)
			if err != nil {
				panic(err)
			}
			fmt.Println(clientping)
			serverpong, err := parse.Pack(clientping)
			if err != nil {
				panic(err)
			}
			response, err := parse.ConstructPack(serverpong, 1)
			if err != nil {
				panic(err)
			}
			conn.Write(response)
		}
	}

	return
}

func Loop() error {
	l, err := net.Listen("tcp", ":25565")
	if err != nil {
		return err
	}
	for {
		c, err := l.Accept()
		if err != nil {
			break
		}
		go handler(c)
	}
	return nil
}
