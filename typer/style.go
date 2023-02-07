package typer

import "github.com/charmbracelet/lipgloss"

var (
	// 没有敲过的单字的颜色
	RootColor = lipgloss.NewStyle().Foreground(lipgloss.Color("#48494b"))
	// 正确
	CorrectColor = lipgloss.NewStyle().Foreground(lipgloss.Color("#79a617"))
	// 错误
	IncorrectColor = lipgloss.NewStyle().Foreground(lipgloss.Color("#a61717")).Background(lipgloss.Color("#ff7979"))

	// 松开按钮
	Btn = lipgloss.NewStyle().Background(lipgloss.Color("#696969")).Align(lipgloss.Center).Padding(0, 1)

	// 按下按钮
	BtnPress = Btn.Copy().Background(lipgloss.Color("#79a617"))
)

var CatLegType = []string{"🍪", "⛄"}
var CatLegTypeColor = []string{"#F2921D", "#F5F5F5"}

const (
	Gap = " "

	Cat1 = `         /\_/\  
    ____/ o o \ 
  /~____  =-= / 
 (______)__m_m) `

	Cat2 = ` 
 /\___/\
꒰˶• ༝ - ˶꒱
./%s%s
`

	Cat3 = ` 
 /\___/\
꒰˶• ༝ - ˶꒱
./づ~⛄
`

	Cat4 = `
 ∧,,,∧
(  ̳• · • ̳)
/    づ♡ 
`

	Logo = `  _____                          
 |_   _|_   _  _ __    ___  _ __ 
   | | | | | || '_ \  / _ \| '__|
   | | | |_| || |_) ||  __/| |   
   |_|  \__, || .__/  \___||_|   
        |___/ |_|                `
)
