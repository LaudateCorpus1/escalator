package k8s

import (
	"time"

	"k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/kubernetes"
	v1lister "k8s.io/client-go/listers/core/v1"
	"k8s.io/client-go/tools/cache"
)

// PodLister provides an interface for anything that can list a pod
type PodLister interface {
	List() ([]*v1.Pod, error)
}

// AllPodsLister lists all pods regardless of state
type AllPodsLister struct {
	podLister v1lister.PodLister
}

// NewAllPodsLister creates a new lister and informerSynced for all pods
func NewAllPodsLister(client kubernetes.Interface, namespace string) (PodLister, cache.InformerSynced) {
	selector := fields.Everything()
	podsListWatch := cache.NewListWatchFromClient(
		client.CoreV1().RESTClient(),
		"pods",
		namespace,
		selector,
	)
	podIndexer, podController := cache.NewIndexerInformer(
		podsListWatch,
		&v1.Pod{},
		1*time.Hour,
		cache.ResourceEventHandlerFuncs{},
		cache.Indexers{
			cache.NamespaceIndex: cache.MetaNamespaceIndexFunc,
		},
	)
	podLister := v1lister.NewPodLister(podIndexer)
	go podController.Run(wait.NeverStop)
	return &AllPodsLister{
		podLister,
	}, podController.HasSynced
}

// List lists all pods from the cache
func (lister *AllPodsLister) List() ([]*v1.Pod, error) {
	return lister.podLister.List(labels.Everything())
}

// NodeLister provides an interface for anything that can list a pod
type NodeLister interface {
	List() ([]*v1.Node, error)
}

// AllNodesLister lists all nodes regardless of state
type AllNodesLister struct {
	nodeLister v1lister.NodeLister
}

// NewAllNodesLister creates a new lister and informerSynced for all nodes
func NewAllNodesLister(client kubernetes.Interface, namespace string) (NodeLister, cache.InformerSynced) {
	selector := fields.Everything()
	podsListWatch := cache.NewListWatchFromClient(
		client.CoreV1().RESTClient(),
		"nodes",
		namespace,
		selector,
	)
	nodeIndexer, podController := cache.NewIndexerInformer(
		podsListWatch,
		&v1.Node{},
		1*time.Hour,
		cache.ResourceEventHandlerFuncs{},
		cache.Indexers{
			cache.NamespaceIndex: cache.MetaNamespaceIndexFunc,
		},
	)
	nodeLister := v1lister.NewNodeLister(nodeIndexer)
	go podController.Run(wait.NeverStop)
	return &AllNodesLister{
		nodeLister,
	}, podController.HasSynced
}

// List lists all pods from the cache
func (lister *AllNodesLister) List() ([]*v1.Node, error) {
	return lister.nodeLister.List(labels.Everything())
}
