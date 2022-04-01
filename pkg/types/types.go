package types

import (
	deploy "github.com/Ishan27g/ryo-Faas/pkg/proto"
)

type FunctionJson struct {
	Name     string `json:"name,omitempty"`
	FilePath string `json:"filePath,omitempty"`
	Dir      string `json:"packageDir,omitempty"`
}
type FunctionJsonRsp struct {
	Name   string `json:"name,omitempty"`
	Status string `json:"status,omitempty"`
	Url    string `json:"url,omitempty"`

	Proxy   string `json:"proxy,omitempty"`
	IsAsync bool   `json:"IsAsync,omitempty"`
	IsMain  bool   `json:"IsMain,omitempty"`
}

func JsonFunctionToRpc(jFn FunctionJson) []*deploy.Function {
	var d []*deploy.Function
	d = append(d, &deploy.Function{
		Entrypoint: jFn.Name,
		FilePath:   jFn.FilePath,
		Dir:        jFn.Dir,
	})
	return d
}

func RpcFunctionRspToJson(rFn *deploy.Function) FunctionJsonRsp {
	return FunctionJsonRsp{
		Name:    rFn.GetEntrypoint(),
		Url:     rFn.GetUrl(),
		Status:  rFn.GetStatus(),
		Proxy:   rFn.GetProxyServiceAddr(),
		IsAsync: rFn.GetAsync(),
		IsMain:  rFn.GetIsMain(),
	}
}
