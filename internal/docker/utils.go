package docker

import "github.com/docker/docker/api/types"

func convertPorts(ports []types.Port) []Port {
	res := make([]Port, len(ports))
	for i, p := range ports {
		res[i].IP = p.IP
		res[i].PrivatePort = p.PrivatePort
		res[i].PublicPort = p.PublicPort
	}
	return res
}
