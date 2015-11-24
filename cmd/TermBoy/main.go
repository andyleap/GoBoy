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
	cg, err := games[0].Connect()
	
	if err != nil {
		fmt.Println(err)
		return
	}
	
	name := ui.NewPar("")
	name.Height = 3
	name.BorderLabel = "Name"
	
	caps := ui.NewPar("")
	caps.Height = 3
	caps.BorderLabel = "Caps"
	
	clock := ui.NewPar("")
	clock.Height = 3
	clock.BorderLabel = "Time"
	
	xp := ui.NewGauge()
	xp.Height = 3
	xp.BorderLabel = "XP"
	xp.PercentColorHighlighted = ui.ColorBlack
	
	hp := ui.NewGauge()
	hp.Height = 3
	hp.BorderLabel = "HP"
	hp.PercentColorHighlighted = ui.ColorBlack
	
	inv := ui.NewGauge()
	inv.Height = 3
	inv.BorderLabel = "Weight"
	inv.PercentColorHighlighted = ui.ColorBlack
	
	quests := ui.NewList()
	quests.Height = 12
	quests.BorderLabel = "Quests"
	
	ui.Body.AddRows(
		ui.NewRow(
			ui.NewCol(3, 0, name),
			ui.NewCol(1, 0, caps),
			ui.NewCol(2, 0, clock),
			ui.NewCol(6, 0, xp),
		),
		ui.NewRow(
			ui.NewCol(6, 0, hp),
			ui.NewCol(6, 0, inv),
		),
		ui.NewRow(
			ui.NewCol(5, 0, quests),
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
		if clockVal, ok := cg.Path("PlayerInfo.TimeHour"); ok {
			hours := int(clockVal.(float32))
			minutes := int((clockVal.(float32) - float32(hours))*60)
			clock.Text = fmt.Sprintf("%2.0d:%2.0d", hours, minutes)
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
		{
			maxWVal, ok1 := cg.Path("PlayerInfo.MaxWeight")
			curWVal, ok2 := cg.Path("PlayerInfo.CurrWeight")
			if ok1 && ok2 {
				maxW := maxWVal.(float32)
				curW := curWVal.(float32)
				
				inv.Percent = int(((curW / maxW) * 100))
				inv.Label = fmt.Sprintf("%.0f / %.0f", curW, maxW)
				if curW > maxW {
					inv.Percent = 100
					inv.BarColor = ui.ColorRed
				} else {
					inv.BarColor = ui.ColorGreen
				}
			}
		}
		if questListVal, ok := cg.Path("Quests"); ok {
			quests.Items = []string{}
			questList := questListVal.(*GoBoy.DataArray)
			for l1 := 0; l1 < questList.Len(); l1++ {
				quest := questList.Get(l1)
				current, ok := quest.Path("enabled")
				if ok && current.(bool) {
					active, _ := quest.Path("active")
					name, _ := quest.Path("text")
					text := name.(string)
					if active.(bool) {
						text = "[" + text + "](fg-black,bg-green)"
					}
					quests.Items = append(quests.Items, text)
				}
			}
		}
		ui.Body.Width = ui.TermWidth()
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