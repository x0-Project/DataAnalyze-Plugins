package main

import (
	"HaE-AnalyzeEngine/Logger"
	"HaE-AnalyzeEngine/NatsClient"
	"HaE-AnalyzeEngine/PluginInterface"
	"HaE-AnalyzeEngine/Proto"
	"HaE-AnalyzeEngine/Utils"
	"crypto/tls"
	"crypto/x509"
	"github.com/elazarl/goproxy"
	"github.com/google/uuid"
	"google.golang.org/protobuf/proto"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/httputil"
	"os"
)

const maxChunkSize = 1000 * 1000 // 1 MB

const topic = "test"

type MyPlugin struct{}

func setCA(caCert, caKey []byte) error {
	goproxyCa, err := tls.X509KeyPair(caCert, caKey)
	if err != nil {
		return err
	}
	if goproxyCa.Leaf, err = x509.ParseCertificate(goproxyCa.Certificate[0]); err != nil {
		return err
	}
	goproxy.GoproxyCa = goproxyCa
	goproxy.OkConnect = &goproxy.ConnectAction{Action: goproxy.ConnectAccept, TLSConfig: goproxy.TLSConfigFromCA(&goproxyCa)}
	goproxy.MitmConnect = &goproxy.ConnectAction{Action: goproxy.ConnectMitm, TLSConfig: goproxy.TLSConfigFromCA(&goproxyCa)}
	goproxy.HTTPMitmConnect = &goproxy.ConnectAction{Action: goproxy.ConnectHTTPMitm, TLSConfig: goproxy.TLSConfigFromCA(&goproxyCa)}
	goproxy.RejectConnect = &goproxy.ConnectAction{Action: goproxy.ConnectReject, TLSConfig: goproxy.TLSConfigFromCA(&goproxyCa)}
	return nil
}

func initMitm() {

	go func() {
		port := ":8081"
		pwd, _ := os.Getwd()

		caCert, _ := ioutil.ReadFile(pwd + "/Tests/ca.pem")
		caKey, _ := ioutil.ReadFile(pwd + "/Tests/ca.key.pem")
		err := setCA(caCert, caKey)
		if err != nil {
			return
		}
		proxy := goproxy.NewProxyHttpServer()
		//关闭默认日志打印
		proxy.Verbose = false
		proxy.Logger = log.New(ioutil.Discard, "", 0)

		//设置handler
		proxy.OnRequest().HandleConnect(goproxy.AlwaysMitm)
		proxy.OnRequest().DoFunc(handleRequest)
		proxy.OnResponse().DoFunc(handleResponse)
		Logger.InfoLogger.Printf("Mitm Server Start Success , Address 0.0.0.0" + port)
		log.Fatal(http.ListenAndServe(port, proxy))
	}()

}

func handleRequest(req *http.Request, ctx *goproxy.ProxyCtx) (*http.Request, *http.Response) {

	ctx.UserData = uuid.New().String()
	rawRequest, _ := httputil.DumpRequest(req, true)

	if len(rawRequest) > maxChunkSize {
		chunkCount := (len(rawRequest) + maxChunkSize - 1) / maxChunkSize // 计算需要的块数
		for i := 0; i < chunkCount; i++ {
			start := i * maxChunkSize
			end := start + maxChunkSize
			if end > len(rawRequest) {
				end = len(rawRequest)
			}
			chunk := rawRequest[start:end]
			networkData := &Proto.NetworkData{
				RawData:     chunk,
				IsChunked:   true,
				ChunkNum:    int32(i + 1),
				TraceID:     ctx.UserData.(string),
				ServiceHost: req.Host,
				ServicePort: 9999,
				Topic:       topic,
				ReqType:     "REQ",
			}
			data, _ := proto.Marshal(networkData)
			sendData, _ := Utils.Compress(data)
			nc, _ := NatsClient.GetNatsConn("")
			err := nc.Publish("test", sendData)
			if err != nil {
				return nil, nil
			}
			Logger.InfoLogger.Println(ctx.UserData.(string) + ": REQ chunk send " + string((i + 1)))
		}

	} else {
		networkData := &Proto.NetworkData{
			RawData:     rawRequest,
			IsChunked:   false,
			ChunkNum:    -1,
			TraceID:     ctx.UserData.(string),
			ServiceHost: req.Host,
			ServicePort: 9999,
			Topic:       topic,
			ReqType:     "REQ",
		}
		data, _ := proto.Marshal(networkData)
		sendData, _ := Utils.Compress(data)
		nc, _ := NatsClient.GetNatsConn("")
		err := nc.Publish("test", sendData)
		if err != nil {
			return nil, nil
		}
	}

	return req, nil
}

func handleResponse(resp *http.Response, ctx *goproxy.ProxyCtx) *http.Response {

	rawResponse, _ := httputil.DumpResponse(resp, true)

	if len(rawResponse) > maxChunkSize {
		chunkCount := (len(rawResponse) + maxChunkSize - 1) / maxChunkSize // 计算需要的块数
		for i := 0; i < chunkCount; i++ {
			start := i * maxChunkSize
			end := start + maxChunkSize
			if end > len(rawResponse) {
				end = len(rawResponse)
			}
			chunk := rawResponse[start:end]
			networkData := &Proto.NetworkData{
				RawData:     chunk,
				IsChunked:   true,
				ChunkNum:    int32(i + 1),
				TraceID:     ctx.UserData.(string),
				ServiceHost: resp.Request.Host,
				ServicePort: 9999,
				Topic:       topic,
				ReqType:     "RES",
			}
			data, _ := proto.Marshal(networkData)
			sendData, _ := Utils.Compress(data)
			nc, _ := NatsClient.GetNatsConn("")
			err := nc.Publish("test", sendData)
			if err != nil {
				return nil
			}
			Logger.InfoLogger.Println(ctx.UserData.(string) + ": RES chunk send " + string((i + 1)))
		}

	} else {
		networkData := &Proto.NetworkData{
			RawData:     rawResponse,
			IsChunked:   false,
			ChunkNum:    -1,
			TraceID:     ctx.UserData.(string),
			ServiceHost: resp.Request.Host,
			ServicePort: 9999,
			Topic:       topic,
			ReqType:     "RES",
		}
		data, _ := proto.Marshal(networkData)
		sendData, _ := Utils.Compress(data)
		nc, _ := NatsClient.GetNatsConn("")
		err := nc.Publish("test", sendData)
		if err != nil {
			return nil
		}
	}

	return resp
}

func (p *MyPlugin) Init() {

	//载入时打印信息并初始化MITM
	Logger.InfoLogger.Printf("Loaded Plugin: %s, version: %s, Author: %s\n", p.Name(), p.Version(), p.Author())
	initMitm()
  
}

func (p *MyPlugin) ProcessBegin(data *Proto.NetworkData, params map[string]interface{}) error {

	return nil
}

func (p *MyPlugin) ProcessEnd(data *Proto.NetworkData, params map[string]interface{}) error {

	return nil
}

func (p *MyPlugin) Name() string {
	return "DataAnalyze-MITM"
}

func (p *MyPlugin) Version() string {
	return "0.1"
}

func (p *MyPlugin) Author() string {
	return "depy"
}

func (p *MyPlugin) Topic() string {
	return "test"
}

func RegisterPlugin() PluginInterface.Processor {
	return &MyPlugin{}
}
