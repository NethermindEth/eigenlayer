package docker

import (
	"context"
	"errors"
	"fmt"
	"io"
	"sort"
	"testing"
	"time"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/events"
	"github.com/docker/docker/client"
)

func EventsRange(ctx context.Context, since time.Time, until time.Time) (EventSet, error) {
	dockerClient, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		return nil, err
	}
	defer dockerClient.Close()

	eventsChan, errChan := dockerClient.Events(ctx, types.EventsOptions{
		Since: since.Format(time.RFC3339Nano),
		Until: until.Format(time.RFC3339Nano),
	})
	var events []events.Message
	for {
		select {
		case err := <-errChan:
			if errors.Is(err, io.EOF) {
				sort.SliceStable(events, func(i, j int) bool {
					return events[i].Time < events[j].Time
				})
				return events, nil
			}
			return nil, err
		case event := <-eventsChan:
			events = append(events, event)
		}
	}
}

type EventSet []events.Message

type EventPredicate interface {
	fmt.Stringer
	check(events.Message) bool
}

func (e EventSet) Check(t *testing.T, p EventPredicate) {
	t.Helper()
	for _, event := range e {
		if p.check(event) {
			t.Log("Event OK", p)
			return
		}
	}
	t.Errorf("failed to find event matching predicate: %s", p.String())
}

func (e EventSet) CheckInOrder(t *testing.T, ps ...EventPredicate) {
	t.Helper()
	var i int
	for pi, p := range ps {
		i = e.checkFromIndex(i, p)
		if i == -1 {
			if pi == 0 {
				t.Errorf(`failed to find event matching predicate: "(%d) %s"`, pi, p)
			} else {
				t.Errorf(`failed to find event matching predicate sequence: |(%d) %s| -> |(%d) %s|`, pi-1, ps[pi-1], pi, p)
			}
			return
		}
	}
}

func (e EventSet) checkFromIndex(i int, p EventPredicate) int {
	for ; i < len(e); i++ {
		if p.check(e[i]) {
			return i
		}
	}
	return -1
}
