package rpc

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"

	"github.com/microsvs/base/pkg/types"
	"github.com/microsvs/base/pkg/utils"
	"github.com/opentracing-contrib/go-stdlib/nethttp"
)

type Resp struct {
	Data    map[string]interface{} `json:"data"`
	ErrResp types.CustomError      `json:"error"`
}

func CallService(ctx context.Context, dns string, data string) (map[string]interface{}, error) {
	var (
		resp *http.Response
		err  error
		body []byte
	)
	url := fmt.Sprintf("http://%s/graphql", dns)
	resp, err = httpPostWithContext(ctx, url, "application/graphql", data)
	if err != nil {
		return nil, fmt.Errorf("[CallService] http request failed, err=%s", err.Error())
	}
	if body, err = ioutil.ReadAll(resp.Body); err != nil {
		return nil, fmt.Errorf("[CallService] read io.Reader failed. err=%s", err.Error())
	}
	resp.Body.Close()
	var retResp = new(Resp)
	if err = json.Unmarshal(body, retResp); err != nil {
		return nil, fmt.Errorf("[CallService] json decode failed. err=%s", err.Error())
	}
	if retResp.ErrResp.ErrCode > 0 {
		return nil, errors.New(retResp.ErrResp.ErrMsg)
	}
	return retResp.Data, nil
}

func httpPostWithContext(
	ctx context.Context, url string, contentType string, data string) (
	resp *http.Response, err error) {
	var (
		req *http.Request
	)
	if req, err = http.NewRequest("POST", url, strings.NewReader(data)); err != nil {
		return nil, err
	}
	req, ht := nethttp.TraceRequest(
		utils.GetGlobalTracer(),
		req,
		nethttp.OperationName("HTTP POST: "+getOperationName(data)),
	)
	defer ht.Finish()
	if err = ContextToHTTPRequest(ctx, req); err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", contentType)
	client := &http.Client{Transport: &nethttp.Transport{}}
	return client.Do(req)
}

func getOperationName(data string) string {
	if len(data) <= 0 {
		return "unkown operation name"
	}
	data = strings.Replace(data, "\t", "", -1)
	data = strings.Replace(data, "\n", "", -1)
	fields := strings.Split(data, "{")
	return fields[0] + " | " + fields[1]
}
