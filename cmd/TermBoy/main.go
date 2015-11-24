package main

import (
	"fmt"
	"time"
	"github.com/andyleap/GoBoy"
	ui "github.com/gizak/termui"
)

func main() {

	ui.ColorMap["fg"] = ui.ColorGreen
	ui.ColorMap["bg"] = ui.ColorBlack
	ui.ColorMap["border.fg"] = ui.ColorGreen
	ui.ColorMap["label.fg"] = ui.ColorGreen
	ui.ColorMap["par.fg"] = ui.ColorGreen
	ui.ColorMap["par.label.bg"] = ui.ColorGreen
	ui.ColorMap["par.label.fg"] = ui.ColorBlack
	ui.ColorMap["gauge.bar.bg"] = ui.ColorGreen
	err := ui.Init()
	if err != nil {
	    panic(err)
	}
	defer ui.Close()
	
	DisplayDiscovering()
	
	games, _ := GoBoy.DiscoverGames()
	if len(games) == 0 {
		return
	}
	cg, _ := games[0].Connect()
	
	name := ui.NewPar("")
	name.Height = 3
	name.BorderLabel = "Name"
	
	caps := ui.NewPar("")
	caps.Height = 3
	caps.BorderLabel = "Caps"
	
	xp := ui.NewGauge()
	xp.Height = 3
	xp.BorderLabel = "XP"
	xp.PercentColorHighlighted = ui.ColorBlack
	
	hp := ui.NewGauge()
	hp.Height = 3
	hp.BorderLabel = "HP"
	hp.PercentColorHighlighted = ui.ColorBlack
	
	rads := ui.NewGauge()
	rads.Height = 3
	rads.BorderLabel = "HP"
	rads.PercentColorHighlighted = ui.ColorBlack
	
	ui.Body.AddRows(
		ui.NewRow(
			ui.NewCol(3, 0, name),
			ui.NewCol(3, 0, caps),
			ui.NewCol(6, 0, xp),
		),
		ui.NewRow(
			ui.NewCol(6, 0, hp),
		),
	)
	
	
	for {
		if nameVal, ok := cg.Path("PlayerInfo.PlayerName"); ok {
			name.Text = nameVal.(string)
		}
		if capsVal, ok := cg.Path("PlayerInfo.Caps"); ok {
			caps.Text = fmt.Sprintf("%d", capsVal.(int32))
		}
		if xpVal, ok := cg.Path("PlayerInfo.XPProgressPct"); ok {
			xp.Percent = int(xpVal.(float32) * 100)
		}
		{
			maxHpVal, ok1 := cg.Path("PlayerInfo.MaxHP")
			curHpVal, ok2 := cg.Path("PlayerInfo.CurrHP")
			if ok1 && ok2 {
				maxHp := maxHpVal.(float32)
				curHp := curHpVal.(float32)
				
				hp.Percent = int(((curHp / maxHp) * 100))
				hp.Label = fmt.Sprintf("%.0f / %.0f", curHp, maxHp)
			}
		}
		ui.Body.Align()
		ui.Render(ui.Body)
		time.Sleep(1 * time.Second)
	}
	
	
	//ui.Loop()
}

func DisplayDiscovering() {
	p := ui.NewPar("Discovering Pip-Boys")
	p.Width = 22
	p.Height = 3
	discoverUi := ui.NewGrid(
		ui.NewRow(
			ui.NewCol(4, 4, p),
		),
	)
	discoverUi.Width = ui.Body.Width
	discoverUi.Align()
	ui.Render(discoverUi)
}