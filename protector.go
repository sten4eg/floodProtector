package floodProtecor

import (
	"errors"
	"net"
	"time"
)

// const fastConnectionLimit = 15
const normalConnectionTime = 700
const fastConnectionTime = 350
const maxConnectionPerIP = 50
const banTime = time.Minute
const safeConnInterval = 5000

type State int64

const (
	StateNormal State = iota
	StateWarn
	StateBlocked
)

type ConnectionInfo struct {
	connCount    int64
	lastConnTime int64
	lastConn     int64
	state        State
	blockExpire  time.Time
	_            [8]byte
}
type Storage interface {
	Store(string, ConnectionInfo)
	Load(string) (ConnectionInfo, bool)
}

func (ci *ConnectionInfo) UpdateState(currentTime int64, connectionTime int64) {
	switch ci.state {
	case StateNormal:
		if ci.isSuspicious(connectionTime) {
			ci.state = StateWarn
		}
	case StateWarn:
		if ci.isFlooding(connectionTime) {
			ci.state = StateBlocked
			ci.blockExpire = time.Now().Add(banTime)
		} else if ci.isBackToNormal(currentTime) {
			ci.state = StateNormal
			ci.connCount = 0
		}
	case StateBlocked:
		if time.Now().After(ci.blockExpire) {
			ci.state = StateNormal
			ci.connCount = 0
		}
	}
}

func (ci *ConnectionInfo) isSuspicious(connectionTime int64) bool {
	return ci.connCount > 2 && connectionTime < fastConnectionTime
}

func (ci *ConnectionInfo) isFlooding(connectionTime int64) bool {
	return ci.connCount > maxConnectionPerIP || connectionTime < normalConnectionTime
}

func (ci *ConnectionInfo) isBackToNormal(currentTime int64) bool {
	return currentTime-ci.lastConnTime > safeConnInterval
}

func AcceptTCP(conn *net.TCPConn, storage Storage) (*net.TCPConn, error) {
	ip, _, err := net.SplitHostPort(conn.RemoteAddr().String())
	if err != nil {
		_ = conn.Close()
		return nil, err
	}

	curTime := time.Now().UnixMilli()
	ci, ok := storage.Load(ip)
	if !ok {
		ci = ConnectionInfo{
			state:        StateNormal,
			connCount:    1,
			lastConnTime: curTime,
		}
		storage.Store(ip, ci)
	} else {
		connectionTime := curTime - ci.lastConnTime
		ci.connCount++
		ci.lastConnTime = curTime
		ci.UpdateState(curTime, connectionTime)
		storage.Store(ip, ci)
	}

	if ci.state == StateBlocked {
		_ = conn.Close()
		return nil, errors.New("соединение закрыто FloodProtection")
	}

	return conn, nil
}
