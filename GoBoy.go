// GoBoy project GoBoy.go
package GoBoy

import (
	"log"
	"strconv"
	"strings"
	"math"
	"fmt"
	"encoding/binary"
	"io"
	"encoding/json"
	"sync"
	"time"
	"net"
)

type Game struct{
	IP net.IP
	MachineType string
	IsBusy bool
}

const (
	KEEP_ALIVE = iota
	CONNECTION_ACCEPTED
	CONNECTION_REFUSED
	DATA_UPDATE
	LOCAL_MAP_UPDATE
	COMMAND
	COMMAND_RESULT
	COUNT
)

const (
	VAL_BOOL = iota
	VAL_INT_8
	VAL_UINT_8
	VAL_INT_32
	VAL_UINT_32
	VAL_FLOAT
	VAL_STRING
	VAL_ARRAY
	VAL_OBJECT
)

func DiscoverGames() ([]*Game, error) {
	listen, err := net.ListenUDP("udp4", &net.UDPAddr{})
	log.Println(listen.LocalAddr().String())
	if err != nil {
		return nil, err
	}
	addrs, _ := net.InterfaceAddrs()
	done := sync.WaitGroup{}
	done.Add(1)
	games := []*Game{}
	listen.SetDeadline(time.Now().Add(5 * time.Second))
	go func () {
		for {
	        data := make([]byte, 4096)
	        read, remoteAddr, err := listen.ReadFromUDP(data)
			if err != nil && err.(net.Error).Timeout() {
				done.Done()
				return
			}
			game := &Game{
				IP: remoteAddr.IP,
			}
			json.Unmarshal(data[:read], &game)
			games = append(games, game)
		}
	}()
	for _, addr := range addrs {
		listen.WriteToUDP([]byte(`{"cmd": "autodiscover"}`), &net.UDPAddr{
			IP: broadcast(addr.(*net.IPNet)),
			Port: 28000,
		})
	}
	listen.WriteToUDP([]byte(`{"cmd": "autodiscover"}`),
		&net.UDPAddr{
			IP: net.IPv4bcast,
			Port: 28000,
		})
	done.Wait()
	return games, nil
}

type ConnectedGame struct {
	Game *Game
	conn *net.TCPConn
	keepAliveRunning bool
	nodes map[uint32]*DataNode
	nodeLock sync.Mutex
}

func (g *Game) Connect() (*ConnectedGame, error) {
	addr := &net.TCPAddr{
		IP: g.IP,
		Port: 27000,
	}
	conn, err := net.DialTCP("tcp4", nil, addr)
	if err != nil {
		return nil, err
	}
	data := make([]byte, 5)
	io.ReadFull(conn, data)
	msgType := data[4]
	packetLen := binary.LittleEndian.Uint32(data)
	packet := make([]byte, packetLen)
	io.ReadFull(conn, packet)
	
	switch(msgType) {
		case CONNECTION_ACCEPTED:
		cg := &ConnectedGame{
			Game: g,
			conn: conn,
			nodes: make(map[uint32]*DataNode),
		}
		go cg.handlePackets()
		return cg, nil
		case CONNECTION_REFUSED:
		fallthrough
		default:
		conn.Close()
		return nil, fmt.Errorf("Fallout 4 on %v refused connection", g.IP)
	}
}

func (cg *ConnectedGame) handlePackets() {
	for {
		data := make([]byte, 5)
		io.ReadFull(cg.conn, data)
		msgType := data[4]
		packetLen := binary.LittleEndian.Uint32(data)
		packet := make([]byte, packetLen)
		io.ReadFull(cg.conn, packet)
		switch(msgType) {
		case KEEP_ALIVE:
			if !cg.keepAliveRunning {
				go func() {
					for {
						time.Sleep(1 * time.Second)
						cg.conn.Write([]byte{0, 0, 0, 0, 0})
					}
				}()
			}
		case DATA_UPDATE:
			offset := uint32(0)
			for offset < packetLen {
				valueType := packet[offset]
				offset++
				nodeID := binary.LittleEndian.Uint32(packet[offset:offset+4])
				offset += 4
				var val interface{}
				switch(valueType){
				case VAL_BOOL:
					val = (packet[offset] != 0)
					offset++
				case VAL_INT_8:
					val = int8(packet[offset])
					offset++
				case VAL_UINT_8:
					val = uint8(packet[offset])
					offset++
				case VAL_INT_32:
					val = int32(binary.LittleEndian.Uint32(packet[offset:offset+4]))
					offset+=4
				case VAL_UINT_32:
					val = binary.LittleEndian.Uint32(packet[offset:offset+4])
					offset+=4
				case VAL_FLOAT:
					val = math.Float32frombits(binary.LittleEndian.Uint32(packet[offset:offset+4]))
					offset+=4
				case VAL_STRING:
					var count uint32
					for count = uint32(0); packet[offset+count] != 0; count++ {}
					val = string(packet[offset:offset+count])
					offset += count + 1
				case VAL_ARRAY:
					count := binary.LittleEndian.Uint16(packet[offset:offset+2])
					offset += 2
					nodes := make([]uint32, count)
					for n := range nodes {
						nodes[n] = binary.LittleEndian.Uint32(packet[offset:offset+4])
						offset += 4
					}
					val = &DataArray{
						cg: cg,
						nodes: nodes,
					}
				case VAL_OBJECT:
					node := cg.getNode(nodeID)
					var do *DataObject
					if node != nil {
						do, _ = node.Val.(*DataObject)
					}
					if do == nil {
						do = &DataObject{
							cg: cg,
							nodes: make(map[string]uint32),
						}
					}
					do.Lock()
					count := int(binary.LittleEndian.Uint16(packet[offset:offset+2]))
					offset += 2
					for l1 := 0; l1 < count; l1++ {
						valID := binary.LittleEndian.Uint32(packet[offset:offset+4])
						offset+=4
						var strcount uint32
						for strcount = 0; packet[offset+strcount] != 0; strcount++ {}
						key := string(packet[offset:offset+strcount])
						offset += strcount + 1
						do.nodes[key] = valID
					}
					count = int(binary.LittleEndian.Uint16(packet[offset:offset+2]))
					offset += 2
					for l1 := 0; l1 < count; l1++ {
						valID := binary.LittleEndian.Uint32(packet[offset:offset+4])
						offset+=4
						var delKey string
						for key, val := range do.nodes {
							if val == valID {
								delKey = key
								break
							}
						}
						delete(do.nodes, delKey)
					}
					do.Unlock()
					val = do
				}
				cg.setNode(nodeID, val)
			}
		}
	}
}

func (cg *ConnectedGame) getNode(nodeID uint32) *DataNode {
	cg.nodeLock.Lock()
	defer cg.nodeLock.Unlock()
	return cg.nodes[nodeID]
}

func (cg *ConnectedGame) setNode(nodeID uint32, val interface{}) {
	cg.nodeLock.Lock()
	defer cg.nodeLock.Unlock()
	curNode, ok := cg.nodes[nodeID]
	if ok {
		curNode.Val = val
		return
	}
	curNode = &DataNode{
		cg: cg,
		nodeID: nodeID,
		Val: val,
	}
	cg.nodes[nodeID] = curNode
}

func (cg *ConnectedGame) DataRoot() *DataNode {
	return cg.getNode(0)
}

func (cg *ConnectedGame) Path(path string) (interface{}, bool) {
	parts := strings.Split(path, ".")
	curNode := cg.DataRoot()
	if curNode == nil {
		return nil, false
	}
	for _, part := range parts {
		switch cn := curNode.Val.(type) {
		case *DataArray:
			index, err := strconv.ParseInt(part, 0, 32)
			if err != nil {
				return nil, false
			}
			curNode = cn.Get(int(index))
		case *DataObject:
			curNode = cn.Get(part)
		default:
			return nil, false
		}
		if curNode == nil {
			return nil, false
		}
	}
	return curNode.Val, true
}

func (do *DataNode) Path(path string) (interface{}, bool) {
	parts := strings.Split(path, ".")
	curNode := do
	for _, part := range parts {
		switch cn := curNode.Val.(type) {
		case *DataArray:
			index, err := strconv.ParseInt(part, 0, 32)
			if err != nil {
				return nil, false
			}
			curNode = cn.Get(int(index))
		case *DataObject:
			curNode = cn.Get(part)
		default:
			return nil, false
		}
		if curNode == nil {
			return nil, false
		}
	}
	return curNode.Val, true
}

func (do *DataObject) Path(path string) (interface{}, bool) {
	parts := strings.SplitN(path, ".", 1)
	return do.Get(parts[0]).Path(parts[1])
}

func broadcast(n *net.IPNet) net.IP {
	ip := n.IP.To4()
	if ip == nil {
		ip = n.IP
		return net.IP{
			ip[0] | ^n.Mask[0], ip[1] | ^n.Mask[1], ip[2] | ^n.Mask[2],
			ip[3] | ^n.Mask[3], ip[4] | ^n.Mask[4], ip[5] | ^n.Mask[5],
			ip[6] | ^n.Mask[6], ip[7] | ^n.Mask[7], ip[8] | ^n.Mask[8],
			ip[9] | ^n.Mask[9], ip[10] | ^n.Mask[10], ip[11] | ^n.Mask[11],
			ip[12] | ^n.Mask[12], ip[13] | ^n.Mask[13], ip[14] | ^n.Mask[14],
			ip[15] | ^n.Mask[15]}
	}
	ip = net.IPv4(
		ip[0]|^n.Mask[0],
		ip[1]|^n.Mask[1],
		ip[2]|^n.Mask[2],
		ip[3]|^n.Mask[3])
	return ip
}