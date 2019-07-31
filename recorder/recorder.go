package recorder

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
)

type ResponseRecord struct {
	Headers http.Header
	Body    string
}

type RequestRecord struct {
	Headers  http.Header
	Body     string
	Response ResponseRecord
}

type Recorder struct {
	Records map[string]map[string]RequestRecord
}

var RRRecorder *Recorder

func init() {
	RRRecorder = &Recorder{
		Records: make(map[string]map[string]RequestRecord),
	}
}

type bodyClone struct {
	*bytes.Reader
}

func (r *bodyClone) Close() error {
	return nil
}

func (r *Recorder) StoreRequest(req *http.Request) error {
	var b []byte
	var err error

	if req.Body != nil {
		save := req.Body
		b, err = ioutil.ReadAll(req.Body)
		//req.Body.Close()

		if err != nil {
			return err
		}
		req.Body = save
	}

	key := fmt.Sprintf("%s/%s", req.URL.Path, req.URL.RawQuery)

	if _, ok := r.Records[key]; !ok {
		r.Records[key] = make(map[string]RequestRecord)
	}
	r.Records[key][req.Method] = RequestRecord{
		Headers: req.Header,
		Body:    string(b),
	}
	return nil
}

func (r *Recorder) StoreResponse(res *http.Response) error {
	var b []byte
	var err error

	if res.Body != nil {
		b, err = ioutil.ReadAll(res.Body)
		//res.Body.Close()

		if err != nil {
			return err
		}
		res.Body = &bodyClone{
			Reader: bytes.NewReader(b),
		}
	}

	key := fmt.Sprintf("%s/%s", res.Request.URL.Path, res.Request.URL.RawQuery)

	if _, ok := r.Records[key]; !ok {
		return errors.New("Cannot find original request " + res.Request.URL.String())
	}
	if _, ok := r.Records[key][res.Request.Method]; !ok {
		return errors.New("Cannot find method " + res.Request.Method + " for URL " + res.Request.URL.String())
	}

	record := r.Records[key][res.Request.Method]
	record.Response.Body = string(b)
	record.Response.Headers = res.Header

	r.Records[key][res.Request.Method] = record

	return nil
}

func (r *Recorder) Dump(filename string) error {

	if len(r.Records) == 0 {
		return nil
	}

	b, err := json.Marshal(r.Records)
	if err != nil {
		return nil
	}

	return ioutil.WriteFile(filename, b, 0600)
}

func (r *Recorder) Load(filename string) error {

	b, err := ioutil.ReadFile(filename)
	if err != nil {
		return err
	}

	var objMap map[string]*json.RawMessage

	if err := json.Unmarshal(b, &objMap); err != nil {
		return err
	}

	//fmt.Printf("%+v\n", objMap)

	for url, method := range objMap {

		var methodObject map[string]*json.RawMessage

		if err := json.Unmarshal(*method, &methodObject); err != nil {
			return err
		}

		//fmt.Printf("%+v\n", methodObject)

		for m, req := range methodObject {

			var reqRecord RequestRecord
			if err := json.Unmarshal(*req, &reqRecord); err != nil {
				return err
			}

			//fmt.Printf("%+v\n", reqRecord)

			if r.Records[url] == nil {
				r.Records[url] = make(map[string]RequestRecord)
			}

			r.Records[url][m] = reqRecord
		}

	}

	return nil
}

func (r *Recorder) GetRecordedResponse(url, method string) (RequestRecord, error) {
	if _, ok := r.Records[url]; !ok {
		return RequestRecord{}, errors.New("Cannot find request " + url)
	}
	if _, ok := r.Records[url][method]; !ok {
		return RequestRecord{}, errors.New("Cannot find method " + method + " for URL " + url)
	}

	rec := r.Records[url][method]

	return rec, nil
}
