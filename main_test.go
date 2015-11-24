package GoBoy

import (
	"io/ioutil"
	"encoding/json"
	"fmt"
	"time"
	"testing"
)

func TestDiscover(t *testing.T) {
	games, err := DiscoverGames()
	fmt.Println(games, err)
	if len(games) == 0 {
		return
	}
	cg, err := games[0].Connect()
	fmt.Println(err)
	time.Sleep(10 * time.Second)
	marshalled, _ := json.MarshalIndent(cg.DataRoot(), "", "	")
	ioutil.WriteFile("data.json", marshalled, 0666)
}