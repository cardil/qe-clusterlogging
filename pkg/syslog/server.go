package syslog

import (
	"fmt"
	"log/slog"
	"net"
	"os"
	"strconv"

	pkgserver "github.com/cardil/qe-clusterlogging/pkg/server"
	"gopkg.in/mcuadros/go-syslog.v2"
)

type Handler func(syslog.LogPartsChannel)

func Serve(handle Handler) pkgserver.Server {
	channel := make(syslog.LogPartsChannel)
	handler := syslog.NewChannelHandler(channel)

	server := syslog.NewServer()
	server.SetFormat(syslog.RFC5424)
	server.SetHandler(handler)
	port := 514
	if !canBind(port) {
		port = 8514
	}
	if eport, set := os.LookupEnv("PORT"); set {
		iport, perr := strconv.Atoi(eport)
		if perr == nil {
			port = iport
		}
	}
	bind := fmt.Sprint("0.0.0.0:", port)
	if err := server.ListenUDP(bind); err != nil {
		return &pkgserver.ErrorServer{Error: err}
	}
	if err := server.ListenTCP(bind); err != nil {
		return &pkgserver.ErrorServer{Error: err}
	}

	if err := server.Boot(); err != nil {
		return &pkgserver.ErrorServer{Error: err}
	}

	go handle(channel)

	slog.Info("Started Syslog server", "port", port)
	return &syslogServer{server}
}

func canBind(port int) bool {
	ln, err := net.Listen("tcp", fmt.Sprintf(":%d", port))

	if err != nil {
		return false
	}

	_ = ln.Close()
	return true
}

type syslogServer struct {
	server *syslog.Server
}

func (s *syslogServer) Run() error {
	s.server.Wait()
	return nil
}

func (s *syslogServer) Kill() error {
	return s.server.Kill()
}
