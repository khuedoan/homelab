package cluster

import "testing"

func TestNodeClassification(t *testing.T) {
	source := credentials{"secret", "cluster"}
	for _, tc := range []struct {
		name     string
		observed snapshot
		want     nodeState
	}{
		{"fresh", snapshot{}, fresh},
		{"published before start", snapshot{hasToken: true, token: "secret"}, prepared},
		{"joining", snapshot{hasToken: true, token: "secret", hasState: true, active: true}, pending},
		{"joined", snapshot{hasToken: true, token: "secret", hasState: true, active: true, clusterID: "cluster"}, member},
		{"other credential", snapshot{hasToken: true, token: "other"}, ""},
		{"empty credential", snapshot{hasToken: true}, ""},
		{"foreign datastore", snapshot{hasState: true}, ""},
		{"foreign server token", snapshot{hasToken: true, token: "secret", serverToken: "other"}, ""},
		{"stopped datastore", snapshot{hasToken: true, token: "secret", hasState: true}, ""},
		{"wrong cluster", snapshot{hasToken: true, token: "secret", hasState: true, active: true, clusterID: "other"}, ""},
		{"unmanaged process", snapshot{active: true}, ""},
	} {
		t.Run(tc.name, func(t *testing.T) {
			state, err := tc.observed.classify(source)
			if state != tc.want || (err != nil) != (tc.want == "") {
				t.Fatalf("got %s, %v; want %s", state, err, tc.want)
			}
		})
	}
}
