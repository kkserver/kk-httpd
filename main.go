package main

import (
	"fmt"
	"github.com/kkserver/kk-lib/kk"
	"github.com/kkserver/kk-lib/kk/inifile"
	"github.com/kkserver/kk-lib/kk/json"
	"io"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"
)

const VERSION = "1.0.0"

type Httpd struct {
	Name        string
	Address     string
	HttpAddress string
	Alias       string
	Options     map[string]interface{}
}

func main() {

	log.SetFlags(log.Llongfile | log.LstdFlags)

	log.Printf("VERSION: %s\n", VERSION)

	env := "./config/env.ini"

	if len(os.Args) > 1 {
		env = os.Args[1]
	}

	var httpd = Httpd{}

	err := inifile.DecodeFile(&httpd, "./app.ini")

	if err != nil {
		log.Panicln(err)
	}

	err = inifile.DecodeFile(&httpd, env)

	if err != nil {
		log.Panicln(err)
	}

	var https = map[int64]chan kk.Message{}

	var reply, getname = kk.TCPClientConnect(httpd.Name, httpd.Address, httpd.Options, func(message *kk.Message) {

		var i = strings.LastIndex(message.To, ".")
		var id, _ = strconv.ParseInt(message.To[i+1:], 10, 64)
		var ch, ok = https[id]

		if ok && ch != nil {
			if message.Method == "REQUEST" {
				ch <- *message
				delete(https, id)
			} else {
				var m = kk.Message{"UNAVAILABLE", "", "", "", []byte("")}
				ch <- m
				delete(https, id)
			}
		}

	})

	var uuid int64 = time.Now().UnixNano()

	var http_handler = func(w http.ResponseWriter, r *http.Request) {

		var id int64 = uuid + 1
		uuid = id
		var ch = make(chan kk.Message, 2048)
		defer close(ch)

		var body = make([]byte, r.ContentLength)
		var contentType = r.Header.Get("Content-Type")
		var to = r.RequestURI[len(httpd.Alias):]
		var n, err = r.Body.Read(body)
		defer r.Body.Close()

		if err != nil && err != io.EOF {
			log.Println(err)
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte(err.Error()))
			return
		} else if int64(n) != r.ContentLength {
			log.Printf("%d %d\n", n, r.ContentLength)
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		var trackId = ""

		{
			var ip = r.Header.Get("X-CLIENT-IP")

			if ip == "" {
				ip = r.Header.Get("X-Real-IP")
			}

			if ip == "" {
				ip = r.RemoteAddr
			}

			var cookie, err = r.Cookie("kk")

			if err != nil {
				var v = http.Cookie{}
				v.Name = "kk"
				v.Value = strconv.FormatInt(time.Now().UnixNano(), 10)
				v.Expires = time.Now().Add(24 * 3600 * time.Second)
				v.HttpOnly = true
				v.MaxAge = 24 * 3600
				v.Path = "/"
				http.SetCookie(w, &v)
				cookie = &v
			}

			trackId = cookie.Value

			var b, _ = json.Encode(map[string]string{"code": trackId, "ip": ip,
				"User-Agent": r.Header.Get("User-Agent"),
				"Referer":    r.Header.Get("Referer"),
				"Path":       r.RequestURI,
				"Host":       r.Host,
				"Protocol":   r.Proto})

			var m = kk.Message{"MESSAGE", getname(), "kk.message.http.request", "text/json", b}

			kk.GetDispatchMain().Async(func() {
				reply(&m)
			})

		}

		kk.GetDispatchMain().Async(func() {

			https[id] = ch

			var m = kk.Message{"REQUEST", fmt.Sprintf("%s%s.%d", getname(), trackId, id), to, contentType, body}

			if !reply(&m) {
				var r = kk.Message{"TIMEOUT", "", "", "", []byte("")}
				ch <- r
				delete(https, id)
			}

		})

		kk.GetDispatchMain().AsyncDelay(func() {

			var ch = https[id]

			if ch != nil {
				var r = kk.Message{"TIMEOUT", "", "", "", []byte("")}
				ch <- r
				delete(https, id)
			}

		}, time.Second)

		var m, ok = <-ch

		if !ok {
			w.WriteHeader(http.StatusGatewayTimeout)
		} else {
			if m.Method == "TIMEOUT" {
				w.WriteHeader(http.StatusGatewayTimeout)
			} else if m.Method == "UNAVAILABLE" {
				w.WriteHeader(http.StatusServiceUnavailable)
			} else if m.Method == "REQUEST" {
				w.Header().Add("From", m.From)
				if strings.HasPrefix(m.Type, "text") {
					w.Header().Add("Content-Type", m.Type+"; charset=utf-8")
				} else {
					w.Header().Add("Content-Type", m.Type)
				}
				w.Header().Add("Content-Length", strconv.Itoa(len(m.Content)))
				w.WriteHeader(http.StatusOK)
				w.Write(m.Content)
			} else {
				w.WriteHeader(http.StatusUnsupportedMediaType)
			}
		}
	}

	go func() {

		http.HandleFunc(httpd.Alias, http_handler)

		log.Println("httpd " + httpd.HttpAddress)

		log.Fatal(http.ListenAndServe(httpd.HttpAddress, nil))

	}()

	kk.DispatchMain()

}
