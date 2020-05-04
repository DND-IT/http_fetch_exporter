package collector

import (
	"github.com/andelf/go-curl"
	"github.com/prometheus/common/log"
	"sort"
	"sync"
	"time"
)

type server struct {
	IP             string
	AVGFetchTime   float64
	MaxFetchTime   float64
	MinFetchTime   float64
	TotalFetchTime float64
	FetchCount     uint64
}

type Engine struct {
	Server *sync.Map
	URL    string
}

func New(URL string) *Engine {
	e := Engine{
		Server: new(sync.Map),
		URL:    URL,
	}

	return &e
}

func (e Engine) Start() {

	go func() {
		for {
			<-time.After(time.Second)
			e.doRequest()
		}
	}()

}

func (e Engine) doRequest() {

	var err error
	var primaryIP, totalTime interface{}

	easy := curl.EasyInit()
	defer easy.Cleanup()

	err = easy.Setopt(curl.OPT_TIMEOUT, 60)
	if err != nil {
		log.Error(err)
		return
	}

	err = easy.Setopt(curl.OPT_URL, e.URL)
	if err != nil {
		log.Error(err)
		return
	}

	// make a callback function
	voidFunc := func(_ []byte, _ interface{}) bool {
		return true
	}

	// get content and send to the void
	err = easy.Setopt(curl.OPT_WRITEFUNCTION, voidFunc)
	if err != nil {
		log.Error(err)
		return
	}

	if err := easy.Perform(); err != nil {
		log.Error(err)
		return
	}

	totalTime, err = easy.Getinfo(curl.INFO_TOTAL_TIME)
	if err != nil {
		log.Error(err)
		return
	}

	primaryIP, err = easy.Getinfo(curl.INFO_PRIMARY_IP)

	var primIP = primaryIP.(string)

	if err == nil && primIP != "" {
		if item, ok := e.Server.Load(primIP); ok {
			var s = item.(*server)
			s.FetchCount++
			var tt = totalTime.(float64)

			s.TotalFetchTime += tt
			if s.FetchCount > 0 {
				s.AVGFetchTime = s.TotalFetchTime / float64(s.FetchCount)
			}

			if s.MaxFetchTime < tt {
				s.MaxFetchTime = tt
			}

			if s.MinFetchTime == 0 {
				s.MinFetchTime = tt
			}

			if s.MinFetchTime > tt {
				s.MinFetchTime = tt
			}

			e.Server.Store(primIP, s)

		} else {
			var s = new(server)
			s.IP = primIP
			s.FetchCount++

			var tt = totalTime.(float64)

			s.TotalFetchTime += tt
			if s.FetchCount > 0 {
				s.AVGFetchTime = s.TotalFetchTime / float64(s.FetchCount)
			}

			if s.MaxFetchTime < tt {
				s.MaxFetchTime = tt
			}

			if s.MinFetchTime == 0 {
				s.MinFetchTime = tt
			}

			if s.MinFetchTime > tt {
				s.MinFetchTime = tt
			}

			e.Server.Store(primIP, s)

		}
	}

}

func (e *Engine) DumpServer() []server {
	var out []server
	var keys []string

	e.Server.Range(func(k, _ interface{}) bool {
		keys = append(keys, k.(string))
		return true
	})

	sort.StringsAreSorted(keys)

	for _, k := range keys {
		if v, ok := e.Server.Load(k); ok {

			var s = v.(*server)

			out = append(out, *s)

		}
	}

	return out
}
