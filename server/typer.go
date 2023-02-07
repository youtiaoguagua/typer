package server

import (
	"fmt"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/guptarohit/asciigraph"
	"github.com/youtiaoguagua/typer/typer"
	"github.com/youtiaoguagua/typer/util"
	"sort"
	"strconv"
	"strings"
	"time"
)

var (
	charsPerWord = 5.
	globalWith   = 60
)

type TypeModel struct {
	Player   *Player
	WinHigh  int
	WinWidth int
}

type AnimationMsg struct {
}

type TypeChangeMsg struct {
}

func NewTypeModel(p *Player) *TypeModel {
	pty, _, _ := (*p.session).Pty()
	typeModel := TypeModel{
		Player:   p,
		WinHigh:  pty.Window.Height,
		WinWidth: pty.Window.Width,
	}

	return &typeModel
}

func (t TypeModel) Init() tea.Cmd {
	return nil
}

func (t TypeModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	player := t.Player
	typeWordInfoMap := player.Room.TypeWordInfo
	m := typeWordInfoMap[player.Id]
	room := player.Room

	switch msg := msg.(type) {
	case tea.KeyMsg:
		if msg.Type == tea.KeyCtrlC {
			return t, tea.Quit
		}

		if room.roomType == online && room.roomStatus <= countDown {
			return t, nil
		}

		if m.Start.IsZero() {
			m.Start = time.Now()
		}

		if msg.Type == tea.KeyBackspace && len(m.Typed) > 0 {
			m.Typed = m.Typed[:len(m.Typed)-1]
			return t, nil
		}

		if !(msg.Type == tea.KeyRunes || msg.Type == tea.KeySpace) {
			return t, nil
		}

		if len(m.Text) == len(m.Typed) {
			return t, nil
		}

		char := msg.Runes[0]
		next := m.Text[len(m.Typed)]

		if next == '\n' {
			m.Typed = append(m.Typed, next)

			if char == ' ' {
				return t, nil
			}
		}

		m.Typed = append(m.Typed, msg.Runes...)

		if char == next {
			m.Score += 1
		}

		player.broadCastMsg(TypeChangeMsg{})
		if len(m.Text) == len(m.Typed) {
			wpms := make([]float64, len(m.Wpms))
			copy(wpms, m.Wpms)

			m.CompleteWpm = CompleteWpm{
				Wpms: wpms,
				Wpm:  m.Wpm,
			}
			m.End = time.Now()
			room.roomStatus = typeEnd

			return t, nil
		}
	case AnimationMsg:
		if (m.AnimationType + 1) >= len(typer.CatLegType) {
			m.AnimationType = 0
		} else {
			m.AnimationType++
		}
	case tea.WindowSizeMsg:
		t.WinWidth = msg.Width
		t.WinHigh = msg.Height
	case TypeChangeMsg:
		return t, nil
	case tea.MouseMsg:
		switch {
		case msg.Type != tea.MouseLeft:
			return t, nil
		case m.Zm.Get("numbers").InBounds(msg):
			m.WordConf.Numbers = !m.WordConf.Numbers
			room.ChangeWord()
		case m.Zm.Get("change").InBounds(msg):
			room.ChangeWord()
		case m.Zm.Get("online").InBounds(msg):
			player.changeOnlineStatus()
		case m.Zm.Get("onlineStart").InBounds(msg):
			player.Room.roomStatus = countDown
			go func() {
				for i := 0; i < room.CountDown; i++ {
					player.Room.CountDown--
					time.Sleep(1 * time.Second)
				}
				room.roomStatus = processing

				for _, v := range room.TypeWordInfo {
					v.Start = time.Now()
				}
			}()
		}

		for _, v := range []string{"10", "25", "50"} {
			if m.Zm.Get("length" + v).InBounds(msg) {
				length, _ := strconv.Atoi(v)
				m.WordConf.Length = length
				room.ChangeWord()
				break
			}
		}

		player.broadCastMsg(TypeChangeMsg{})

		return t, nil
	}
	return t, nil
}

func (t TypeModel) View() string {
	player := t.Player
	typeWordInfoMap := t.Player.Room.TypeWordInfo
	wordInfo := typeWordInfoMap[player.Id]
	res := strings.Builder{}

	wpm, graph := wordInfo.getWpmInfo()

	// 标题
	{
		bar := getTopBar()
		res.WriteString(bar)
		res.WriteString("\n")
	}

	// 顶部吉祥物
	{

		banner := getBanner(wordInfo)
		res.WriteString(banner)
		res.WriteString("\n")
	}

	// 控制拦
	{
		btn := wordInfo.getControlBtn(player)
		res.WriteString(btn)
		res.WriteString("\n\n")
	}

	// 单词输入
	{
		{
			if player.Room.roomType == online {
				res.WriteString("You")
				res.WriteString("\n")
			}
			text := player.Room.TypeWordInfo[player.Id].renderWord()
			res.WriteString(text)
			res.WriteString("\n\n")
		}

		{
			if player.Room.roomType == online && len(player.Room.player) > 1 {
				res.WriteString("Others")
				res.WriteString("\n")
				for k, v := range player.Room.TypeWordInfo {
					if player.Id == k {
						continue
					}
					text := v.renderWord()
					res.WriteString(text)
					res.WriteString("\n\n")
				}
			}

		}

	}

	// 一言
	//{
	//	res.WriteString("\n")
	//	hitokoto := generateHitokoto(wordInfo)
	//	res.WriteString(hitokoto)
	//	res.WriteString("\n")
	//}

	// 底部状态栏
	{
		res.WriteString("\n")
		bar := wordInfo.getBottomStatusBar(wpm, player)
		res.WriteString(bar)
		res.WriteString("\n\n")
	}

	result := lipgloss.PlaceHorizontal(t.WinWidth, lipgloss.Center, lipgloss.NewStyle().Width(globalWith).Render(res.String()))

	// 底部图表
	var graphNew string
	{
		if player.Room.roomType == offline {
			w := (t.WinWidth - globalWith) / 2
			graphNew = lipgloss.NewStyle().MarginLeft(w).Render(graph)
		}
	}

	result = lipgloss.JoinVertical(lipgloss.Top, result, graphNew)
	result = lipgloss.Place(t.WinWidth, t.WinHigh, lipgloss.Center, lipgloss.Top, result)

	return wordInfo.Zm.Scan(result)
}

func (t *TypeWordInfo) getBottomStatusBar(wpm float64, player *Player) string {
	statusBarStyle := lipgloss.NewStyle().
		Foreground(lipgloss.AdaptiveColor{Light: "#343433", Dark: "#C1C6B2"}).
		Background(lipgloss.AdaptiveColor{Light: "#D9DCCF", Dark: "#353533"})

	statusNugget := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#FFFDF5")).
		Padding(0, 1)

	status := getStatusMessage(t, player)

	statusKey := lipgloss.NewStyle().
		Inherit(statusBarStyle).
		Foreground(lipgloss.Color("#FFFDF5")).
		Background(lipgloss.Color("#FF5F87")).
		Padding(0, 1).
		MarginRight(1).Render("STATUS")

	line := "offline"
	if player.Room.roomType == online {
		line = "online"
	}
	encoding := statusNugget.Copy().
		Background(lipgloss.Color("#A550DF")).
		Align(lipgloss.Right).Render(line)

	fishCake := statusNugget.Copy().Background(lipgloss.Color("#6124DF")).Render(fmt.Sprintf("WPM:%.2f", wpm))

	statusVal := lipgloss.NewStyle().Inherit(statusBarStyle).Copy().
		Width(globalWith - lipgloss.Width(statusKey) - lipgloss.Width(encoding) - lipgloss.Width(fishCake)).
		Render(status)

	bar := lipgloss.JoinHorizontal(lipgloss.Top,
		statusKey,
		statusVal,
		encoding,
		fishCake,
	)

	return bar
}

func getStatusMessage(t *TypeWordInfo, player *Player) string {
	var status string
	if player.Room.roomType == offline {
		switch {
		case t.Start.IsZero():
			status = "Wait input"
		case len(t.Text) == len(t.Typed):
			status = "Complete 🎉🎉🎉🎉"
		default:
			status = "Typing"
		}
		return status
	}

	r := player.Room
	switch {
	case r.roomStatus == waitJoin:
		status = "wait someone join"
	case r.roomStatus == joined:
		status = "wait room owner click to start"
	case r.roomStatus == countDown:
		status = fmt.Sprintf("%d seconds to start", r.CountDown)
	case r.roomStatus == processing:
		switch {
		case t.Start.IsZero():
			status = "Wait input"
		default:
			status = "Typing"
		}
	case r.roomStatus == typeEnd:
		vals := make([]*TypeWordInfo, 0, len(r.TypeWordInfo))
		for _, v := range r.TypeWordInfo {
			if !v.End.IsZero() {
				vals = append(vals, v)
			}
		}

		sort.Slice(vals, func(i, j int) bool {
			return vals[i].End.Before(vals[j].End)
		})

		if vals[0] == r.TypeWordInfo[player.Id] {
			status = "Complete 🎉🎉🎉🎉"
		} else {
			status = "keep going 😅😅"
		}
	}

	return status
}

func getTopBar() string {
	bar := lipgloss.NewStyle().
		Background(lipgloss.Color("#806d9e")).
		Foreground(lipgloss.Color("#FFFDF5")).
		Align(lipgloss.Center).
		Width(globalWith).
		Bold(true).
		Render("TYPER")
	return bar
}

func generateHitokoto(wordInfo *TypeWordInfo) string {
	fromStr := strings.Builder{}
	fromStr.WriteString("——")
	if wordInfo.Hitokoto.FromWho != "" {
		fromStr.WriteString(wordInfo.Hitokoto.FromWho)
	}

	if wordInfo.Hitokoto.From != "" {
		fromStr.WriteString(fmt.Sprintf("「%s」", wordInfo.Hitokoto.From))
	}

	from := lipgloss.NewStyle().Width(globalWith).Align(lipgloss.Right).Foreground(lipgloss.Color("#AAE3E2")).Render(fromStr.String())

	// 一言
	hitokoto := lipgloss.JoinVertical(lipgloss.Top,
		lipgloss.NewStyle().Foreground(lipgloss.Color("#CDE990")).Render(wordInfo.Hitokoto.Hitokoto),
		from,
	)

	return hitokoto
}

func (t *TypeWordInfo) getWpmInfo() (wpm float64, graph string) {
	wpms := t.Wpms

	if len(t.Text) == len(t.Typed) {
		wpm = t.CompleteWpm.Wpm
		wpms = t.CompleteWpm.Wpms
	} else {
		// 计算wpm
		wpm = (t.Score / charsPerWord) / (time.Since(t.Start).Minutes())
		t.Wpm = wpm

		if len(t.Typed) > int(charsPerWord) {
			t.Wpms = append(t.Wpms, wpm)
		}
	}

	graph = asciigraph.Plot(
		wpms,
		asciigraph.Height(5),
		asciigraph.Width(globalWith-10),
		asciigraph.Precision(2),
		asciigraph.SeriesColors(asciigraph.YellowGreen),
	)

	return wpm, graph
}

func (t *TypeWordInfo) InitHitokoto() {
	hitokoto := Hitokoto{}
	if err := util.FetchData("https://v1.hitokoto.cn/", &hitokoto); err != nil {
		hitokoto = Hitokoto{Hitokoto: "迎着风，拥抱彩虹！", From: "你的答案", FromWho: "黄霄雲"}
	}

	t.Hitokoto = hitokoto
}

func (t *TypeWordInfo) getControlBtn(player *Player) string {
	conf := t.WordConf

	// 数字控制
	numBtn := conf.GetNumBtn()
	control := lipgloss.JoinHorizontal(lipgloss.Left, t.Zm.Mark("numbers", numBtn), typer.Gap, typer.Gap)

	// 单词控制
	for _, v := range []string{"10", "25", "50"} {
		lengthBtn := conf.GetLengthBtn(v)
		control += lipgloss.JoinHorizontal(lipgloss.Left, t.Zm.Mark("length"+v, lengthBtn), typer.Gap)
	}

	// 重新开始
	//change := lipgloss.NewStyle().Background(lipgloss.Color("#8EA7E9")).Foreground(lipgloss.Color("#FAD6A5")).Padding(0, 1).Render("restart")
	change := lipgloss.NewStyle().Background(lipgloss.Color("#696969")).Foreground(lipgloss.Color("#FAD6A5")).Padding(0, 1).Render("restart")
	control += lipgloss.JoinHorizontal(lipgloss.Left, t.Zm.Mark("change", change), typer.Gap)

	// 在线or离线
	var onlineStr string
	if player.Room.roomType == offline {
		onlineStr = typer.Btn.Render("offline")
	} else {
		onlineStr = typer.BtnPress.Render("online")
	}
	control += lipgloss.JoinHorizontal(lipgloss.Left, t.Zm.Mark("online", onlineStr))

	// 已经加入，展示开始按钮
	room := player.Room
	if room.roomType == online {
		if room.owner != nil && room.owner.Id == player.Id {
			var waitStr string
			if room.roomStatus == waitJoin {
				waitStr = typer.Btn.Render("wait")
				control += lipgloss.JoinHorizontal(lipgloss.Left, " -> wait join")

			} else if room.roomStatus == joined {
				waitStr = typer.BtnPress.Render("start")
				control += lipgloss.JoinHorizontal(lipgloss.Left, typer.Gap, t.Zm.Mark("onlineStart", waitStr))
			}
		}

		if room.roomStatus == countDown {
			countDownStr := typer.Btn.Render(strconv.Itoa(room.CountDown))
			control += lipgloss.JoinHorizontal(lipgloss.Left, typer.Gap, countDownStr)
		}
	}

	return control
}

func getBanner(t *TypeWordInfo) string {
	// 吉祥物设置
	cat := lipgloss.NewStyle().Foreground(lipgloss.Color("#fa8231")).Render(typer.Cat2)

	color := typer.CatLegTypeColor[t.AnimationType]
	legWithCol := lipgloss.NewStyle().Foreground(lipgloss.Color(color)).Render("づ")

	legThing := typer.CatLegType[t.AnimationType]

	cat = fmt.Sprintf(cat, legWithCol, legThing)

	// 信息框
	w := globalWith - lipgloss.Width(cat)

	fromStr := strings.Builder{}
	fromStr.WriteString("——")
	fromStr.WriteString(t.Hitokoto.FromWho)

	if t.Hitokoto.From != "" {
		fromStr.WriteString(fmt.Sprintf("「%s」", t.Hitokoto.From))
	}

	from := lipgloss.NewStyle().Width(w).Align(lipgloss.Right).Foreground(lipgloss.Color("#AAE3E2")).Render(fromStr.String())

	content := lipgloss.NewStyle().Width(w - 4).Foreground(lipgloss.Color("#CDE990")).Render(t.Hitokoto.Hitokoto)

	// 一言
	hitokoto := lipgloss.JoinVertical(lipgloss.Top,
		"\n"+content,
		from,
	)

	res := lipgloss.JoinHorizontal(lipgloss.Top, cat, hitokoto)

	return res
}
