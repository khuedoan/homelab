package cluster

import "fmt"

type nodeState string

const (
	fresh    nodeState = "fresh"
	prepared nodeState = "prepared"
	pending  nodeState = "pending"
	member   nodeState = "member"
)

type snapshot struct {
	hasToken, hasState, active    bool
	token, serverToken, clusterID string
}

func (state snapshot) classify(source credentials) (nodeState, error) {
	if state.hasToken && state.token != source.token || state.serverToken != "" && state.serverToken != source.token {
		return "", fmt.Errorf("conflicting credentials")
	}
	if state.hasState || state.serverToken != "" {
		if !state.hasToken || !state.active {
			return "", fmt.Errorf("existing datastore has no verifiable active membership")
		}
		if state.clusterID == "" {
			return pending, nil
		}
		if state.clusterID != source.clusterID {
			return "", fmt.Errorf("node belongs to a different cluster")
		}
		return member, nil
	}
	if state.hasToken {
		return prepared, nil
	}
	if state.active {
		return "", fmt.Errorf("k3s is running without managed enrollment credentials")
	}
	return fresh, nil
}
