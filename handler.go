package main

import (
	"fmt"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"

	"github.com/diegohce/httprecorder/recorder"
)

func createReplayProxies() *http.ServeMux {
	mux := http.NewServeMux()

	if err := recorder.RRRecorder.Load(httprConfig.Filename); err != nil {
		log.Error().Println(err)
	}

	//log.Debug().Printf("%+v\n", recorder.RRRecorder)

	mux.HandleFunc("/", func(rw http.ResponseWriter, r *http.Request) {

		key := fmt.Sprintf("%s/%s", r.URL.Path, r.URL.RawQuery)

		rec, err := recorder.RRRecorder.GetRecordedResponse(key, r.Method)
		if err != nil {
			rw.WriteHeader(500)
			fmt.Fprintf(rw, "%s", err.Error())
			return
		}

		log.Debug().Printf("%+v\n", rec.Response)

		for hName, hValue := range rec.Response.Headers {
			for _, v := range hValue {
				rw.Header().Add(hName, v)
			}
		}

		fmt.Fprintf(rw, "%s", rec.Response.Body)
	})

	return mux
}

func createRecordingProxies() *http.ServeMux {

	if err := recorder.RRRecorder.Load(httprConfig.Filename); err != nil {
		log.Error().Println(err)
	}
	mux := http.NewServeMux()

	for path, hostconfig := range httprConfig.Paths {
		log.Info().Println(path, "->", hostconfig.Host)

		url, err := url.Parse(hostconfig.Host)
		if err != nil {
			log.Error().Println(err, "parsing", hostconfig.Host)
			continue
		}

		p := newRecordingProxy(url)
		//p.ErrorHandler = errorHandler
		p.ModifyResponse = modifyResponseRecording
		mux.Handle(path, p)
	}

	if url, err := url.Parse(httprConfig.DefaultHost.Host); err != nil {
		log.Error().Println(err, "parsing", httprConfig.DefaultHost.Host)
	} else {
		p := newRecordingProxy(url)
		//p.ErrorHandler = errorHandler
		p.ModifyResponse = modifyResponseRecording
		mux.Handle("/", p)
	}

	return mux
}

func modifyResponseRecording(res *http.Response) error {

	err := recorder.RRRecorder.StoreResponse(res)
	if err != nil {
		log.Error().Println(err)
	}
	return err

	return nil
}

func singleJoiningSlash(a, b string) string {
	aslash := strings.HasSuffix(a, "/")
	bslash := strings.HasPrefix(b, "/")
	switch {
	case aslash && bslash:
		return a + b[1:]
	case !aslash && !bslash:
		return a + "/" + b
	}
	return a + b
}

func newRecordingProxy(target *url.URL) *httputil.ReverseProxy {
	targetQuery := target.RawQuery
	director := func(req *http.Request) {
		req.URL.Scheme = target.Scheme
		req.URL.Host = target.Host
		req.URL.Path = singleJoiningSlash(target.Path, req.URL.Path)
		if targetQuery == "" || req.URL.RawQuery == "" {
			req.URL.RawQuery = targetQuery + req.URL.RawQuery
		} else {
			req.URL.RawQuery = targetQuery + "&" + req.URL.RawQuery
		}
		if _, ok := req.Header["User-Agent"]; !ok {
			// explicitly disable User-Agent so it's not set to default value
			req.Header.Set("User-Agent", "")
		}
		if err := recorder.RRRecorder.StoreRequest(req); err != nil {
			log.Error().Println("Cannot record request", req.URL.String(), err)
		}
	}
	return &httputil.ReverseProxy{Director: director}
}
