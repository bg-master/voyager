/*
Copyright The Voyager Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package operator

import (
	"github.com/appscode/go/log"

	"github.com/golang/glog"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/tools/cache"
	"kmodules.xyz/client-go/tools/queue"
)

func (op *Operator) initNamespaceWatcher() {
	op.nsInformer = op.kubeInformerFactory.Core().V1().Namespaces().Informer()
	op.nsQueue = queue.New("Namespace", op.MaxNumRequeues, op.NumThreads, op.reconcileNamespace)
	op.nsInformer.AddEventHandler(queue.NewDeleteHandler(op.nsQueue.GetQueue()))
	op.nsLister = op.kubeInformerFactory.Core().V1().Namespaces().Lister()
}

func (op *Operator) reconcileNamespace(key string) error {
	_, exists, err := op.nsInformer.GetIndexer().GetByKey(key)
	if err != nil {
		glog.Errorf("Fetching object with key %s from store failed with %v", key, err)
		return err
	}
	if !exists {
		glog.Warningf("Namespace %s does not exist anymore\n", key)
		if _, name, err := cache.SplitMetaNamespaceKey(key); err != nil {
			return err
		} else {
			op.deleteCRDs(name)
		}
	}
	return nil
}

func (op *Operator) deleteCRDs(ns string) {
	if resources, err := op.VoyagerClient.VoyagerV1beta1().Certificates(ns).List(metav1.ListOptions{}); err == nil {
		for _, resource := range resources.Items {
			err := op.VoyagerClient.VoyagerV1beta1().Certificates(resource.Namespace).Delete(resource.Name, &metav1.DeleteOptions{})
			log.Error(err)
		}
	}
	if resources, err := op.VoyagerClient.VoyagerV1beta1().Ingresses(ns).List(metav1.ListOptions{}); err == nil {
		for _, resource := range resources.Items {
			err := op.VoyagerClient.VoyagerV1beta1().Ingresses(resource.Namespace).Delete(resource.Name, &metav1.DeleteOptions{})
			log.Error(err)
		}
	}
}
