package recorder

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
)

type responseRecord struct {
	Headers http.Header
	Body    string
}

type requestRecord struct {
	Headers  http.Header
	Body     string
	Response responseRecord
}

type recorder struct {
	Records map[string]map[string]requestRecord
}

// RRRecorder The requests and responses recorder.
var RRRecorder *recorder

func init() {
	RRRecorder = &recorder{
		Records: make(map[string]map[string]requestRecord),
	}
}

type bodyClone struct {
	*bytes.Reader
}

func (r *bodyClone) Close() error {
	return nil
}

func (r *recorder) StoreRequest(req *http.Request) error {
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
		r.Records[key] = make(map[string]requestRecord)
	}
	r.Records[key][req.Method] = requestRecord{
		Headers: req.Header,
		Body:    string(b),
	}
	return nil
}

func (r *recorder) StoreResponse(res *http.Response) error {
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

func (r *recorder) Dump(filename string) error {

	if len(r.Records) == 0 {
		return nil
	}

	b, err := json.Marshal(r.Records)
	if err != nil {
		return nil
	}

	return ioutil.WriteFile(filename, b, 0600)
}

func (r *recorder) Load(filename string) error {

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

			var reqRecord requestRecord
			if err := json.Unmarshal(*req, &reqRecord); err != nil {
				return err
			}

			//fmt.Printf("%+v\n", reqRecord)

			if r.Records[url] == nil {
				r.Records[url] = make(map[string]requestRecord)
			}

			r.Records[url][m] = reqRecord
		}

	}

	return nil
}

func (r *recorder) GetRecordedResponse(url, method string) (requestRecord, error) {
	if _, ok := r.Records[url]; !ok {
		return requestRecord{}, errors.New("Cannot find request " + url)
	}
	if _, ok := r.Records[url][method]; !ok {
		return requestRecord{}, errors.New("Cannot find method " + method + " for URL " + url)
	}

	rec := r.Records[url][method]

	return rec, nil
}
