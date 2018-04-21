package main

import (
	pb "github.com/flaboy/svagent/proto"
	"github.com/urfave/cli"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/keepalive"
	"log"
	"net"
	"time"
)

func agent_start(ctx *cli.Context) (err error) {
	to_addr := ctx.String("local")
	sv_addr := ctx.String("remote")
	for {
		agent_loop(ctx, to_addr, sv_addr)
		time.Sleep(time.Second)
	}
	return
}

func agent_loop(ctx *cli.Context, to_addr, sv_addr string) {
	opts := []grpc.DialOption{
		grpc.WithInsecure(),
		grpc.WithBackoffConfig(
			grpc.BackoffConfig{MaxDelay: time.Second * 3},
		), grpc.WithKeepaliveParams(
			keepalive.ClientParameters{
				Time:                time.Second * 30,
				Timeout:             time.Second * 3,
				PermitWithoutStream: true,
			},
		),
	}

	conn, err := grpc.Dial(sv_addr, opts...)
	if err != nil {
		return
	}
	defer conn.Close()

	c := pb.NewAgentClient(conn)
	// ar := &AgentRequest{}
	ar, err := c.Register(context.Background())
	if err != nil {
		log.Println(err)
		return
	}

	// err = ar.Send(&pb.Frame{1, 0, []byte("aaa")})
	bridge := &Birdge{Agent_RegisterClient: ar, to_addr: to_addr}
	bridge.mapper = make(map[int64]net.Conn)
	for {
		frame, err := ar.Recv()
		if err != nil {
			log.Println(err)
			return
		}
		bridge.Handle(frame)
	}
	// ar.Recv() (*AgentResponse, error)
}

type Birdge struct {
	pb.Agent_RegisterClient
	mapper  map[int64]net.Conn
	to_addr string
}

func (me *Birdge) Handle(frame *pb.Frame) {
	var err error
	conn, ok := me.mapper[frame.Channel]

	if !ok {
		log.Println("conn", frame.Channel)
		dailer := &net.Dialer{}
		conn, err = dailer.Dial("tcp4", me.to_addr)
		if err != nil {
			me.Send(&pb.Frame{Channel: frame.Channel, Flag: pb.Frame_Close})
			log.Println("ERROR", err)
			return
		}
		me.mapper[frame.Channel] = conn

		go func() {
			b := make([]byte, 4096)
			for {
				n, err := conn.Read(b)
				if err != nil {
					// log.Println(err)
					me.Send(&pb.Frame{Channel: frame.Channel, Flag: pb.Frame_Close})
					me.Handle(&pb.Frame{Channel: frame.Channel, Flag: pb.Frame_Close})
					return
				} else {
					me.Send(&pb.Frame{Channel: frame.Channel, Flag: pb.Frame_Data, Body: b[0:n]})
				}
			}
		}()
	}

	switch frame.Flag {
	case pb.Frame_Close:
		conn.Close()
	case pb.Frame_Data:
		conn.Write(frame.Body)
	}
}
