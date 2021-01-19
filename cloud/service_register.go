package cloud

import (
	"errors"
	"github.com/sirupsen/logrus"
	"github.com/yyxing/glu/util"
	"io/ioutil"
	"math/rand"
	"net/http"
	"strconv"
	"strings"
	"time"
)

type NamingClient struct {
	servers     []ServerConfig
	client      http.Client
	namespaceId string
}

type ServerConfig struct {
	IpAddr string
	Port   uint64
	Scheme string
}

func NewNamingClient(servers []ServerConfig, timeout time.Duration, namespaceId string) *NamingClient {
	return &NamingClient{
		servers: servers,
		client: http.Client{
			Transport: http.DefaultTransport,
			Timeout:   timeout,
		},
		namespaceId: namespaceId,
	}
}
func (c *NamingClient) RegisterInstance(instance Instance) (bool, error) {
	// 向远程发送注册信息
	params := make(map[string]string)
	params["namespaceId"] = c.namespaceId
	params["serviceName"] = instance.ServiceName
	params["ip"] = instance.Ip
	params["port"] = strconv.Itoa(instance.Port)
	params["weight"] = strconv.FormatFloat(instance.Weight, 'f', -1, 64)
	params["healthy"] = strconv.FormatBool(instance.Healthy)
	params["ephemeral"] = strconv.FormatBool(instance.Ephemeral)
	_, err := c.reqApi(http.MethodPost, ServicePath, params)
	if err != nil {
		return false, err
	}
	if instance.Ephemeral {
		c.sendHeartBeat(BeatInfo{
			Port:        instance.Port,
			Ip:          instance.Ip,
			ServiceName: instance.ServiceName,
			Weight:      instance.Weight,
			Ephemeral:   instance.Ephemeral,
		})
	}
	return true, nil
}

func (c *NamingClient) sendHeartBeat(info BeatInfo) {
	api := ServiceBasePath + "/instance/beat"
	params := make(map[string]string)
	params["namespaceId"] = c.namespaceId
	params["serviceName"] = info.ServiceName
	params["beat"] = util.ToJsonString(info)
	heartBeat := time.NewTicker(5 * time.Second)
	go func() {
		for range heartBeat.C {
			_, err := c.reqApi(http.MethodPut, api, params)
			if err != nil {
				logrus.Info(err)
			}
		}
	}()
}
func (c *NamingClient) reqApi(method, path string, params map[string]string) (result string, err error) {
	if c.servers == nil || len(c.servers) <= 0 {
		return "", errors.New("server list is empty")
	}
	if len(c.servers) == 1 {
		for i := 0; i < RetryRequestTimes; i++ {
			resp, err := c.request(method, getAddress(c.servers[0])+path, params)
			result := getResponseInfo(resp)
			if err == nil {
				logrus.Info(result)
				return "", nil
			}
			logrus.Printf("api<%s>,method:<%s>, params:<%s>, call domain error:<%+v> , result:<%s>", path, method,
				util.ToJsonString(params), err, result)
		}
		return "", errors.New("retry " + strconv.Itoa(RetryRequestTimes) + " times request failed!")
	} else {
		index := rand.Intn(len(c.servers))
		for i := 1; i <= len(c.servers); i++ {
			resp, err := c.request(method, getAddress(c.servers[index])+path, params)
			result := getResponseInfo(resp)
			if err == nil {
				return "", nil
			}
			logrus.Printf("api<%s>,method:<%s>, params:<%s>, call domain error:<%+v> , result:<%s>", path, method,
				util.ToJsonString(params), err, result)
			index = (index + i) % len(c.servers)
		}
		return "", errors.New("retry " + strconv.Itoa(RetryRequestTimes) + " times request failed!")
	}
}
func getResponseInfo(response *http.Response) string {
	if response == nil {
		return ""
	}
	body := response.Body
	defer response.Body.Close()
	bytes, err := ioutil.ReadAll(body)
	if err != nil {
		return ""
	}
	return string(bytes)
}
func getAddress(server ServerConfig) string {
	if strings.HasPrefix(server.IpAddr, "http://") || strings.HasPrefix(server.IpAddr, "https://") {
		return server.IpAddr + ":" + strconv.Itoa(int(server.Port))
	}
	return server.Scheme + "://" + server.IpAddr + ":" + strconv.Itoa(int(server.Port))
}
func (c *NamingClient) request(method, path string, params map[string]string) (response *http.Response, err error) {
	header := make(map[string][]string)
	header["Connection"] = []string{"Keep-Alive"}
	header["Request-Module"] = []string{"Devil-Naming"}
	header["Content-Type"] = []string{"application/x-www-form-urlencoded;charset=utf-8"}
	switch method {
	case http.MethodGet:
		response, err = c.get(path, header, params)
		break
	case http.MethodPost:
		response, err = c.post(path, header, params)
		break
	case http.MethodPut:
		response, err = c.put(path, header, params)
		break
	case http.MethodDelete:
		response, err = c.delete(path, header, params)
		break
	}
	return
}

func (c *NamingClient) get(path string, header http.Header, params map[string]string) (response *http.Response, err error) {
	return
}

func (c *NamingClient) post(path string, header http.Header, params map[string]string) (response *http.Response, err error) {
	body := util.GetUrlFormedMap(params)
	request, reqErr := http.NewRequest(http.MethodPost, path, strings.NewReader(body))
	if reqErr != nil {
		err = reqErr
		return
	}
	request.Header = header
	resp, errDo := c.client.Do(request)
	if errDo != nil {
		err = errDo
	} else {
		response = resp
	}
	return
}

func (c *NamingClient) delete(path string, header http.Header, params map[string]string) (response *http.Response, err error) {
	return
}

func (c *NamingClient) put(path string, header http.Header, params map[string]string) (response *http.Response, err error) {
	var body string
	for key, value := range params {
		if len(value) > 0 {
			body += key + "=" + value + "&"
		}
	}
	if strings.HasSuffix(body, "&") {
		body = body[:len(body)-1]
	}
	request, errNew := http.NewRequest(http.MethodPut, path, strings.NewReader(body))
	if errNew != nil {
		err = errNew
		return
	}
	request.Header = header
	resp, errDo := c.client.Do(request)
	if errDo != nil {
		logrus.Println(errDo)
		err = errDo
	} else {
		response = resp
	}
	return
}
