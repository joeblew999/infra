package web

import (
	"bytes"
	_ "embed"
	"fmt"
	"html/template"
	"time"

	"github.com/joeblew999/infra/pkg/config"
	"github.com/joeblew999/infra/pkg/nats"
)

//go:embed templates/cluster_cards.html
var clusterCardsTemplate string

var (
	clusterCardsTpl = template.Must(template.New("nats-cluster-cards").Parse(clusterCardsTemplate))
)

type clusterTemplateData struct {
	LastUpdatedDisplay string
	LastUpdatedISO     string
	ClusterName        string
	Environment        string
	Summary            clusterSummary
	Nodes              []clusterNode
	ClientURLs         []string
}

type clusterSummary struct {
	Total          int
	Running        int
	Stopped        int
	HealthPercent  string
	HealthBarClass string
	HealthHeadline string
}

type clusterNode struct {
	Name        string
	Region      string
	ClientPort  string
	ClusterPort string
	HTTPPort    string
	Status      string
	Border      string
	StatusBadge string
}

// RenderClusterCards renders the cluster dashboard partial for DataStar SSE updates.
func RenderClusterCards(data clusterTemplateData) (string, error) {
	var buf bytes.Buffer
	if err := clusterCardsTpl.Execute(&buf, data); err != nil {
		return "", fmt.Errorf("execute cluster cards template: %w", err)
	}
	return buf.String(), nil
}

func buildClusterTemplateData(cluster nats.ClusterConfig, isLocal bool) clusterTemplateData {
	running := 0
	nodes := make([]clusterNode, len(cluster.Nodes))
	clientURLs := make([]string, 0, len(cluster.Nodes))

	for i, node := range cluster.Nodes {
		status := node.Status
		if status == "" {
			status = "unknown"
		}
		if status == "running" {
			running++
		}

		nodes[i] = clusterNode{
			Name:        node.Name,
			Region:      node.Region,
			ClientPort:  fmt.Sprintf("%d", node.Port),
			ClusterPort: fmt.Sprintf("%d", node.ClusterPort),
			HTTPPort:    fmt.Sprintf("%d", node.HTTPPort),
			Status:      status,
			Border:      nodeBorder(status),
			StatusBadge: nodeBadge(status),
		}

		clientURLs = append(clientURLs, fmt.Sprintf("nats://%s:%d", hostForNode(node, isLocal), node.Port))
	}

	total := len(cluster.Nodes)
	stopped := total - running
	healthPercent, barClass, headline := summarizeHealth(total, running)

	env := cluster.Environment
	if env == "" {
		if isLocal {
			env = config.EnvDevelopment
		} else {
			env = config.EnvProduction
		}
	}

	return clusterTemplateData{
		LastUpdatedDisplay: time.Now().Format("15:04:05"),
		LastUpdatedISO:     time.Now().Format(time.RFC3339),
		ClusterName:        cluster.ClusterName,
		Environment:        env,
		Summary: clusterSummary{
			Total:          total,
			Running:        running,
			Stopped:        stopped,
			HealthPercent:  healthPercent,
			HealthBarClass: barClass,
			HealthHeadline: headline,
		},
		Nodes:      nodes,
		ClientURLs: clientURLs,
	}
}

func summarizeHealth(total, running int) (percent string, barClass string, headline string) {
	if total == 0 {
		return "0", "bg-gray-400", "No nodes detected"
	}
	pct := int(float64(running) / float64(total) * 100)
	switch {
	case pct >= 90:
		barClass = "bg-emerald-500"
		headline = "Cluster healthy"
	case pct >= 60:
		barClass = "bg-amber-500"
		headline = "Partial availability"
	default:
		barClass = "bg-rose-500"
		headline = "Attention required"
	}
	return fmt.Sprintf("%d", pct), barClass, headline
}

func nodeBorder(status string) string {
	switch status {
	case "running":
		return "border border-emerald-200 dark:border-emerald-600"
	case "stopped":
		return "border border-rose-200 dark:border-rose-600"
	default:
		return "border border-amber-200 dark:border-amber-600"
	}
}

func nodeBadge(status string) string {
	switch status {
	case "running":
		return "bg-emerald-500/80 text-white"
	case "stopped":
		return "bg-rose-500/80 text-white"
	case "error", "failed":
		return "bg-rose-600/80 text-white"
	default:
		return "bg-amber-500/80 text-gray-900"
	}
}

func hostForNode(node nats.ClusterNode, isLocal bool) string {
	if isLocal {
		return "localhost"
	}
	if node.Region != "" {
		return fmt.Sprintf("%s.%s", node.Name, node.Region)
	}
	return node.Name
}
