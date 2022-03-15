package types

import deploy "github.com/Ishan27g/ryo-Faas/proto"

type FunctionLogs struct {
	Function FunctionJsonRsp `json:"function,omitempty"`
	Logs     string          `json:"logs,omitempty"`
}
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
	AtAgent string `json:"atAgent,omitempty"`
}

type DeployedFunction struct {
	Name        string `types:"Name,omitempty"` // The name of the function
	Port        string
	Status      string `types:"Status"`
	GenFilePath string

	AgentAddr  string
	AgentFnUrl string `types:"AgentFnUrl"` // available at url

}

func registeredFnToJsonRsp(function DeployedFunction) FunctionJsonRsp {
	return FunctionJsonRsp{
		Name:    function.Name,
		Status:  function.Status,
		Url:     function.AgentFnUrl,
		AtAgent: function.AgentAddr,
	}
}

func JsonFunctionRspToRpc(jFn FunctionJsonRsp) *deploy.Function {
	return &deploy.Function{
		Entrypoint:       jFn.Name,
		Status:           jFn.Status,
		Url:              jFn.Url,
		ProxyServiceAddr: jFn.Proxy,
		AtAgent:          jFn.AtAgent,
	}
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
func RpcFunctionToJson(rFn *deploy.Function) FunctionJson {
	return FunctionJson{
		Name:     rFn.GetEntrypoint(),
		FilePath: rFn.GetFilePath(),
		Dir:      rFn.GetDir(),
	}
}
func RpcFunctionRspToJson(rFn *deploy.Function) FunctionJsonRsp {
	return FunctionJsonRsp{
		Name:    rFn.Entrypoint,
		Url:     rFn.GetUrl(),
		Status:  rFn.GetStatus(),
		Proxy:   rFn.ProxyServiceAddr,
		AtAgent: rFn.AtAgent,
	}
}
