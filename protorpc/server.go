// Copyright 2009 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

/*
	Package rpc provides access to the exported methods of an object across a
	network or other I/O connection.  A server registers an object, making it visible
	as a service with the name of the type of the object.  After registration, exported
	methods of the object will be accessible remotely.  A server may register multiple
	objects (services) of different types but it is an error to register multiple
	objects of the same type.

	Only methods that satisfy these criteria will be made available for remote access;
	other methods will be ignored:

		- the method is exported.
		- the method has two arguments, both exported (or builtin) types.
		- the method's second argument is a pointer.
		- the method has return type error.

	In effect, the method must look schematically like

		func (t *T) MethodName(argType T1, replyType *T2) error

	where T, T1 and T2 can be marshaled by encoding/gob.
	These requirements apply even if a different codec is used.
	(In the future, these requirements may soften for custom codecs.)

	The method's first argument represents the arguments provided by the caller; the
	second argument represents the result parameters to be returned to the caller.
	The method's return value, if non-nil, is passed back as a string that the client
	sees as if created by errors.New.  If an error is returned, the reply parameter
	will not be sent back to the client.

	The server may handle requests on a single connection by calling ServeConn.  More
	typically it will create a network listener and call Accept or, for an HTTP
	listener, HandleHTTP and http.Serve.

	A client wishing to use the service establishes a connection and then invokes
	NewClient on the connection.  The convenience function Dial (DialHTTP) performs
	both steps for a raw network connection (an HTTP connection).  The resulting
	Client object has two methods, Call and Go, that specify the service and method to
	call, a pointer containing the arguments, and a pointer to receive the result
	parameters.

	The Call method waits for the remote call to complete while the Go method
	launches the call asynchronously and signals completion using the Call
	structure's Done channel.

	Unless an explicit codec is set up, package encoding/gob is used to
	transport the data.

	Here is a simple example.  A server wishes to export an object of type Arith:

		package server

		type Args struct {
			A, B int
		}

		type Quotient struct {
			Quo, Rem int
		}

		type Arith int

		func (t *Arith) Multiply(args *Args, reply *int) error {
			*reply = args.A * args.B
			return nil
		}

		func (t *Arith) Divide(args *Args, quo *Quotient) error {
			if args.B == 0 {
				return errors.New("divide by zero")
			}
			quo.Quo = args.A / args.B
			quo.Rem = args.A % args.B
			return nil
		}

	The server calls (for HTTP service):

		arith := new(Arith)
		rpc.Register(arith)
		rpc.HandleHTTP()
		l, e := net.Listen("tcp", ":1234")
		if e != nil {
			log.Fatal("listen error:", e)
		}
		go http.Serve(l, nil)

	At this point, clients can see a service "Arith" with methods "Arith.Multiply" and
	"Arith.Divide".  To invoke one, a client first dials the server:

		client, err := rpc.DialHTTP("tcp", serverAddress + ":1234")
		if err != nil {
			log.Fatal("dialing:", err)
		}

	Then it can make a remote call:

		// Synchronous call
		args := &server.Args{7,8}
		var reply int
		err = client.Call("Arith.Multiply", args, &reply)
		if err != nil {
			log.Fatal("arith error:", err)
		}
		fmt.Printf("Arith: %d*%d=%d", args.A, args.B, reply)

	or

		// Asynchronous call
		quotient := new(Quotient)
		divCall := client.Go("Arith.Divide", args, quotient, nil)
		replyCall := <-divCall.Done	// will be equal to divCall
		// check errors, print, etc.

	A server implementation will often provide a simple, type-safe wrapper for the
	client.
*/
package protorpc

import (
	"bufio"
	// "encoding/gob"
	"errors"
	"io"
	"log"
	"net"
	"net/http"
	"reflect"
	"strings"
	"sync"
	"unicode"
	"unicode/utf8"
)

const (
	// Defaults used by HandleHTTP
	DefaultRPCPath   = "/_goRPC_"
	DefaultDebugPath = "/debug/rpc"
)

// Precompute the reflect type for error.  Can't use error directly
// because Typeof takes an empty interface value.  This is annoying.
var typeOfError = reflect.TypeOf((*error)(nil)).Elem()

type methodType struct {
	sync.Mutex // protects counters
	method     reflect.Method
	ArgType    reflect.Type
	ReplyType  reflect.Type
	numCalls   uint
}

type service struct {
	name   string                 // name of service
	rcvr   reflect.Value          // receiver of methods for the service
	typ    reflect.Type           // type of the receiver
	method map[string]*methodType // registered methods
}

// Request is a header written before every RPC call.  It is used internally
// but documented here as an aid to debugging, such as when analyzing
// network traffic.
// type Request struct {
// 	ServiceMethod string   // format: "Service.Method"
// 	Seq           uint64   // sequence number chosen by client
// 	next          *Request // for free list in Server
// }

// Response is a header written before every RPC return.  It is used internally
// but documented here as an aid to debugging, such as when analyzing
// network traffic.
// type Response struct {
// 	ServiceMethod string    // echoes that of the Request
// 	Seq           uint64    // echoes that of the request
// 	Error         string    // error, if any.
// 	next          *Response // for free list in Server
// }

// Server represents an RPC Server.
type Server struct {
	mu         sync.RWMutex // protects the serviceMap
	serviceMap map[string]*service
	reqLocks   map[*bufio.Reader]*sync.Mutex // protects freeReq
	freeReqs   map[*bufio.Reader]*Request
	reqPool    sync.Pool
	respLocks  map[*bufio.Writer]*sync.Mutex // protects freeResp
	freeResps  map[*bufio.Writer]*Response
	respPool   sync.Pool

	lockPool sync.Pool

	rdBuf sync.Pool
	wrBuf sync.Pool

	codec ServerCodec
}

// NewServer returns a new Server.
func NewServer() *Server {
	return &Server{
		serviceMap: make(map[string]*service),
		reqPool: sync.Pool{
			New: func() interface{} {
				return new(Request)
			},
		},
		respPool: sync.Pool{
			New: func() interface{} {
				return new(Response)
			},
		},
		lockPool: sync.Pool{
			New: func() interface{} {
				return new(sync.Mutex)
			},
		},
		reqLocks:  make(map[*bufio.Reader]*sync.Mutex),
		freeReqs:  make(map[*bufio.Reader]*Request),
		respLocks: make(map[*bufio.Writer]*sync.Mutex),
		freeResps: make(map[*bufio.Writer]*Response),
		codec:     NewPbServerCodec(),
	}
}

// DefaultServer is the default instance of *Server.
var DefaultServer = NewServer()

// Is this an exported - upper case - name?
func isExported(name string) bool {
	rune, _ := utf8.DecodeRuneInString(name)
	return unicode.IsUpper(rune)
}

// Is this type exported or a builtin?
func isExportedOrBuiltinType(t reflect.Type) bool {
	for t.Kind() == reflect.Ptr {
		t = t.Elem()
	}
	// PkgPath will be non-empty even for an exported type,
	// so we need to check the type name as well.
	return isExported(t.Name()) || t.PkgPath() == ""
}

// Register publishes in the server the set of methods of the
// receiver value that satisfy the following conditions:
//	- exported method
//	- two arguments, both of exported type
//	- the second argument is a pointer
//	- one return value, of type error
// It returns an error if the receiver is not an exported type or has
// no suitable methods. It also logs the error using package log.
// The client accesses each method using a string of the form "Type.Method",
// where Type is the receiver's concrete type.
func (server *Server) Register(rcvr interface{}) error {
	return server.register(rcvr, "", false)
}

// RegisterName is like Register but uses the provided name for the type
// instead of the receiver's concrete type.
func (server *Server) RegisterName(name string, rcvr interface{}) error {
	return server.register(rcvr, name, true)
}

func (server *Server) register(rcvr interface{}, name string, useName bool) error {
	server.mu.Lock()
	defer server.mu.Unlock()
	if server.serviceMap == nil {
		server.serviceMap = make(map[string]*service)
	}
	s := new(service)
	s.typ = reflect.TypeOf(rcvr)
	s.rcvr = reflect.ValueOf(rcvr)
	sname := reflect.Indirect(s.rcvr).Type().Name()
	if useName {
		sname = name
	}
	if sname == "" {
		s := "rpc.Register: no service name for type " + s.typ.String()
		log.Print(s)
		return errors.New(s)
	}
	if !isExported(sname) && !useName {
		s := "rpc.Register: type " + sname + " is not exported"
		log.Print(s)
		return errors.New(s)
	}
	if _, present := server.serviceMap[sname]; present {
		return errors.New("rpc: service already defined: " + sname)
	}
	s.name = sname

	// Install the methods
	s.method = suitableMethods(s.typ, true)

	if len(s.method) == 0 {
		str := ""

		// To help the user, see if a pointer receiver would work.
		method := suitableMethods(reflect.PtrTo(s.typ), false)
		if len(method) != 0 {
			str = "rpc.Register: type " + sname + " has no exported methods of suitable type (hint: pass a pointer to value of that type)"
		} else {
			str = "rpc.Register: type " + sname + " has no exported methods of suitable type"
		}
		log.Print(str)
		return errors.New(str)
	}
	server.serviceMap[s.name] = s
	return nil
}

// suitableMethods returns suitable Rpc methods of typ, it will report
// error using log if reportErr is true.
func suitableMethods(typ reflect.Type, reportErr bool) map[string]*methodType {
	methods := make(map[string]*methodType)
	for m := 0; m < typ.NumMethod(); m++ {
		method := typ.Method(m)
		mtype := method.Type
		mname := method.Name
		// Method must be exported.
		if method.PkgPath != "" {
			continue
		}
		// Method needs three ins: receiver, *args, *reply.
		if mtype.NumIn() != 3 {
			if reportErr {
				log.Println("method", mname, "has wrong number of ins:", mtype.NumIn())
			}
			continue
		}
		// First arg need not be a pointer.
		argType := mtype.In(1)
		if !isExportedOrBuiltinType(argType) {
			if reportErr {
				log.Println(mname, "argument type not exported:", argType)
			}
			continue
		}
		// Second arg must be a pointer.
		replyType := mtype.In(2)
		if replyType.Kind() != reflect.Ptr {
			if reportErr {
				log.Println("method", mname, "reply type not a pointer:", replyType)
			}
			continue
		}
		// Reply type must be exported.
		if !isExportedOrBuiltinType(replyType) {
			if reportErr {
				log.Println("method", mname, "reply type not exported:", replyType)
			}
			continue
		}
		// Method needs one out.
		if mtype.NumOut() != 1 {
			if reportErr {
				log.Println("method", mname, "has wrong number of outs:", mtype.NumOut())
			}
			continue
		}
		// The return type of the method must be error.
		if returnType := mtype.Out(0); returnType != typeOfError {
			if reportErr {
				log.Println("method", mname, "returns", returnType.String(), "not error")
			}
			continue
		}
		methods[mname] = &methodType{method: method, ArgType: argType, ReplyType: replyType}
	}
	return methods
}

// A value sent as a placeholder for the server's response value when the server
// receives an invalid request. It is never decoded by the client since the Response
// contains an error when it is used.
var invalidRequest = struct{}{}

func (server *Server) sendResponse(wg *sync.WaitGroup, sending *sync.Mutex, req *Request, reply interface{}, w *bufio.Writer, c io.Closer, errmsg string) {
	resp := server.getResponse(w)
	// Encode the response header
	resp.ServiceMethod = req.ServiceMethod
	if errmsg != "" {
		resp.Error = errmsg
		reply = invalidRequest
	}
	resp.Seq = req.Seq
	sending.Lock()
	err := server.codec.WriteResponse(w, c, resp, reply)
	if debugLog && err != nil {
		log.Println("rpc: writing response:", err)
	}
	sending.Unlock()
	wg.Done()
	server.freeResponse(w, resp)
}

func (m *methodType) NumCalls() (n uint) {
	m.Lock()
	n = m.numCalls
	m.Unlock()
	return n
}

func (s *service) call(server *Server, wg *sync.WaitGroup, sending *sync.Mutex, mtype *methodType, req *Request, argv, replyv reflect.Value, r *bufio.Reader, w *bufio.Writer, c io.Closer) {
	mtype.Lock()
	mtype.numCalls++
	mtype.Unlock()
	function := mtype.method.Func
	// Invoke the method, providing a new value for the reply.
	returnValues := function.Call([]reflect.Value{s.rcvr, argv, replyv})
	// The return value for the method is an error.
	errInter := returnValues[0].Interface()
	errmsg := ""
	if errInter != nil {
		errmsg = errInter.(error).Error()
	}
	server.sendResponse(wg, sending, req, replyv.Interface(), w, c, errmsg)
	server.freeRequest(r, req)
}

// type gobServerCodec struct {
// 	rwc    io.ReadWriteCloser
// 	dec    *gob.Decoder
// 	enc    *gob.Encoder
// 	encBuf *bufio.Writer
// 	closed bool
// }

// func (c *gobServerCodec) ReadRequestHeader(r *Request) error {
// 	return c.dec.Decode(r)
// }

// func (c *gobServerCodec) ReadRequestBody(body interface{}) error {
// 	return c.dec.Decode(body)
// }

// func (c *gobServerCodec) WriteResponse(r *Response, body interface{}) (err error) {
// 	if err = c.enc.Encode(r); err != nil {
// 		if c.encBuf.Flush() == nil {
// 			// Gob couldn't encode the header. Should not happen, so if it does,
// 			// shut down the connection to signal that the connection is broken.
// 			log.Println("rpc: gob error encoding response:", err)
// 			c.Close()
// 		}
// 		return
// 	}
// 	if err = c.enc.Encode(body); err != nil {
// 		if c.encBuf.Flush() == nil {
// 			// Was a gob problem encoding the body but the header has been written.
// 			// Shut down the connection to signal that the connection is broken.
// 			log.Println("rpc: gob error encoding body:", err)
// 			c.Close()
// 		}
// 		return
// 	}
// 	return c.encBuf.Flush()
// }

// func (c *gobServerCodec) Close() error {
// 	if c.closed {
// 		// Only call c.rwc.Close once; otherwise the semantics are undefined.
// 		return nil
// 	}
// 	c.closed = true
// 	return c.rwc.Close()
// }

// getReadBuf return a bufio.Reader whit connection.
func (server *Server) getReadBuf(conn io.ReadWriter) *bufio.Reader {
	if v := server.rdBuf.Get(); v != nil {
		r := v.(*bufio.Reader)
		r.Reset(conn)
		return r
	}
	return bufio.NewReader(conn)
}

// getReadBuf return a bufio.Write whit connection.
func (server *Server) getWriteBuf(conn io.ReadWriter) *bufio.Writer {
	if v := server.wrBuf.Get(); v != nil {
		r := v.(*bufio.Writer)
		r.Reset(conn)
		return r
	}
	return bufio.NewWriter(conn)
}

// putReadBuf put a bufio.Reader into pool.
func (server *Server) putReadBuf(r *bufio.Reader) {
	r.Reset(nil)
	server.rdBuf.Put(r)
}

// putWriteBuf put a bufio.Write into pool.
func (server *Server) putWriteBuf(w *bufio.Writer) {
	w.Reset(nil)
	server.wrBuf.Put(w)
}

// initRead init a Request and request lock.
func (server *Server) initRequest(r *bufio.Reader) {
	server.mu.Lock()
	server.reqLocks[r] = server.lockPool.Get().(*sync.Mutex)
	req := server.reqPool.Get()
	log.Print("request:", req)
	server.freeReqs[r] = req.(*Request)
	server.mu.Unlock()
}

// initWrite init a Response and response lock.
func (server *Server) initResponse(w *bufio.Writer) {
	server.mu.Lock()
	l := server.lockPool.Get()
	server.respLocks[w] = l.(*sync.Mutex)
	resp := server.respPool.Get()
	log.Print("response:", resp)
	server.freeResps[w] = resp.(*Response)
	server.mu.Unlock()
}

// delRead delete a Request and request lock.
func (server *Server) delRequest(r *bufio.Reader) {
	server.mu.Lock()
	// pur reader in pool
	server.putReadBuf(r)
	// put req in pool
	req := server.freeReqs[r]
	delete(server.freeReqs, r)
	server.reqPool.Put(req)
	// put req lock in pool
	reql := server.reqLocks[r]
	delete(server.reqLocks, r)
	server.lockPool.Put(reql)
	server.mu.Unlock()
}

// delWrite delete a Response and response lock.
func (server *Server) delResponse(w *bufio.Writer) {
	server.mu.Lock()
	server.putWriteBuf(w)
	// put resp in pool
	resp := server.freeResps[w]
	delete(server.freeResps, w)
	server.reqPool.Put(resp)
	// put resp lock in pool
	respl := server.respLocks[w]
	delete(server.respLocks, w)
	server.lockPool.Put(respl)
	server.mu.Unlock()
}

// ServeConn runs the server on a single connection.
// ServeConn blocks, serving the connection until the client hangs up.
// The caller typically invokes ServeConn in a go statement.
// ServeConn uses the gob wire format (see package gob) on the
// connection.  To use an alternate codec, use ServeCodec.
// func (server *Server) ServeConn(conn io.ReadWriteCloser) {
// 	buf := bufio.NewWriter(conn)
// 	srv := &gobServerCodec{
// 		rwc:    conn,
// 		dec:    gob.NewDecoder(conn),
// 		enc:    gob.NewEncoder(buf),
// 		encBuf: buf,
// 	}
// 	server.ServeCodec(srv)
// }

func (server *Server) ServeConn(conn io.ReadWriteCloser) {
	r := server.getReadBuf(conn)
	w := server.getWriteBuf(conn)
	server.initRequest(r)
	server.initResponse(w)
	sending := server.respLocks[w]
	wg := new(sync.WaitGroup)
	for {
		service, mtype, req, argv, replyv, keepReading, err := server.readRequest(r)
		wg.Add(1)
		if err != nil {
			if debugLog && err != io.EOF {
				log.Println("rpc:", err)
			}
			if !keepReading {
				break
			}
			// send a response if we actually managed to read a header.
			if req != nil {
				server.sendResponse(wg, sending, req, invalidRequest, w, conn, err.Error())
				server.freeRequest(r, req)
			}
			continue
		}
		go service.call(server, wg, sending, mtype, req, argv, replyv, r, w, conn)
	}
	wg.Wait()
	server.delRequest(r)
	server.delResponse(w)
}

// ServeCodec is like ServeConn but uses the specified codec to
// decode requests and encode responses.
//func (server *Server) ServeCodec(r *bufio.Reader, w *bufio.Writer, c io.Closer) {
//	sending := new(sync.Mutex)
//	for {
//		service, mtype, req, argv, replyv, keepReading, err := server.readRequest(r)
//		if err != nil {
//			if debugLog && err != io.EOF {
//				log.Println("rpc:", err)
//			}
//			if !keepReading {
//				break
//			}
//			// send a response if we actually managed to read a header.
//			if req != nil {
//				server.sendResponse(sending, req, invalidRequest, w, c, err.Error())
//				server.freeRequest(req)
//			}
//			continue
//		}
//		go service.call(server, sending, mtype, req, argv, replyv, w, c)
//	}
//	server.putReadBuf(r)
//	server.putWriteBuf(w)
//}

// ServeRequest is like ServeCodec but synchronously serves a single request.
// It does not close the codec upon completion.
func (server *Server) ServeRequest(conn io.ReadWriter) error {
	r := server.getReadBuf(conn)
	w := server.getWriteBuf(conn)
	sending := new(sync.Mutex)
	wg := new(sync.WaitGroup)
	service, mtype, req, argv, replyv, keepReading, err := server.readRequest(r)
	wg.Add(1)
	if err != nil {
		if !keepReading {
			server.putReadBuf(r)
			server.putWriteBuf(w)
			return err
		}
		// send a response if we actually managed to read a header.
		if req != nil {
			server.sendResponse(wg, sending, req, invalidRequest, w, nil, err.Error())
			server.freeRequest(r, req)
		}
		server.putReadBuf(r)
		server.putWriteBuf(w)
		return err
	}
	service.call(server, wg, sending, mtype, req, argv, replyv, r, w, nil)
	wg.Done()
	server.putReadBuf(r)
	server.putWriteBuf(w)
	return nil
}

func (server *Server) getRequest(r *bufio.Reader) *Request {
	// server.reqLock.Lock()
	// req := server.freeReq
	// if req == nil {
	// 	req = new(Request)
	// } else {
	// 	server.freeReq = req.next
	// 	*req = Request{}
	// }
	// server.reqLock.Unlock()
	// return req
	var req *Request
	reql, ok := server.reqLocks[r]
	if ok {
		reql.Lock()
		req = server.freeReqs[r]
		if req == nil {
			req = new(Request)
		} else {
			server.freeReqs[r] = req.Next
			*req = Request{}
		}
		reql.Unlock()
	} else {
		req = server.reqPool.Get().(*Request)
	}
	return req
}

func (server *Server) freeRequest(r *bufio.Reader, req *Request) {
	// server.reqLock.Lock()
	// req.next = server.freeReq
	// server.freeReq = req
	// server.reqLock.Unlock()
	reql, ok := server.reqLocks[r]
	if ok {
		reql.Lock()
		req.Next = server.freeReqs[r]
		server.freeReqs[r] = req
		reql.Unlock()
	} else {
		server.reqPool.Put(req)
	}
}

func (server *Server) getResponse(w *bufio.Writer) *Response {
	// server.respLock.Lock()
	// resp := server.freeResp
	// if resp == nil {
	// 	resp = new(Response)
	// } else {
	// 	server.freeResp = resp.next
	// 	*resp = Response{}
	// }
	// server.respLock.Unlock()
	// return resp
	var resp *Response
	respl, ok := server.respLocks[w]
	if ok {
		respl.Lock()
		resp = server.freeResps[w]
		if resp == nil {
			resp = new(Response)
		} else {
			server.freeResps[w] = resp.Next
			*resp = Response{}
		}
		respl.Unlock()
	} else {
		resp = server.respPool.Get().(*Response)
	}
	return resp
}

func (server *Server) freeResponse(w *bufio.Writer, resp *Response) {
	// server.respLock.Lock()
	// resp.next = server.freeResp
	// server.freeResp = resp
	// server.respLock.Unlock()
	respl, ok := server.respLocks[w]
	if ok {
		respl.Lock()
		resp.Next = server.freeResps[w]
		server.freeResps[w] = resp
		respl.Unlock()
	} else {
		server.respPool.Put(resp)
	}
}

func (server *Server) readRequest(r *bufio.Reader) (service *service, mtype *methodType, req *Request, argv, replyv reflect.Value, keepReading bool, err error) {
	service, mtype, req, keepReading, err = server.readRequestHeader(r)
	if err != nil {
		if !keepReading {
			return
		}
		// discard body
		//		codec.ReadRequestBody(nil)
		server.codec.ReadRequestBody(r, nil)
		return
	}

	// Decode the argument value.
	argIsValue := false // if true, need to indirect before calling.
	if mtype.ArgType.Kind() == reflect.Ptr {
		argv = reflect.New(mtype.ArgType.Elem())
	} else {
		argv = reflect.New(mtype.ArgType)
		argIsValue = true
	}
	// argv guaranteed to be a pointer now.
	if err = server.codec.ReadRequestBody(r, argv.Interface()); err != nil {
		return
	}
	if argIsValue {
		argv = argv.Elem()
	}

	replyv = reflect.New(mtype.ReplyType.Elem())
	return
}

func (server *Server) readRequestHeader(r *bufio.Reader) (service *service, mtype *methodType, req *Request, keepReading bool, err error) {
	// Grab the request header.
	req = server.getRequest(r)
	err = server.codec.ReadRequestHeader(r, req)
	if err != nil {
		req = nil
		if err == io.EOF || err == io.ErrUnexpectedEOF {
			return
		}
		err = errors.New("rpc: server cannot decode request: " + err.Error())
		return
	}

	// We read the header successfully.  If we see an error now,
	// we can still recover and move on to the next request.
	keepReading = true

	dot := strings.LastIndex(req.ServiceMethod, ".")
	if dot < 0 {
		err = errors.New("rpc: service/method request ill-formed: " + req.ServiceMethod)
		return
	}
	serviceName := req.ServiceMethod[:dot]
	methodName := req.ServiceMethod[dot+1:]

	// Look up the request.
	server.mu.RLock()
	service = server.serviceMap[serviceName]
	server.mu.RUnlock()
	if service == nil {
		err = errors.New("rpc: can't find service " + req.ServiceMethod)
		return
	}
	mtype = service.method[methodName]
	if mtype == nil {
		err = errors.New("rpc: can't find method " + req.ServiceMethod)
	}
	return
}

// Accept accepts connections on the listener and serves requests
// for each incoming connection.  Accept blocks; the caller typically
// invokes it in a go statement.
func (server *Server) Accept(lis net.Listener) {
	for {
		conn, err := lis.Accept()
		if err != nil {
			log.Fatal("rpc.Serve: accept:", err.Error()) // TODO(r): exit?
		}
		go server.ServeConn(conn)
	}
}

// Register publishes the receiver's methods in the DefaultServer.
func Register(rcvr interface{}) error { return DefaultServer.Register(rcvr) }

// RegisterName is like Register but uses the provided name for the type
// instead of the receiver's concrete type.
func RegisterName(name string, rcvr interface{}) error {
	return DefaultServer.RegisterName(name, rcvr)
}

// A ServerCodec implements reading of RPC requests and writing of
// RPC responses for the server side of an RPC session.
// The server calls ReadRequestHeader and ReadRequestBody in pairs
// to read requests from the connection, and it calls WriteResponse to
// write a response back.  The server calls Close when finished with the
// connection. ReadRequestBody may be called with a nil
// argument to force the body of the request to be read and discarded.
type ServerCodec interface {
	ReadRequestHeader(*bufio.Reader, *Request) error
	ReadRequestBody(*bufio.Reader, interface{}) error
	// WriteResponse must be safe for concurrent use by multiple goroutines.
	WriteResponse(*bufio.Writer, io.Closer, *Response, interface{}) error
}

// ServeConn runs the DefaultServer on a single connection.
// ServeConn blocks, serving the connection until the client hangs up.
// The caller typically invokes ServeConn in a go statement.
// ServeConn uses the gob wire format (see package gob) on the
// connection.  To use an alternate codec, use ServeCodec.
func ServeConn(conn io.ReadWriteCloser) {
	DefaultServer.ServeConn(conn)
}

// ServeCodec is like ServeConn but uses the specified codec to
// decode requests and encode responses.
//func ServeCodec(codec ServerCodec) {
//	DefaultServer.ServeCodec(codec)
//}

// ServeRequest is like ServeCodec but synchronously serves a single request.
// It does not close the codec upon completion.
func ServeRequest(conn io.ReadWriter) error {
	return DefaultServer.ServeRequest(conn)
}

// Accept accepts connections on the listener and serves requests
// to DefaultServer for each incoming connection.
// Accept blocks; the caller typically invokes it in a go statement.
func Accept(lis net.Listener) { DefaultServer.Accept(lis) }

// Can connect to RPC service using HTTP CONNECT to rpcPath.
var connected = "200 Connected to Go RPC"

// ServeHTTP implements an http.Handler that answers RPC requests.
func (server *Server) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	if req.Method != "CONNECT" {
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		w.WriteHeader(http.StatusMethodNotAllowed)
		io.WriteString(w, "405 must CONNECT\n")
		return
	}
	conn, _, err := w.(http.Hijacker).Hijack()
	if err != nil {
		log.Print("rpc hijacking ", req.RemoteAddr, ": ", err.Error())
		return
	}
	io.WriteString(conn, "HTTP/1.0 "+connected+"\n\n")
	server.ServeConn(conn)
}

// HandleHTTP registers an HTTP handler for RPC messages on rpcPath,
// and a debugging handler on debugPath.
// It is still necessary to invoke http.Serve(), typically in a go statement.
func (server *Server) HandleHTTP(rpcPath, debugPath string) {
	http.Handle(rpcPath, server)
	http.Handle(debugPath, debugHTTP{server})
}

// HandleHTTP registers an HTTP handler for RPC messages to DefaultServer
// on DefaultRPCPath and a debugging handler on DefaultDebugPath.
// It is still necessary to invoke http.Serve(), typically in a go statement.
func HandleHTTP() {
	DefaultServer.HandleHTTP(DefaultRPCPath, DefaultDebugPath)
}
