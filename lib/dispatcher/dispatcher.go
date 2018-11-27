package dispatcher

import (
	"errors"
	"path/filepath"
	"sync"
)

type Dispatcher interface {
	Register(path string)
	Cancel(path string)
	Dispatch(path string) (string, error)
}

var (
	rootPath   = "/"
	empty      = struct{}{}
	ErrNoRoute = errors.New("no route")
)

type routeNode struct {
	symbol         byte
	subTree        map[byte]*routeNode
	registeredPath map[string]struct{}
	mtx            sync.RWMutex
}

func NewDispatcher() Dispatcher {
	return &routeNode{
		subTree:        make(map[byte]*routeNode),
		registeredPath: make(map[string]struct{}),
	}
}

func (rn *routeNode) Register(path string) {
	trimedPath := filepath.Join(path)
	if trimedPath == "" {
		return
	}
	rn.mtx.Lock()
	defer rn.mtx.Unlock()

	if _, ok := rn.registeredPath[trimedPath]; ok {
		return
	}

	rn.registeredPath[trimedPath] = empty
	if trimedPath != "/" {
		trimedPath = trimedPath + "/"
	}
	curNode := rn
	for i := 0; i < len(trimedPath); i++ {
		curByte := []byte(trimedPath[i : i+1])[0]
		if _, ok := curNode.subTree[curByte]; !ok {

			curNode.subTree[curByte] = &routeNode{
				symbol:  curByte,
				subTree: make(map[byte]*routeNode),
			}
		}
		curNode = curNode.subTree[curByte]
	}
}

func (rn *routeNode) Cancel(path string) {
	panic("implement me")
}

func (rn *routeNode) Dispatch(path string) (string, error) {

	trimedPath := filepath.Join(path)
	if trimedPath == "" || trimedPath == rootPath {
		if _, ok := rn.registeredPath[rootPath]; ok {
			return rootPath, nil
		}
		return "", ErrNoRoute
	}
	rn.mtx.RLock()
	defer rn.mtx.RUnlock()

	if trimedPath != "/" {
		trimedPath = trimedPath + "/"
	}
	ret := ""
	tmp := ""
	curNode := rn
	for i := 0; i < len(trimedPath); i++ {
		curByte := []byte(trimedPath[i : i+1])[0]
		node, ok := curNode.subTree[curByte]
		if !ok {
			break
		}
		curNode = node
		tmp = tmp + string([]byte{curByte})
		if string([]byte{curByte}) == "/" {
			ret = ret + tmp
			tmp = ""
		}
	}
	ret = filepath.Join(ret)
	if _, ok := rn.registeredPath[ret]; ok {
		return ret, nil
	}
	return "", ErrNoRoute
}
