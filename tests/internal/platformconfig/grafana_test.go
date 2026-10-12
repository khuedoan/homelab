package platformconfig

import (
	"os"
	"strings"
	"testing"
)

func TestGrafanaMatchesKanidmGroupSPN(t *testing.T) {
	grafana, err := os.ReadFile("../../../infra/_modules/platform/grafana.tf")
	if err != nil {
		t.Fatal(err)
	}
	kanidm, err := os.ReadFile("../../../infra/_modules/platform/kanidm.tf")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(kanidm), `domain = "auth.${var.domain}"`) {
		t.Fatal("update the Grafana group SPN if the Kanidm domain changes")
	}
	want := `contains(groups[*], 'grafana_admins@auth.${var.domain}') && 'GrafanaAdmin' || 'Viewer'`
	if !strings.Contains(string(grafana), want) {
		t.Fatal("Grafana must match the full Kanidm group SPN and default other groups to Viewer")
	}
	if !strings.Contains(string(grafana), `"openid profile email groups"`) {
		t.Fatal("Grafana must request the groups scope containing SPNs")
	}
}
