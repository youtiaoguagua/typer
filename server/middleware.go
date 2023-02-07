package server

import (
	"fmt"
	"github.com/charmbracelet/ssh"
	"github.com/charmbracelet/wish"
	"golang.org/x/exp/slog"
	"net"
	"runtime/debug"
)

func TyperMiddleware(srv *Server) wish.Middleware {
	return func(h ssh.Handler) ssh.Handler {
		return func(s ssh.Session) {
			user := s.User()

			defer func() {
				if p, ok := srv.players[user]; ok {
					p.Room.deletePlayer(p)
					if len(p.Room.player) <= 1 {
						delete(srv.rooms, p.Room.id)
					}
				}
				delete(srv.players, user)

				if r := recover(); r != nil {
					fmt.Printf("panic: %v\n%s", r, string(debug.Stack()))
				}
			}()

			_, _, active := s.Pty()
			ip := s.RemoteAddr().(*net.TCPAddr).IP.String()

			slog.Info("player login in typer", "addr", ip, "user", user)
			if !active {
				_, _ = s.Write([]byte(help("No TTY")))
				_ = s.Exit(1)

				return
			}

			player := srv.findPlayer(s.User())

			if player != nil {
				_, _ = s.Write([]byte(fmt.Sprintf("Player %s with ip %s is already in the room ! \n", (*player.session).User(), ip)))
				_ = s.Exit(1)

				return
			}

			// 初始化玩家
			player = NewPlayer(&s, srv)

			// 初始化离线房间
			room := NewRoom(offline, srv, player)
			room.initDateForUser(player)
			room.player[user] = player

			srv.players[user] = player
			player.Room = room

			player.StartTyper(&s)

			slog.Info("player left", "addr", ip, "user", s.User())

			// 删除当前玩家的房间
			if len(player.Room.player) <= 1 {
				delete(srv.rooms, player.Room.id)
			}

			// 玩家房间删除玩家
			player.Room.deletePlayer(player)

			// 删除玩家
			delete(srv.players, user)
			//_ = s.Close()

			h(s)
		}
	}

}

func help(s string) string {
	help := `Typer: Practice typing in your terminal
Usage: ssh [<name>@]<host> -p <port>
%s
	`
	help = fmt.Sprintf(help, s)
	return help
}
