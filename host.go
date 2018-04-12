package main

import (
	"container/list"
	"errors"
	pb "github.com/flaboy/svagent/proto"
	"github.com/urfave/cli"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
	"log"
	"math"
	"net"
	"os"
	"sync/atomic"
)

func host_start(ctx *cli.Context) (err error) {
	start_service()
	return
}

func start_service() {
	sv_addr := "0.0.0.0:6667"
	lis, err := net.Listen("tcp", sv_addr)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	gs := grpc.NewServer()
	sv := &Service{}

	sv.Init()
	go sv.Start()

	pb.RegisterAgentServer(gs, sv)
	reflection.Register(gs)
	log.Println("start server in " + sv_addr)
	if err := gs.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}

type Service struct {
	agents         *list.List
	channel_id_seq int64
	agents_map     map[int64]*Bind
}

type Bind struct {
	agent *Agent
	ch    chan *pb.Frame
}

func (me *Service) Start() {
	l, err := net.Listen("tcp4", "0.0.0.0:5555")
	if err != nil {
		log.Println("ERROR", l)
		os.Exit(1)
	}

	for {
		c, err := l.Accept()
		go me.HandleConn(c)

		if err != nil {
			log.Println("NOTICE", err)
		}
	}
}

func (me *Service) Init() {
	me.agents = list.New()
	me.agents_map = make(map[int64]*Bind)
}

func (me *Service) Register(rs pb.Agent_RegisterServer) (err error) {
	// el := &list.Element{Value: &Agent{rs}}
	el := me.agents.PushBack(&Agent{rs})
	defer me.agents.Remove(el)

	for {
		frame, err := rs.Recv()
		if err != nil {
			log.Println("NOTICE", err)
			return nil
		}

		me.Dispatch(frame)
	}
	return
}

func (me *Service) Dispatch(frame *pb.Frame) {
	bind, ok := me.agents_map[frame.Channel]
	if !ok {
		return
	}
	bind.ch <- frame
}

func (me *Service) HandleConn(c net.Conn) {
	log.Println("handle connection from ", c.RemoteAddr())
	var channel_id = atomic.AddInt64(&me.channel_id_seq, 1)
	bind, err := me.pickAgent(channel_id)
	if err != nil {
		c.Close()
		log.Println(err)
		return
	}

	ch_upstream := make(chan []byte)
	ch_error := make(chan error)

	go func() {
		buf := make([]byte, 4096)
		for {
			n, err := c.Read(buf)
			if err != nil {
				// log.Println(err)
				ch_error <- err
			} else {
				ch_upstream <- buf[0:n]
			}
		}
	}()

	for {
		select {
		case frame := <-bind.ch:
			switch frame.Flag {
			case pb.Frame_Close:
				c.Close()
			case pb.Frame_Data:
				c.Write(frame.Body)
			}
		case buf := <-ch_upstream:
			bind.agent.Send(&pb.Frame{channel_id, pb.Frame_Data, buf})
		case <-ch_error:
			// log.Println(err)
			bind.agent.Send(&pb.Frame{channel_id, pb.Frame_Close, nil})
			return
		}

	}
}

func (me *Service) pickAgent(channel_id int64) (bind *Bind, err error) {
	if me.agents.Len() == 0 {
		return nil, errors.New("no backend")
	}
	n := int(math.Mod(float64(channel_id), float64(me.agents.Len())))
	el := me.agents.Front()
	for i := 0; i < n; i++ {
		el = el.Next()
	}

	bind = &Bind{agent: (el.Value).(*Agent), ch: make(chan *pb.Frame)}
	me.agents_map[channel_id] = bind
	return
}

type Agent struct {
	pb.Agent_RegisterServer
}
