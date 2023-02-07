package server

import (
	zone "github.com/lrstanley/bubblezone"
	wrap "github.com/mitchellh/go-wordwrap"
	"github.com/youtiaoguagua/typer/typer"
	"golang.org/x/exp/slog"
	"strconv"
	"strings"
	"sync"
	"time"
)

type RoomType int
type RoomStatus int

const (
	offline RoomType = iota
	online
)

const (
	waitJoin RoomStatus = iota
	joined
	countDown
	processing
	typeEnd
)

type TypeWordInfo struct {
	Text          []rune
	Typed         []rune
	Start         time.Time
	End           time.Time
	Score         float64
	Wpms          []float64
	Wpm           float64
	CompleteWpm   CompleteWpm
	Correct       bool
	AnimationChan chan struct{}
	AnimationType int
	Hitokoto      Hitokoto
	Zm            *zone.Manager
	WordConf      *typer.WordConfig
	once          sync.Once
}

type CompleteWpm struct {
	Wpms []float64
	Wpm  float64
}

type Room struct {
	id           string
	roomType     RoomType
	roomStatus   RoomStatus
	player       map[string]*Player
	owner        *Player
	TypeWordInfo map[string]*TypeWordInfo
	CountDown    int
}

func NewRoom(roomType RoomType, srv *Server, owner *Player) *Room {
	room := Room{
		id:           strconv.FormatInt(time.Now().UnixMilli(), 10),
		roomType:     roomType,
		roomStatus:   waitJoin,
		player:       map[string]*Player{},
		TypeWordInfo: map[string]*TypeWordInfo{},
		CountDown:    5,
		owner:        owner,
	}

	srv.rooms[room.id] = &room
	return &room
}

func (r *Room) initDateForUser(player *Player) {
	if _, ok := r.player[player.Id]; ok {
		return
	}

	// 初始化玩家数据
	config := typer.WordConfig{
		Length:      25,
		Numbers:     false,
		Punctuation: false,
	}

	wordText := wrap.WrapString(typer.GetWordData(&config), uint(globalWith))

	wordInfo := TypeWordInfo{
		Text:          []rune(wordText),
		Typed:         []rune{},
		Wpms:          []float64{0},
		AnimationType: 0,
		AnimationChan: make(chan struct{}, 1),
		Zm:            zone.New(),
		WordConf:      &config,
	}
	wordInfo.InitHitokoto()

	r.TypeWordInfo[player.Id] = &wordInfo

	// 配置文件共享
	if r.roomType == online && len(r.player) == 1 {
		for k, v := range r.TypeWordInfo {
			if r.owner != nil && k != r.owner.Id {
				continue
			}

			t := make([]rune, len(v.Text))
			copy(t, v.Text)
			info := r.TypeWordInfo[player.Id]
			info.WordConf = v.WordConf
			info.Text = v.Text
			r.roomStatus = joined

			break
		}
	}

	wordInfo.once.Do(func() {
		go func() {
			for {
				select {
				case <-wordInfo.AnimationChan:
					return
				default:
					if player.program == nil {
						return
					}
					player.program.Send(AnimationMsg{})
					slog.Debug("AnimationType is running", "room", r.id)

					if player.Room.owner == nil {
						player.Room.owner = player
					}
				}
				time.Sleep(1200 * time.Millisecond)
			}
		}()
	})
}

func (r *Room) deletePlayer(p *Player) {
	if r == nil {
		return
	}
	if r.owner != nil && r.owner.Id == p.Id {
		r.owner = nil
	}
	r.roomStatus = waitJoin
	r.TypeWordInfo[p.Id].Close()
	delete(p.Room.player, p.Id)
	delete(r.TypeWordInfo, p.Id)
	r.ChangeWord()
}

func (r *Room) ChangeWord() {
	if len(r.TypeWordInfo) <= 0 {
		return
	}
	wordInfos := make([]*TypeWordInfo, 0, len(r.TypeWordInfo))
	for _, v := range r.TypeWordInfo {
		wordInfos = append(wordInfos, v)
	}
	wordInfos[0].ChangeWord()

	for _, v := range wordInfos[1:] {
		v.ChangeWord()
		v.Text = wordInfos[0].Text
	}

}

func (t *TypeWordInfo) ChangeWord() {
	wordText := wrap.WrapString(typer.GetWordData(t.WordConf), uint(globalWith))
	t.Text = []rune(wordText)
	t.Typed = []rune{}
	t.Wpms = []float64{0}
	t.Score = 0
	t.Start = time.Time{}

}

func (t *TypeWordInfo) Close() {
	if len(t.AnimationChan) == 0 {
		t.AnimationChan <- struct{}{}
	}
}

func (t *TypeWordInfo) renderWord() string {
	text := strings.Builder{}
	for i, c := range t.Text {
		if i < len(t.Typed) {
			x := t.Typed[i]
			if x == c {
				text.WriteString(typer.CorrectColor.Render(string(c)))
			} else {
				text.WriteString(typer.IncorrectColor.Render(string(c)))
			}
		} else {
			text.WriteString(typer.RootColor.Render(string(c)))
		}

	}
	return text.String()
}
