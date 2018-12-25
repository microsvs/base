package base

import (
	"context"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"

	"github.com/microsvs/base/pkg/errors"
	"github.com/microsvs/base/pkg/rpc"
)

func FilterLimitRateHandler(w http.ResponseWriter, r *http.Request) {
	var (
		err  error
		data map[string]interface{}
		ctx  context.Context = r.Context()
	)
	//频率控制
	ctx = context.WithValue(ctx, rpc.KeyRawRequest, r)
	if data, err = rpc.CallService(ctx, Service2Url(rpc.FGSTraffic), "query{traffic}"); err != nil {
		//服务失败即通过
		fmt.Printf("[callservice] call limit rate service failed. err=%s\n", err.Error())
		return
	}
	if ok := data["traffic"].(bool); !ok {
		//错误
		GLReturnError(errors.FGETrafficControl, w)
		panic("_HALT_")
	}
	return
}

func FilterSignHandler(w http.ResponseWriter, r *http.Request) {
	//签名
	var (
		values url.Values = r.URL.Query()
		data   map[string]interface{}
		err    error
	)
	req := fmt.Sprintf(`query{sign(nonce:"%s",sign:"%s",token:"%s",appid:"%s")}`,
		values.Get("nonce"), values.Get("sign"), values.Get("token"), values.Get("appid"))
	if data, err = rpc.CallService(nil, Service2Url(rpc.FGSSign), req); err != nil {
		fmt.Printf("call sign service failed, err=%s\n", err.Error())
		//服务失败即通过
		return
	}
	if signOk := data["sign"].(bool); !signOk {
		//错误,打印body
		body, _ := ioutil.ReadAll(r.Body)
		fmt.Printf("sign failed，token=%s,url=%v,body=%v\n", values.Get("token"), r.URL.Query(), string(body))
		GLReturnError(errors.FGECheckSignFail, w)
		panic("_HALT_")
	}
	return
}
