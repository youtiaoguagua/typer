package server

import (
	"context"
	"fmt"
	"github.com/charmbracelet/ssh"
	"github.com/charmbracelet/wish"
	"golang.org/x/exp/slog"
	"os"
	"os/signal"
	"syscall"
	"time"
)

type CommandInfo struct {
	Host          string
	Port          int
	ServerKeyPath string
}

type Server struct {
	host    string
	port    int
	srv     *ssh.Server
	rooms   map[string]*Room
	players map[string]*Player
}

func NewServer(commandInfo CommandInfo) (*Server, error) {
	s := &Server{
		host:    commandInfo.Host,
		port:    commandInfo.Port,
		rooms:   make(map[string]*Room),
		players: make(map[string]*Player),
	}
	ws, err := wish.NewServer(
		ssh.PasswordAuth(passwordHandler),
		ssh.PublicKeyAuth(publicKeyHandler),
		wish.WithHostKeyPath(commandInfo.ServerKeyPath),
		wish.WithAddress(fmt.Sprintf("%s:%d", commandInfo.Host, commandInfo.Port)),
		wish.WithIdleTimeout(30*time.Second),
		wish.WithMaxTimeout(10*time.Minute),
		wish.WithMiddleware(
			TyperMiddleware(s),
			//wrecover.Middleware(TyperMiddleware(s)),
		),
	)
	if err != nil {
		return nil, err
	}
	s.srv = ws
	return s, nil
}

func Start(commandInfo CommandInfo) {
	s, err := NewServer(commandInfo)

	if err != nil {
		panic(err)
		//slog.Error("start server err", err)
	}

	done := make(chan os.Signal, 1)
	signal.Notify(done, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)

	slog.Info(fmt.Sprintf("Starting SSH server on %s:%d", commandInfo.Host, commandInfo.Port))

	go func() {
		if err = s.srv.ListenAndServe(); err != nil {
			slog.Error("start server err", err)
			return
		}
	}()

	<-done
	slog.Info("Stopping SSH server")
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer func() { cancel() }()
	if err := s.srv.Shutdown(ctx); err != nil {
		slog.Error("stop server err", err)
	}
}

func passwordHandler(ctx ssh.Context, password string) bool {
	return true
}

func publicKeyHandler(ctx ssh.Context, key ssh.PublicKey) bool {
	return true
}

func (srv *Server) findPlayer(user string) *Player {
	v, ok := srv.players[user]
	if !ok {
		return nil
	}
	return v
}

func (srv *Server) deleteRoom(id string) {
	delete(srv.rooms, id)
}
