// mcastrpc
package mcastrpc

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net"
	"reflect"
)

const (
	MAX_READ_BYTES         = 8042
	JSONRPC_V2             = "2.0"
	ERROR_INVALID_REQUEST  = -32600
	ERROR_INVALID_JSON     = -32700
	ERROR_METHOD_NOT_EXIST = -32601
)

type invalidRequest struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

type mcastRequest struct {
	Jsonrpc string          `json:"jsonrpc"`
	Id      int             `json:"id"`
	Method  string          `json:"method"`
	Params  json.RawMessage `json:"params"`
}

type mcastResponse struct {
	Jsonrpc string         `json:"jsonrpc"`
	Id      int            `json:"id"`
	Result  interface{}    `json:"result"`
	Error   invalidRequest `json:"error"`
}

type Server struct {
	services *serviceMap
}

func NewServer() *Server {
	return &Server{
		services: new(serviceMap),
	}
}

func (t *Server) Register(receiver interface{}, name string) error {
	return t.services.register(receiver, name)
}

func (t *Server) ListenAndServe(host string, port int) error {
	mcaddr, err := net.ResolveUDPAddr("udp", fmt.Sprintf("%s:%d", host, port))
	if err != nil {
		return err
	}

	l, err := net.ListenMulticastUDP("udp", nil, mcaddr)
	if err != nil {
		return err
	}
	log.Println("[mcast] Started listening for connections")
	for {
		var req mcastRequest
		var response mcastResponse

		buff := make([]byte, MAX_READ_BYTES)
		n, addr, err := l.ReadFromUDP(buff)
		if err != nil {
			log.Println(err)
			continue
		}

		if err := req.FromJSON(buff[0:n]); err != nil {
			response.Error = invalidRequest{
				Code:    ERROR_INVALID_JSON,
				Message: err.Error(),
			}
			t.writeResponse(response, l, addr)
			continue
		}

		response.Id = req.Id
		response.Jsonrpc = JSONRPC_V2

		serviceSpec, methodSpec, errGet := t.services.get(req.Method)
		if errGet != nil {
			response.Error = invalidRequest{
				Code:    ERROR_INVALID_JSON,
				Message: errGet.Error(),
			}
			t.writeResponse(response, l, addr)
			continue
		}

		args := reflect.New(methodSpec.argsType)

		if err := json.Unmarshal(req.Params, args.Interface()); err != nil {
			response.Error = invalidRequest{
				Code:    ERROR_INVALID_REQUEST,
				Message: err.Error(),
			}
			t.writeResponse(response, l, addr)
			continue
		}

		reply := reflect.New(methodSpec.replyType)

		errValue := methodSpec.method.Func.Call([]reflect.Value{
			serviceSpec.rcvr,
			args,
			reply,
		})

		var errResult error
		errInter := errValue[0].Interface()
		if errInter != nil {
			errResult = errInter.(error)
		}

		if errResult == nil {
			response.Result = reply.Interface()
		} else {
			response.Error = invalidRequest{
				Code:    ERROR_METHOD_NOT_EXIST,
				Message: errResult.Error(),
			}
		}

		t.writeResponse(response, l, addr)

	}
}

func (t *Server) writeResponse(res mcastResponse, l *net.UDPConn, addr *net.UDPAddr) {
	if _, err := l.WriteToUDP(res.ToJSON(), addr); err != nil {
		log.Println(err)
	}
}

func (t *mcastRequest) toJSON() []byte {
	buff, _ := json.Marshal(t)
	return buff
}

func (t *mcastRequest) FromJSON(buff []byte) error {
	reader := bytes.NewReader(buff)
	return json.NewDecoder(reader).Decode(t)
}

func (t *mcastResponse) ToJSON() []byte {
	buff, _ := json.Marshal(t)
	return buff
}

func (t *mcastResponse) FromJSON(buff []byte) error {
	reader := bytes.NewReader(buff)
	return json.NewDecoder(reader).Decode(t)
}
