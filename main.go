package main

import (
	"encoding/json"
	"fmt"
	"github.com/kkserver/kk-lib/kk"
	"io"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"
)

func help() {
	fmt.Println("kk-httpd <name> <127.0.0.1:8700> <:8900> /kk/")
}

func main() {

	log.SetFlags(log.Llongfile | log.LstdFlags)

	var args = os.Args
	var name string = ""
	var address string = ""
	var httpaddress string = ""
	var alias string = ""

	if len(args) > 4 {
		name = args[1]
		address = args[2]
		httpaddress = args[3]
		alias = args[4]
	} else {
		help()
		return
	}

	var https = map[int64]chan kk.Message{}

	var reply, getname = kk.TCPClientConnect(name, address, map[string]interface{}{"exclusive": true}, func(message *kk.Message) {

		log.Println(message.String())

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

	var http_handler = func(w http.ResponseWriter, r *http.Request) {

		var id = kk.UUID()
		var ch = make(chan kk.Message)
		defer close(ch)

		var body = make([]byte, r.ContentLength)
		var contentType = r.Header.Get("Content-Type")
		var to = r.RequestURI[len(alias):]
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
				v.Value = strconv.FormatInt(kk.UUID(), 10)
				v.Expires = time.Now().Add(24 * 3600 * time.Second)
				v.HttpOnly = true
				v.MaxAge = 24 * 3600
				v.Path = "/"
				http.SetCookie(w, &v)
				cookie = &v
			}

			trackId = cookie.Value

			var b, _ = json.Marshal(map[string]string{"code": trackId, "ip": ip,
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

		http.HandleFunc(alias, http_handler)

		log.Println("httpd " + httpaddress)

		log.Fatal(http.ListenAndServe(httpaddress, nil))

	}()

	kk.DispatchMain()

}
