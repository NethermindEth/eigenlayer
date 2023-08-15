package docker

import (
	"fmt"

	"github.com/docker/docker/api/types/events"
)

type ImageTagged struct {
	repository string
	tag        string
}

func NewImageTagged(repository, tag string) *ImageTagged {
	return &ImageTagged{
		repository: repository,
		tag:        tag,
	}
}

func (p *ImageTagged) String() string {
	return fmt.Sprintf("image-tag: %s:%s", p.repository, p.tag)
}

func (p *ImageTagged) check(e events.Message) bool {
	return e.Type == events.ImageEventType &&
		e.Action == "tag"
}

type ContainerCreated struct {
	image       string
	containerID *string
}

func NewContainerCreated(image string, containerID *string) *ContainerCreated {
	return &ContainerCreated{
		image:       image,
		containerID: containerID,
	}
}

func (p *ContainerCreated) String() string {
	return fmt.Sprintf(`container-created: image="%s"`, p.image)
}

func (p *ContainerCreated) check(e events.Message) bool {
	image, ok := e.Actor.Attributes["image"]
	if !ok {
		return false
	}
	ok = e.Type == events.ContainerEventType &&
		e.Action == "create" &&
		image == p.image
	if ok {
		*p.containerID = e.Actor.ID
	}
	return ok
}

type ContainerDies struct {
	containerID *string
}

func NewContainerDies(containerID *string) *ContainerDies {
	return &ContainerDies{
		containerID: containerID,
	}
}

func (p *ContainerDies) String() string {
	return fmt.Sprintf(`container-die: containerID="%s"`, *p.containerID)
}

func (p *ContainerDies) check(e events.Message) bool {
	return e.Type == events.ContainerEventType &&
		e.Action == "die" &&
		e.Actor.ID == *p.containerID
}

type ContainerDestroy struct {
	containerID *string
}

func NewContainerDestroy(containerID *string) *ContainerDestroy {
	return &ContainerDestroy{
		containerID: containerID,
	}
}

func (p *ContainerDestroy) String() string {
	return fmt.Sprintf(`container-destroy: containerID="%s"`, *p.containerID)
}

func (p *ContainerDestroy) check(e events.Message) bool {
	return e.Type == events.ContainerEventType &&
		e.Action == "destroy" &&
		e.Actor.ID == *p.containerID
}

type NetworkConnect struct {
	containerID *string
	networkID   *string
}

func NewNetworkConnect(containerID, networkID *string) *NetworkConnect {
	return &NetworkConnect{
		containerID: containerID,
		networkID:   networkID,
	}
}

func (p *NetworkConnect) String() string {
	return fmt.Sprintf(`network-connect: containerID="%s" networkID="%s"`, *p.containerID, *p.networkID)
}

func (p *NetworkConnect) check(e events.Message) bool {
	containerID := e.Actor.Attributes["container"]
	return e.Type == events.NetworkEventType &&
		e.Action == "connect" &&
		containerID == *p.containerID
}

type NetworkDisconnect struct {
	containerID *string
	networkID   *string
}

func NewNetworkDisconnect(containerID, networkID *string) *NetworkDisconnect {
	return &NetworkDisconnect{
		containerID: containerID,
		networkID:   networkID,
	}
}

func (p *NetworkDisconnect) String() string {
	return fmt.Sprintf(`network-disconnect: containerID="%s" networkID="%s"`, *p.containerID, *p.networkID)
}

func (p *NetworkDisconnect) check(e events.Message) bool {
	containerID := e.Actor.Attributes["container"]
	return e.Type == events.NetworkEventType &&
		e.Action == "disconnect" &&
		containerID == *p.containerID
}
