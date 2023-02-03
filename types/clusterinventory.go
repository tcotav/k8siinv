package types

type PodImageState struct {
	Name      string   `json:"name"`
	Namespace string   `json:"namespace"`
	StartTime string   `json:"starttime"`
	Images    []string `json:"images"`
}

type ClusterInventory struct {
	Version     string          `json:"version"`
	ClusterName string          `json:"clustername"`
	GeneratedAt string          `json:"generatedat"`
	ImageState  []PodImageState `json:"imagestate"`
}
