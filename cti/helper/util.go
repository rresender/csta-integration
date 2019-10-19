package helper

import "net"

func ParseJSON(in map[string]interface{}, out map[string]string) {
	for k, v := range in {
		switch concreteVal := v.(type) {
		case map[string]interface{}:
			ParseJSON(v.(map[string]interface{}), out)
		case string:
			out[k] = concreteVal
		}
	}
}

func Lpad(s string, pad string, plength int) string {
	for i := len(s); i < plength; i++ {
		s = pad + s
	}
	return s
}

func GetQueueKey(invokeID string, appName string) string {
	return appName + "queue-" + invokeID
}

func GetMonitorCrossRefIDKey(monitorCrossRefID string, appName string) string {
	return appName + "MonitorCrossRefID" + monitorCrossRefID
}

func GetInvokeIDKey(invokeID string, appName string) string {
	return appName + "invokeID" + invokeID
}

func GetLocalIP() string {
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		return ""
	}
	for _, address := range addrs {
		if ipnet, ok := address.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
			if ipnet.IP.To4() != nil {
				return ipnet.IP.String()
			}
		}
	}
	return ""
}
