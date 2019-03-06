package alicloud

import (
	"strconv"
	"strings"

	"github.com/golang/glog"

	"k8s.io/api/core/v1"
)

func GetServiceTag(service *v1.Service) (string, string) {
	serviceTagKey := service.Namespace + "_" + service.Name
	serviceTagValue := ""

	servicePort := []string{}
	for _, specPort := range service.Spec.Ports {
		servicePort = append(servicePort, strconv.Itoa(int(specPort.Port)))
	}
	serviceTagValue = strings.Join(servicePort, " ")
	glog.Infof("alicloud: service tag key is %s, value is %s\n", serviceTagKey, serviceTagValue)
	return serviceTagKey, serviceTagValue
}

func convertSlbTagToMap(slbTag string) (portMap map[string]bool) {
	portMap = make(map[string]bool)
	ports := strings.Split(slbTag, " ")
	for _, port := range ports {
		if len(port) == 0 {
			continue
		}
		portMap[port] = true
	}
	return portMap
}
