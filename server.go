package rcon

import (
	"bytes"
	"errors"
	"fmt"
	"net"
	"sync"
	"time"

	"github.com/sirupsen/logrus"
)

// DialFn connect to server using the options
type DialFn func(network, address string) (net.Conn, error)

// Server represents a Source engine game server.
type Server struct {
	addr string

	dial DialFn

	rconPassword string

	timeout time.Duration

	rsock           *rconSocket
	rconInitialized bool

	mu sync.Mutex
}

// ConnectOptions describes the various connections options.
type ConnectOptions struct {
	// Default will use net.Dialer.Dial. You can override the same by
	// providing your own.
	Dial DialFn

	// RCON password.
	RCONPassword string

	Timeout time.Duration
}

// Connect to the source server.
func Connect(addr string, os ...*ConnectOptions) (_ *Server, err error) {
	s := &Server{
		addr: addr,
	}
	if len(os) > 0 {
		o := os[0]
		s.dial = o.Dial
		s.rconPassword = o.RCONPassword
		s.timeout = o.Timeout
	}
	if s.dial == nil {
		s.dial = (&net.Dialer{
			Timeout: 1 * time.Second,
		}).Dial
	}
	if s.rconPassword == "" {
		return s, nil
	}
	if err := s.initRCON(); err != nil {
		return nil, err
	}
	return s, nil
}

func (s *Server) String() string {
	return s.addr
}

func (s *Server) initRCON() (err error) {
	if s.addr == "" {
		return errors.New("rcon: server needs a address")
	}
	log.WithFields(logrus.Fields{
		"addr": s.addr,
	}).Debug("rcon: connecting rcon")
	if s.rsock, err = newRCONSocket(s.dial, s.addr, s.timeout); err != nil {
		return fmt.Errorf("rcon: could not open tcp socket. %+v", err)
	}
	defer func() {
		if err != nil {
			s.rsock.close()
		}
	}()
	if err := s.authenticate(); err != nil {
		return fmt.Errorf("rcon: could not authenticate. %+v", err)
	}
	s.rconInitialized = true
	return nil
}

func (s *Server) authenticate() error {
	log.WithFields(logrus.Fields{
		"addr": s.addr,
	}).Debug("rcon: authenticating")
	req := newRCONRequest(rrtAuth, s.rconPassword)
	data, _ := req.marshalBinary()
	if err := s.rsock.send(data); err != nil {
		return err
	}
	// Receive the empty response value
	data, err := s.rsock.receive()
	if err != nil {
		return err
	}
	log.WithFields(logrus.Fields{
		"data": data,
	}).Debug("rcon: received empty response")
	var resp rconResponse
	if err = resp.unmarshalBinary(data); err != nil {
		return err
	}
	if resp.typ != rrtRespValue || resp.id != req.id {
		return ErrInvalidResponseID
	}
	if resp.id != req.id {
		return ErrInvalidResponseType
	}
	// Receive the actual auth response
	data, err = s.rsock.receive()
	if err != nil {
		return err
	}
	if err := resp.unmarshalBinary(data); err != nil {
		return err
	}
	if resp.typ != rrtAuthResp || resp.id != req.id {
		return ErrRCONAuthFailed
	}
	log.Debug("rcon: authenticated")
	return nil
}

// Close releases the resources associated with this server.
func (s *Server) Close() {
	if s.rconInitialized {
		s.rsock.close()
	}
}

// Send RCON command to the server.
func (s *Server) Send(cmd string) (string, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if !s.rconInitialized {
		return "", ErrRCONNotInitialized
	}
	req := newRCONRequest(rrtExecCmd, cmd)
	data, _ := req.marshalBinary()
	if err := s.rsock.send(data); err != nil {
		return "", fmt.Errorf("rcon: sending rcon request. %+v", err)
	}
	// Send the mirror packet.
	reqMirror := newRCONRequest(rrtRespValue, "")
	data, _ = reqMirror.marshalBinary()
	if err := s.rsock.send(data); err != nil {
		return "", fmt.Errorf("rcon: sending rcon mirror request. %+v", err)
	}
	var (
		buf       bytes.Buffer
		sawMirror bool
	)
	// Start receiving data.
	for {
		data, err := s.rsock.receive()
		if err != nil {
			return "", fmt.Errorf("rcon: receiving rcon response. %+v", err)
		}
		var resp rconResponse
		if err = resp.unmarshalBinary(data); err != nil {
			return "", fmt.Errorf("rcon: decoding response. %+v", err)
		}
		if resp.typ != rrtRespValue {
			return "", ErrInvalidResponseType
		}
		if !sawMirror && resp.id == reqMirror.id {
			sawMirror = true
			continue
		}
		if sawMirror {
			if bytes.Equal(resp.body, trailer) {
				break
			}
			return "", ErrInvalidResponseTrailer
		}
		if req.id != resp.id {
			return "", ErrInvalidResponseID
		}
		_, err = buf.Write(resp.body)
		if err != nil {
			return "", err
		}
	}
	return buf.String(), nil
}

var (
	trailer = []byte{0x00, 0x01, 0x00, 0x00}

	// ErrRCONAuthFailed RCON Authentication failed error
	ErrRCONAuthFailed = errors.New("rcon: authentication failed")
	// ErrRCONNotInitialized RCON connection is not initialized error
	ErrRCONNotInitialized = errors.New("rcon: rcon is not initialized")
	// ErrInvalidResponseType RCON Invalid response type from server error
	ErrInvalidResponseType = errors.New("rcon: invalid response type from server")
	// ErrInvalidResponseID RCON invalid response id from server error
	ErrInvalidResponseID = errors.New("rcon: invalid response id from server")
	// ErrInvalidResponseTrailer RCON invalid response trailer from server error
	ErrInvalidResponseTrailer = errors.New("rcon: invalid response trailer from server")
)
