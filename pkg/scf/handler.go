package scf

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/httputil"
	"strings"
)

var (
	ScfApiProxyUrl string
)

func HandlerHttp(w http.ResponseWriter, r *http.Request) {
	dumpReq, err := httputil.DumpRequest(r, true)
	if err != nil {
		log.Println("hander dump error ", err)
		w.WriteHeader(http.StatusServiceUnavailable)
		return
	}

	event := &DefineEvent{
		URL:     r.URL.String(),
		Content: base64.StdEncoding.EncodeToString(dumpReq),
	}
	bytejson, err := json.Marshal(event)
	if err != nil {
		log.Println("handler event json marshal error: ", err)
		w.WriteHeader(http.StatusServiceUnavailable)
		return
	}

	req, err := http.NewRequest("POST", ScfApiProxyUrl, bytes.NewReader(bytejson))
	if err != nil {
		log.Println("scf send server error ", err)
		w.WriteHeader(http.StatusServiceUnavailable)
		return
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		log.Println("client.Do()", err)
		w.WriteHeader(http.StatusServiceUnavailable)
		return
	}
	bytersp, _ := ioutil.ReadAll(resp.Body)

	var respevent RespEvent
	if err := json.Unmarshal(bytersp, &respevent); err != nil {
		log.Println("respevent unmarshal error: ", err)
		w.WriteHeader(http.StatusServiceUnavailable)
		return
	}
	if resp.StatusCode > 0 && resp.StatusCode != 200 {
		log.Printf(" status error:%d", resp.StatusCode)
		w.WriteHeader(http.StatusServiceUnavailable)
		fmt.Println("rsdata:", respevent.RspData)
		return
	}

	retByte, err := base64.StdEncoding.DecodeString(respevent.RspData)
	if err != nil {
		log.Println("base64 decode: ", err)
		w.WriteHeader(http.StatusServiceUnavailable)
		return
	}
	headers := respevent.RspHeader
	for hkey, hvalue := range headers {
		w.Header().Set(hkey, strings.Join(hvalue, ";"))
	}
	w.WriteHeader(respevent.RspStatus)

	resp.Body.Close()

	w.Write(retByte)
	return
}
