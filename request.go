package nacos

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strings"
)

type ClientConfigParam struct {
	Namespace string
	ConfigGroup string
	ConfigDataId string
}

type ServerConfig struct {
	ContextPath string // Nacos的ContextPath
	IpAddr      string // Nacos的服务地址
	Port        int    // Nacos的服务端口
	Scheme      string // Nacos的服务地址前缀
}

type ConfigClientParam struct {
	ClientConfig *ClientConfigParam
	ServerConfigs []*ServerConfig
}

type ClientServiceParam struct {
	Ip string `json:"ip"` // 实例IP
	Port string `json:"port"` // 实例端口号
	NamespaceID string `json:"namespaceId"` // 命名空间ID
	Weight float32 `json:"weight"` // 配置服务权重
	Enabled bool `json:"enabled"` // 是否上线
	Healthy bool `json:"healthy"` // 是否健康
	Metadata string `json:"metadata"` // 扩展信息
	ClusterName string `json:"clusterName"` // 集群名
	ServiceName string `json:"serviceName"` // 服务名称
	GroupName string `json:"groupName"` // 组名称
	Ephemeral bool `json:"ephemeral"` // 是否为临时实例
}

type ServiceClientParam struct {
	ClientConfig ClientServiceParam
	ServerConfigs []ServerConfig
}

type (

	IConfigClient interface {
		GetConfig(param *ConfigClientParam) ([]byte, error)
	}


	IServiceClient interface {
		RegisterService(param *ServiceClientParam) (bool, error)
		GetService(param *ServiceClientParam) ([]byte, error)
	}

	ConfigClient struct {

	}

	ServiceClient struct {

	}
)

func request(url, data string, header http.Header, method string, cookies ...*http.Cookie) *http.Response {
	client := http.Client{}
	req, err := http.NewRequest(method, url, strings.NewReader(data))
	if err != nil {
		log.Printf("创建Request请求异常：%v\n", err)
		return nil
	}
	req.Header = header
	if len(cookies) > 0{
		for i := range cookies{
			req.AddCookie(cookies[i])
		}
	}
	response, err := client.Do(req)
	if err != nil{
		log.Printf("拉取用户信息异常：%v\n", err)
		return nil
	}
	return response
}

func (client *ConfigClient)GetConfig(param *ConfigClientParam) ([]byte, error) {
	for idx, serverConfigParam := range param.ServerConfigs{
		response := request(fmt.Sprintf("%s://%s:%d/%s%s?tenant=%s&dataId=%s&group=%s",
			serverConfigParam.Scheme, serverConfigParam.IpAddr, serverConfigParam.Port,
			serverConfigParam.ContextPath, "/v1/cs/configs",
			param.ClientConfig.Namespace, param.ClientConfig.ConfigDataId, param.ClientConfig.ConfigGroup),
			"", nil, http.MethodGet)
		if response.StatusCode != 200{
			continue
		}
		configData, err := ioutil.ReadAll(response.Body)
		if err != nil {
			if idx == len(param.ServerConfigs)-1{
				break
			}
			continue
		}
		return configData, nil
	}
	return nil, errors.New("获取配置文件异常！")
}

func (client *ServiceClient)RegisterService(param *ServiceClientParam) (bool, error) {
	var header = http.Header{}
	header.Add("Content-Type", "application/json")
	data, err := json.Marshal(param.ClientConfig)
	if err != nil {
		return false, err
	}
	for idx, serverConfigParam := range param.ServerConfigs{
		response := request(fmt.Sprintf("%s://%s:%d/%s%s",
			serverConfigParam.Scheme, serverConfigParam.IpAddr, serverConfigParam.Port,
			serverConfigParam.ContextPath, "/v1/ns/instance"),
			string(data), header, http.MethodPost)
		if response.StatusCode != 200{
			if idx == len(param.ServerConfigs)-1{
				break
			}
			continue
		}
		return true, nil
	}
	return false, errors.New("创建服务异常！")
}

func (client *ServiceClient)GetService(param *ServiceClientParam) (map[string]interface{}, error){
	for idx, serverConfigParam := range param.ServerConfigs{
		response := request(fmt.Sprintf("%s://%s:%d/%s%s?serviceName=%s&groupName=%s&namespaceId=%s&clusters=%s&healthyOnly=%v",
			serverConfigParam.Scheme, serverConfigParam.IpAddr, serverConfigParam.Port,
			serverConfigParam.ContextPath, "/v1/ns/instance/list",
			param.ClientConfig.ServiceName, param.ClientConfig.GroupName, param.ClientConfig.NamespaceID,
			param.ClientConfig.ClusterName, param.ClientConfig.Healthy),
			"", nil, http.MethodGet)
		if response.StatusCode != 200{
			continue
		}
		serviceData, err := ioutil.ReadAll(response.Body)
		if err != nil {
			if idx == len(param.ServerConfigs)-1{
				break
			}
			continue
		}
		var serviceMap = make(map[string]interface{})
		err = json.Unmarshal(serviceData, &serviceMap)
		if err != nil {
			if idx == len(param.ServerConfigs)-1{
				break
			}
			continue
		}
		return serviceMap, nil
	}
	return nil, errors.New("获取配置文件异常！")
}

