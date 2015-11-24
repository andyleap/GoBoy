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
	cg, err := games[0].Connect()
	fmt.Println(err)
	time.Sleep(10 * time.Second)
	marshalled, _ := json.MarshalIndent(cg.DataRoot(), "", "	")
	fmt.Println(string(marshalled))
	ioutil.WriteFile("data.json", marshalled, 0666)
}