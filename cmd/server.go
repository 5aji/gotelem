package cmd

import (
	"fmt"
	"net"
	"time"

	"github.com/kschamplin/gotelem/internal/gotelem"
	"github.com/tinylib/msgp/msgp"
	"github.com/urfave/cli/v2"
)

const xbeeCategory = "XBee settings"

var serveCmd = &cli.Command{
	Name:    "serve",
	Aliases: []string{"server", "s"},
	Usage:   "Start a telemetry server",
	Flags: []cli.Flag{
		&cli.BoolFlag{Name: "xbee", Aliases: []string{"x"}, Usage: "Find and connect to an XBee"},
	},
	Action: func(ctx *cli.Context) error {
		serve()
		return nil
	},
}


type session struct {
	conn net.Conn
	send chan gotelem.Body
	recv chan gotelem.Body
	quit chan bool
}

func serve() {
	ln, err := net.Listen("tcp", ":8082")
	if err != nil {
		fmt.Printf("Error listening: %v\n", err)
	}
	fmt.Printf("Listening on :8082\n")
	
	for {
		conn, err := ln.Accept()
		if err != nil {
			fmt.Printf("error accepting: %v\n", err)
		}
		go handleCon(conn)
	}
}

func handleCon(conn net.Conn) {
	//	reader := msgp.NewReader(conn)
	writer := msgp.NewWriter(conn)
	for {
		// data := telemnet.StatusBody{
		// 	BatteryPct: 1.2,
		// 	ErrCode: 0,
		// }
		// data.EncodeMsg(writer)
		data := gotelem.StatusBody{
			BatteryPct: 1.2,
			ErrCode:    0,
		}
		data.EncodeMsg(writer)
		writer.Flush()
		time.Sleep(1 * time.Second)
	}
}
