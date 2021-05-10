# 【Go从入门到精通】Day02 造一个自己的路由器引擎


## 探索使用标准库运行HTTP服务程序

### 用标准库提供的方式运行HTTP服务程序

```go
package main

import (
	"fmt"
	"html"
	"log"
	"net/http"
)

func main() {
	http.HandleFunc("/", indexHandler)
	http.HandleFunc("/hello", helloHandler)
	log.Fatal(http.ListenAndServe(":8989", nil))
}

// handler echoes r.URL.Path
func indexHandler(w http.ResponseWriter, req *http.Request) {
	fmt.Fprintf(w, "URL.Path = %q\n", html.EscapeString(req.URL.Path))
}

// handler echoes r.URL.Header
func helloHandler(w http.ResponseWriter, req *http.Request) {
	for k, v := range req.Header {
		fmt.Fprintf(w, "Header[%q] = %q\n", k, v)
	}
}

```


```
% curl http://localhost:8989      
URL.Path = "/"
```

```
% curl http://localhost:8989/hello
Header["User-Agent"] = ["curl/7.64.1"]
Header["Accept"] = ["*/*"]
```

### 一起看一下 http.Handler 接口定义

```go
// A Handler responds to an HTTP request.
//
// ServeHTTP should write reply headers and data to the ResponseWriter
// and then return. Returning signals that the request is finished; it
// is not valid to use the ResponseWriter or read from the
// Request.Body after or concurrently with the completion of the
// ServeHTTP call.
//
// Depending on the HTTP client software, HTTP protocol version, and
// any intermediaries between the client and the Go server, it may not
// be possible to read from the Request.Body after writing to the
// ResponseWriter. Cautious handlers should read the Request.Body
// first, and then reply.
//
// Except for reading the body, handlers should not modify the
// provided Request.
//
// If ServeHTTP panics, the server (the caller of ServeHTTP) assumes
// that the effect of the panic was isolated to the active request.
// It recovers the panic, logs a stack trace to the server error log,
// and either closes the network connection or sends an HTTP/2
// RST_STREAM, depending on the HTTP protocol. To abort a handler so
// the client sees an interrupted response but the server doesn't log
// an error, panic with the value ErrAbortHandler.
type Handler interface {
	ServeHTTP(ResponseWriter, *Request)
}
```

## 尝试改变：使用统一的Handler

将上一节的程序作如下改写：

```go
package main

import (
	"fmt"
	"html"
	"log"
	"net/http"
)

type Engin struct{}

func (e *Engin) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	switch req.URL.Path {
	case "/":
		fmt.Fprintf(w, "URL.Path = %q\n", html.EscapeString(req.URL.Path))
	case "/hello":
		for k, v := range req.Header {
			fmt.Fprintf(w, "Header[%q] = %q\n", k, v)
		}
	default:
		fmt.Fprintf(w, "404 NOT FOUND: %s\n", req.URL)
	}
}

func main() {
	log.Fatal(http.ListenAndServe(":8989", &Engin{}))
}

```

1. 定义了 Engin{} 结构体并实现 http.Handler 接口，作为全局的处理器。
2. 直接在 http.ListenAndServe() 的参数里传入这个全局唯一的处理器


### 观察 http.Request{} 结构体


```go

// A Request represents an HTTP request received by a server
// or to be sent by a client.
//
// The field semantics differ slightly between client and server
// usage. In addition to the notes on the fields below, see the
// documentation for Request.Write and RoundTripper.
type Request struct {
	// Method specifies the HTTP method (GET, POST, PUT, etc.).
	// For client requests, an empty string means GET.
	//
	// Go's HTTP client does not support sending a request with
	// the CONNECT method. See the documentation on Transport for
	// details.
	Method string

	// URL specifies either the URI being requested (for server
	// requests) or the URL to access (for client requests).
	//
	// For server requests, the URL is parsed from the URI
	// supplied on the Request-Line as stored in RequestURI.  For
	// most requests, fields other than Path and RawQuery will be
	// empty. (See RFC 7230, Section 5.3)
	//
	// For client requests, the URL's Host specifies the server to
	// connect to, while the Request's Host field optionally
	// specifies the Host header value to send in the HTTP
	// request.
	URL *url.URL

	// The protocol version for incoming server requests.
	//
	// For client requests, these fields are ignored. The HTTP
	// client code always uses either HTTP/1.1 or HTTP/2.
	// See the docs on Transport for details.
	Proto      string // "HTTP/1.0"
	ProtoMajor int    // 1
	ProtoMinor int    // 0

	// Header contains the request header fields either received
	// by the server or to be sent by the client.
	//
	// If a server received a request with header lines,
	//
	//	Host: example.com
	//	accept-encoding: gzip, deflate
	//	Accept-Language: en-us
	//	fOO: Bar
	//	foo: two
	//
	// then
	//
	//	Header = map[string][]string{
	//		"Accept-Encoding": {"gzip, deflate"},
	//		"Accept-Language": {"en-us"},
	//		"Foo": {"Bar", "two"},
	//	}
	//
	// For incoming requests, the Host header is promoted to the
	// Request.Host field and removed from the Header map.
	//
	// HTTP defines that header names are case-insensitive. The
	// request parser implements this by using CanonicalHeaderKey,
	// making the first character and any characters following a
	// hyphen uppercase and the rest lowercase.
	//
	// For client requests, certain headers such as Content-Length
	// and Connection are automatically written when needed and
	// values in Header may be ignored. See the documentation
	// for the Request.Write method.
	Header Header

	// Body is the request's body.
	//
	// For client requests, a nil body means the request has no
	// body, such as a GET request. The HTTP Client's Transport
	// is responsible for calling the Close method.
	//
	// For server requests, the Request Body is always non-nil
	// but will return EOF immediately when no body is present.
	// The Server will close the request body. The ServeHTTP
	// Handler does not need to.
	//
	// Body must allow Read to be called concurrently with Close.
	// In particular, calling Close should unblock a Read waiting
	// for input.
	Body io.ReadCloser

	// GetBody defines an optional func to return a new copy of
	// Body. It is used for client requests when a redirect requires
	// reading the body more than once. Use of GetBody still
	// requires setting Body.
	//
	// For server requests, it is unused.
	GetBody func() (io.ReadCloser, error)

	// ContentLength records the length of the associated content.
	// The value -1 indicates that the length is unknown.
	// Values >= 0 indicate that the given number of bytes may
	// be read from Body.
	//
	// For client requests, a value of 0 with a non-nil Body is
	// also treated as unknown.
	ContentLength int64

	// TransferEncoding lists the transfer encodings from outermost to
	// innermost. An empty list denotes the "identity" encoding.
	// TransferEncoding can usually be ignored; chunked encoding is
	// automatically added and removed as necessary when sending and
	// receiving requests.
	TransferEncoding []string

	// Close indicates whether to close the connection after
	// replying to this request (for servers) or after sending this
	// request and reading its response (for clients).
	//
	// For server requests, the HTTP server handles this automatically
	// and this field is not needed by Handlers.
	//
	// For client requests, setting this field prevents re-use of
	// TCP connections between requests to the same hosts, as if
	// Transport.DisableKeepAlives were set.
	Close bool

	// For server requests, Host specifies the host on which the
	// URL is sought. For HTTP/1 (per RFC 7230, section 5.4), this
	// is either the value of the "Host" header or the host name
	// given in the URL itself. For HTTP/2, it is the value of the
	// ":authority" pseudo-header field.
	// It may be of the form "host:port". For international domain
	// names, Host may be in Punycode or Unicode form. Use
	// golang.org/x/net/idna to convert it to either format if
	// needed.
	// To prevent DNS rebinding attacks, server Handlers should
	// validate that the Host header has a value for which the
	// Handler considers itself authoritative. The included
	// ServeMux supports patterns registered to particular host
	// names and thus protects its registered Handlers.
	//
	// For client requests, Host optionally overrides the Host
	// header to send. If empty, the Request.Write method uses
	// the value of URL.Host. Host may contain an international
	// domain name.
	Host string

	// Form contains the parsed form data, including both the URL
	// field's query parameters and the PATCH, POST, or PUT form data.
	// This field is only available after ParseForm is called.
	// The HTTP client ignores Form and uses Body instead.
	Form url.Values

	// PostForm contains the parsed form data from PATCH, POST
	// or PUT body parameters.
	//
	// This field is only available after ParseForm is called.
	// The HTTP client ignores PostForm and uses Body instead.
	PostForm url.Values

	// MultipartForm is the parsed multipart form, including file uploads.
	// This field is only available after ParseMultipartForm is called.
	// The HTTP client ignores MultipartForm and uses Body instead.
	MultipartForm *multipart.Form

	// Trailer specifies additional headers that are sent after the request
	// body.
	//
	// For server requests, the Trailer map initially contains only the
	// trailer keys, with nil values. (The client declares which trailers it
	// will later send.)  While the handler is reading from Body, it must
	// not reference Trailer. After reading from Body returns EOF, Trailer
	// can be read again and will contain non-nil values, if they were sent
	// by the client.
	//
	// For client requests, Trailer must be initialized to a map containing
	// the trailer keys to later send. The values may be nil or their final
	// values. The ContentLength must be 0 or -1, to send a chunked request.
	// After the HTTP request is sent the map values can be updated while
	// the request body is read. Once the body returns EOF, the caller must
	// not mutate Trailer.
	//
	// Few HTTP clients, servers, or proxies support HTTP trailers.
	Trailer Header

	// RemoteAddr allows HTTP servers and other software to record
	// the network address that sent the request, usually for
	// logging. This field is not filled in by ReadRequest and
	// has no defined format. The HTTP server in this package
	// sets RemoteAddr to an "IP:port" address before invoking a
	// handler.
	// This field is ignored by the HTTP client.
	RemoteAddr string

	// RequestURI is the unmodified request-target of the
	// Request-Line (RFC 7230, Section 3.1.1) as sent by the client
	// to a server. Usually the URL field should be used instead.
	// It is an error to set this field in an HTTP client request.
	RequestURI string

	// TLS allows HTTP servers and other software to record
	// information about the TLS connection on which the request
	// was received. This field is not filled in by ReadRequest.
	// The HTTP server in this package sets the field for
	// TLS-enabled connections before invoking a handler;
	// otherwise it leaves the field nil.
	// This field is ignored by the HTTP client.
	TLS *tls.ConnectionState

	// Cancel is an optional channel whose closure indicates that the client
	// request should be regarded as canceled. Not all implementations of
	// RoundTripper may support Cancel.
	//
	// For server requests, this field is not applicable.
	//
	// Deprecated: Set the Request's context with NewRequestWithContext
	// instead. If a Request's Cancel field and context are both
	// set, it is undefined whether Cancel is respected.
	Cancel <-chan struct{}

	// Response is the redirect response which caused this request
	// to be created. This field is only populated during client
	// redirects.
	Response *Response

	// ctx is either the client or server context. It should only
	// be modified via copying the whole Request using WithContext.
	// It is unexported to prevent people from using Context wrong
	// and mutating the contexts held by callers of the same request.
	ctx context.Context
}
```

通过以上源代码我们能看到 通过 http.Request{} 结构体，我们可以获得HTTP请求的 Method、Header、Body、Proto等等所有信息。


## 开始打造 Vodka Engine 路由框架

有了标准库中的 http.ListenAndServe()函数，http.Handler 接口并且结合我们对 http.Request 的理解，
很容易想到我们再造一个独一无二的 Handler 来处理所有请求，然后我们自己处理路由逻辑。

### 重写 Engine{} 结构体


```go
package vodka

import (
	"fmt"
	"net/http"
)

// HandlerFunc 定义了 Engine 自己的处理器
type HandlerFunc func(w http.ResponseWriter, r *http.Request)

// Engine 定义了 Vodka 框架的引擎
type Engine struct {
	router map[string]HandlerFunc
}

```



### 让 Engine 实现 http.Handler 接口


```go 


package vodka

import (
	"fmt"
	"net/http"
)

// HandlerFunc 定义了 Engine 自己的处理器
type HandlerFunc func(w http.ResponseWriter, r *http.Request)

// Engine 定义了 Vodka 框架的引擎
type Engine struct {
	router map[string]HandlerFunc
}


func (e *Engine) ServeHTTP(w http.ResponseWriter, req *http.Request) {

	key := req.Method + "-" + req.URL.Path
	if handler, ok := e.router[key]; ok {
		handler(w, req)
	} else {
		fmt.Fprintf(w, "404 NOT FOUND: %s\n", req.URL)
	}
}
```

### 给 Engine 增加 New方法

```go 
package vodka

import (
	"fmt"
	"net/http"
)

// HandlerFunc 定义了 Engine 自己的处理器
type HandlerFunc func(w http.ResponseWriter, r *http.Request)

// Engine 定义了 Vodka 框架的引擎
type Engine struct {
	router map[string]HandlerFunc
}

func New() *Engine {
	return &Engine{router: make(map[string]HandlerFunc)}
}

func (e *Engine) ServeHTTP(w http.ResponseWriter, req *http.Request) {

	key := req.Method + "-" + req.URL.Path
	if handler, ok := e.router[key]; ok {
		handler(w, req)
	} else {
		fmt.Fprintf(w, "404 NOT FOUND: %s\n", req.URL)
	}
}

```

### 给 Engine 增加 addRouter() 私有方法和 GET() POST() 公有方法

```go 
package vodka

import (
	"fmt"
	"net/http"
)

// HandlerFunc 定义了 Engine 自己的处理器
type HandlerFunc func(w http.ResponseWriter, r *http.Request)

// Engine 定义了 Vodka 框架的引擎
type Engine struct {
	router map[string]HandlerFunc
}

func New() *Engine {
	return &Engine{router: make(map[string]HandlerFunc)}
}

func (e *Engine) addRouter(method, pattern string, handler HandlerFunc) {
	key := method + "-" + pattern
	e.router[key] = handler
}

func (e *Engine) GET(pattern string, handler HandlerFunc) {
	e.addRouter("GET", pattern, handler)
}

func (e *Engine) POST(pattern string, handler HandlerFunc) {
	e.addRouter("POST", pattern, handler)
}

func (e *Engine) ServeHTTP(w http.ResponseWriter, req *http.Request) {

	key := req.Method + "-" + req.URL.Path
	if handler, ok := e.router[key]; ok {
		handler(w, req)
	} else {
		fmt.Fprintf(w, "404 NOT FOUND: %s\n", req.URL)
	}
}
```

### 给 Engine 增加 Run() 方法

```go
// Run 调用**标准库**的 `http.ListenAndServe` 方法启动服务器程序
// 因为 **Engine** 实现了 `http.Handler` 接口 所以这里的第二个参数可以直接传 **Engine**。
func (e *Engine) Run(addr string) error {
	return http.ListenAndServe(addr, e)
}

```

### 完整的 engine.go

```go
package vodka

import (
	"fmt"
	"net/http"
)

// HandlerFunc 定义了 Engine 自己的处理器。
type HandlerFunc func(w http.ResponseWriter, r *http.Request)

// Engine 定义了 Vodka 框架的引擎。
type Engine struct {
	router map[string]HandlerFunc
}

func New() *Engine {
	return &Engine{router: make(map[string]HandlerFunc)}
}

func (e *Engine) addRouter(method, pattern string, handler HandlerFunc) {
	key := method + "-" + pattern
	e.router[key] = handler
}

func (e *Engine) GET(pattern string, handler HandlerFunc) {
	e.addRouter("GET", pattern, handler)
}

func (e *Engine) POST(pattern string, handler HandlerFunc) {
	e.addRouter("POST", pattern, handler)
}

// Run 调用**标准库**的 `http.ListenAndServe` 方法启动服务器程序。
// 因为 **Engine** 实现了 `http.Handler` 接口 所以这里的第二个参数可以直接传 **Engine**。
func (e *Engine) Run(addr string) error {
	return http.ListenAndServe(addr, e)
}

// ServeHTTP 实现了 `http.Handler` 接口。
func (e *Engine) ServeHTTP(w http.ResponseWriter, req *http.Request) {

	key := req.Method + "-" + req.URL.Path
	if handler, ok := e.router[key]; ok {
		handler(w, req)
	} else {
		fmt.Fprintf(w, "404 NOT FOUND: %s\n", req.URL)
	}
}

```


### 测试

```
% curl http://localhost:8989/hello
Header["User-Agent"] = ["curl/7.64.1"]
Header["Accept"] = ["*/*"]
% curl http://localhost:8989      
URL.Path = "/"
% curl http://localhost:8989/panda
404 NOT FOUND: /panda

```