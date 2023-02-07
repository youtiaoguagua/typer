package server

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/ssh"
	"golang.org/x/exp/slog"
)

type Player struct {
	Id      string
	srv     *Server
	Room    *Room
	session *ssh.Session
	program *tea.Program
	typer   *TypeModel
}

func (p *Player) StartTyper(s *ssh.Session) {

	go func() {
		_, windowChanges, _ := (*s).Pty()
		for {
			select {
			case <-(*s).Context().Done():
				if p.program != nil {
					p.program.Quit()
					p.program = nil
					return
				}
			case w := <-windowChanges:
				if p != nil {
					p.program.Send(tea.WindowSizeMsg{Width: w.Width, Height: w.Height})
				}
			}
		}
	}()

	if _, err := p.program.Run(); err != nil {
		slog.Error("app exit with error:", err)
	}

	if p.program != nil {
		p.program.Kill()
	}
}

func (p *Player) changeOnlineStatus() {
	if p.Room.roomType == online {
		p.change2Offline()
		return
	}
	p.change2Online()
}

func (p *Player) change2Offline() {
	oldRoom := p.Room
	room := NewRoom(offline, p.srv, p)
	room.owner = p
	room.initDateForUser(p)

	//	如果房间只有一个人，删除旧房间
	if len(oldRoom.player) == 1 {
		p.srv.deleteRoom(oldRoom.id)
	}
	oldRoom.deletePlayer(p)

	p.Room = room
	room.player[p.Id] = p
}

func (p *Player) change2Online() {
	oldRoom := p.Room

	if len(oldRoom.player) == 1 {
		p.srv.deleteRoom(oldRoom.id)
	}
	oldRoom.deletePlayer(p)

	// 查找存在的房间,并且加入
	for _, v := range p.srv.rooms {
		if v.roomType == online && v.roomStatus == waitJoin {
			v.initDateForUser(p)
			p.Room = v
			v.player[p.Id] = p
			return
		}
	}

	// 创建在线房间
	room := NewRoom(online, p.srv, p)
	room.owner = p
	room.initDateForUser(p)

	p.Room = room
	room.player[p.Id] = p
}

func (p *Player) broadCastMsg(msg tea.Msg) {
	go func() {
		for _, v := range p.Room.player {
			if v.Id == p.Id {
				continue
			}

			v.program.Send(msg)
		}
	}()
}

func NewPlayer(s *ssh.Session, srv *Server) *Player {
	player := Player{
		Id:      (*s).User(),
		srv:     srv,
		session: s,
	}

	typeModel := NewTypeModel(&player)
	player.typer = typeModel

	p := tea.NewProgram(
		typeModel,
		tea.WithAltScreen(),
		tea.WithMouseCellMotion(),
		tea.WithInput(*s),
		tea.WithOutput(*s),
	)
	player.program = p

	return &player
}
