package replica

// Replica specifies a modification to a replica config.
// The number of replicas of a resource whose name matches will be set to count.
// This struct is used by the ReplicaCountTransform, and is meant to supplement
// the existing patch functionality with a simpler syntax for replica configuration.
type Replica struct {
	// The name of the resource to change the replica count
	Name string `json:"name,omitempty" yaml:"name,omitempty"`

	// The number of replicas required.
	Count uint `json:"count,omitempty" yaml:"count,omitempty"`
}
