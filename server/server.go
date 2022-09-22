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

	login := false
	var handshakePack *HandshakeClient

	defer conn.Close()
	for {
		reader, err := parse.MsgBody(reader)
		if err != nil {
			if err == io.ErrUnexpectedEOF || err == io.EOF {
				break
			}

		}

		packageId, err := parse.ReadPackId(reader)
		if err != nil {
			panic(err)
		}
		fmt.Println("Package id:", packageId)
		switch packageId {

		case 0:
			if login {
				loginPack := &LoginClient{}
				err = parse.UnPack(reader, loginPack)
				if err != nil {
					if err == io.ErrUnexpectedEOF || err == io.EOF {
						break
					}

				}

				fmt.Println(loginPack.Name)

				/*
				if loginPack.Name != "Rorical" {
					disconnectPack := &DisconnectServer{
						Reason: `{"text": ""}`,
					}
					disconnectpacked, err := parse.Pack(disconnectPack)
					if err != nil {
						panic(err)
					}
					response, err := parse.ConstructPack(disconnectpacked, 0)
					if err != nil {
						panic(err)
					}
					conn.Write(response)
					return
				}
				*/

				server, err := net.Dial("tcp", "172.65.211.101:25565") //target server ip
				if err != nil {
					panic(err)
				}
				serverhandpacked, err := parse.Pack(handshakePack)
				if err != nil {
					panic(err)
				}
				response, err := parse.ConstructPack(serverhandpacked, 0)
				if err != nil {
					panic(err)
				}
				server.Write(response)

				loginpacked, err := parse.Pack(loginPack)
				if err != nil {
					panic(err)
				}
				response, err = parse.ConstructPack(loginpacked, 0)
				if err != nil {
					panic(err)
				}
				server.Write(response)

				defer server.Close()
				pro := proxy.New(conn, server)
				pro.Start()
				return

			} else {
				clienthand := &HandshakeClient{}
				err = parse.UnPack(reader, clienthand)
				if err != nil {
					if err == io.ErrUnexpectedEOF || err == io.EOF {
						break
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
					clienthand.Port = 25565
					clienthand.Address = "mc.hypixel.net"
					handshakePack = clienthand

					login = true
				}
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
