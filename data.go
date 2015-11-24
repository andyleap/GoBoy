package GoBoy

import (
	"bytes"
	"encoding/json"
	"sync"
)

type DataNode struct {
	cg *ConnectedGame
	nodeID uint32
	Val interface{}
}

type DataArray struct {
	cg *ConnectedGame
	nodes []uint32
}

func (da *DataArray) Get(index int) *DataNode {
	return da.cg.getNode(da.nodes[index])
}

func (da *DataArray) Len() int {
	return len(da.nodes)
}

type DataObject struct {
	sync.Mutex
	cg *ConnectedGame
	nodes map[string]uint32
}

func (do *DataObject) Get(key string) *DataNode {
	do.Lock()
	defer do.Unlock()
	return do.cg.getNode(do.nodes[key])
}

func (do *DataNode) MarshalJSON() ([]byte, error) {
	switch dp := do.Val.(type) {
	case *DataArray:
		parts := [][]byte{}
		for part := range dp.nodes {
			data, err := json.Marshal(dp.Get(part))
			if err != nil {
				return nil, err
			}
			parts = append(parts, data)
		}
		return append(append([]byte(`[`), bytes.Join(parts, []byte(`,`))...), byte(']')), nil
	case *DataObject:
		parts := [][]byte{}
		dp.Lock()
		defer dp.Unlock()
		for key, id := range dp.nodes {
			keydata, err := json.Marshal(key)
			if err != nil {
				return nil, err
			}
			data, err := json.Marshal(dp.cg.getNode(id))
			if err != nil {
				return nil, err
			}
			keydata = append(keydata, byte(':'))
			keydata = append(keydata, data...)
			parts = append(parts, keydata)
		}
		return append(append([]byte(`{`), bytes.Join(parts, []byte(`,`))...), byte('}')), nil
	default:
		data, err := json.Marshal(do.Val)
		if err != nil {
			return nil, err
		}
		return data, nil
	}
}