package main

import (
	"context"
	"errors"
	"log"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"
)

// record separator ascii
var (
	rs = ","
)

func main() {
	log.Fatal(Main(context.Background()))
}

type LogFile struct {
	f   *os.File
	mtx sync.Mutex
}

func (f *LogFile) Write(ll *LogLine) (int, error) {
	f.mtx.Lock()
	defer f.mtx.Unlock()
	return f.f.WriteString(ll.String())
}

func newLogFile(filename string) (*LogFile, error) {
	logfile, err := os.OpenFile("./vellum.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0660)
	if err != nil {
		return nil, err
	}
	return &LogFile{
		f: logfile,
	}, nil
}

type LogLine struct {
	Id      string `json:"id"`
	AppName string `json:"appname"`
	Time    string `json:"time"`
}

func (ll *LogLine) String() string {
	return ll.Id + rs + ll.AppName + rs + ll.Time + "\n"
}

func extractLogLine(u *url.URL, then time.Time) (*LogLine, error) {
	pieces := strings.Split(u.Path, "/")
	if len(pieces) != 3 {
		log.Println(pieces)
		return nil, errors.New("url path does not contain enough information for id and appname")
	}

	return &LogLine{
		Id:      pieces[1],
		AppName: pieces[2],
		Time:    strconv.FormatInt(then.UnixNano(), 10),
	}, nil
}

func Main(ctx context.Context) error {
	logfile, err := newLogFile("./vellum.log")
	if err != nil {
		panic(err)
	}

	handleError := func(err error) {
		log.Println(err)
	}

	http.HandleFunc("/", func(res http.ResponseWriter, req *http.Request) {
		now := time.Now()
		// Always return a 200 as we don't want to slow the application
		// down just to log its time
		res.WriteHeader(200)
		logline, err := extractLogLine(req.URL, now)
		if err != nil {
			handleError(err)
			return
		}

		_, err = logfile.Write(logline)
		if err != nil {
			handleError(err)
			return
		}
	})
	return http.ListenAndServe(":8080", nil)
}
