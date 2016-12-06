package main

import (
	"github.com/kkserver/kk-lib/kk"
	"github.com/kkserver/kk-lib/kk/inifile"
	"log"
	"net/http"
	"os"
	"time"
)

const VERSION = "1.0.0"

type Httpd struct {
	Name        string
	Address     string
	HttpAddress string
	Alias       string
	Timeout     int64
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

	go func() {

		http.HandleFunc(httpd.Alias, kk.TCPClientHandleFunc(httpd.Name, httpd.Address, httpd.Options, httpd.Alias, time.Duration(httpd.Timeout)))

		log.Println("httpd " + httpd.HttpAddress)

		log.Fatal(http.ListenAndServe(httpd.HttpAddress, nil))

	}()

	kk.DispatchMain()

}
