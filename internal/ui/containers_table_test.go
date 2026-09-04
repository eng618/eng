package ui

import (
	"strings"
	"testing"
)

func TestRenderContainerTable_Widths(t *testing.T) {
	mockContainers := []ContainerRow{
		{
			Name:    "extremely-long-container-name-for-testing-purpose",
			Service: "extremely-long-service-name",
			State:   "running",
			Health:  "healthy",
			Image:   "docker.io/library/very-long-image-name-with-deep-registry-path:v1.0.0-rc.1",
			Ports: []PortMapping{
				{TargetPort: 80, PublishedPort: 8080},
				{TargetPort: 443, PublishedPort: 8443},
				{TargetPort: 9000, PublishedPort: 9000},
			},
		},
	}

	widths := []int{60, 80, 120}

	for _, w := range widths {
		out := RenderContainerTable("test-stack", mockContainers, w)
		lines := strings.Split(out, "\n")
		if len(lines) == 0 {
			t.Errorf("expected rendered table lines for width %d", w)
		}

		// Ensure border lines don't get broken/scrambled into multiple lines unexpectedly
		for _, line := range lines {
			if strings.HasPrefix(line, "┌") && !strings.HasSuffix(line, "┐") {
				t.Errorf("scrambled top border line for width %d: %s", w, line)
			}
			if strings.HasPrefix(line, "└") && !strings.HasSuffix(line, "┘") {
				t.Errorf("scrambled bottom border line for width %d: %s", w, line)
			}
		}
	}
}
