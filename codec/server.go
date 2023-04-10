package geerpc

import (
	"GeeRPC/codec/codec"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net"
	"reflect"
	"sync"
)

const MAGIC_NNMBER = 0x3bef5c

type Option struct {
	MagicNumber int
	CodecType   codec.Type
}

var DefaultOption = &Option{
	MagicNumber: MAGIC_NNMBER,
	CodecType:   codec.GobType,
}

type Server struct{}

func NewServer() *Server {
	return &Server{}
}

var DefaultServer = NewServer()

func (s *Server) Accept(lis net.Listener) {
	for {
		conn, err := lis.Accept()
		if err != nil {
			log.Printf("rpc server: accept error: %v", err)
			return
		}
		go s.ServeConn(conn)
	}
}

func Accept(lis net.Listener) {
	DefaultServer.Accept(lis)
}

func (s *Server) ServeConn(conn net.Conn) {
	defer func() { _ = conn.Close() }()

	var opt Option

	if err := json.NewDecoder(conn).Decode(&opt); err != nil {
		log.Printf("rpc server: options error: %v", err)
		return
	}

	if opt.MagicNumber != MAGIC_NNMBER {
		log.Printf("rpc server: invalid magic number %x", opt.MagicNumber)
		return
	}

	f := codec.NewCodecFuncMap[opt.CodecType]
	if f == nil {
		log.Printf("rpc server: invalid codec type: %v", opt.CodecType)
		return
	}

	s.serveCodec(f(conn))
}

var invalidRequest = struct{}{}

func (s *Server) serveCodec(cc codec.Codec) {
	sending := new(sync.Mutex)
	wg := new(sync.WaitGroup)

	for {
		req, err := s.readRequest(cc)
		if err != nil {
			if req == nil {
				break
			}
			req.h.Error = err.Error()
			s.sendResponse(cc, req.h, invalidRequest, sending)
		}
		wg.Add(1)
		go s.handleRequest(cc, req, sending, wg)
	}

	wg.Wait()
	_ = cc.Close()
}

type request struct {
	h            *codec.Header
	argv, replyv reflect.Value
}

func (s *Server) readRequestHeader(cc codec.Codec) (*codec.Header, error) {
	var h codec.Header

	if err := cc.ReadHeader(&h); err != nil {
		if err != io.EOF && err != io.ErrUnexpectedEOF {
			log.Printf("rpc server: read header error: %v", err)
		}
		return nil, err
	}
	return &h, nil
}

func (s *Server) readRequest(cc codec.Codec) (*request, error) {
	h, err := s.readRequestHeader(cc)
	if err != nil {
		return nil, err
	}

	req := &request{h: h}

	req.argv = reflect.New(reflect.TypeOf(""))

	if err = cc.ReadBody(req.argv.Interface()); err != nil {
		log.Printf("rpc server: read argv err: %v", err)
		return nil, err
	}

	return req, nil
}

func (s *Server) sendResponse(cc codec.Codec, h *codec.Header, body any, sending *sync.Mutex) {
	sending.Lock()
	defer sending.Unlock()

	if err := cc.Write(h, body); err != nil {
		log.Printf("rpc server: wirte response error: %v", err)
	}
}

func (s *Server) handleRequest(cc codec.Codec, req *request, sending *sync.Mutex, wg *sync.WaitGroup) {
	defer wg.Done()

	log.Printf("%v %v", req.h, req.argv.Elem())
	req.replyv = reflect.ValueOf(fmt.Sprintf("geerpc resp %d", req.h.Seq))
	s.sendResponse(cc, req.h, req.replyv.Interface(), sending)
}
