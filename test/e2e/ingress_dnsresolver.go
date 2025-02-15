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

package e2e

import (
	api "github.com/appscode/voyager/apis/voyager/v1beta1"
	"github.com/appscode/voyager/test/framework"
	"github.com/appscode/voyager/test/test-server/client"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	core "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

var _ = Describe("IngressWithDNSResolvers", func() {
	var (
		f   *framework.Invocation
		ing *api.Ingress

		svcResolveDNSWithNS,
		svcNotResolvesRedirect,
		svcResolveDNSWithoutNS *core.Service
	)

	BeforeEach(func() {
		f = root.Invoke()
		ing = f.Ingress.GetSkeleton()
		f.Ingress.SetSkeletonRule(ing)
	})

	BeforeEach(func() {
		var err error
		svcResolveDNSWithNS = &core.Service{
			ObjectMeta: metav1.ObjectMeta{
				Name:      f.Ingress.UniqueName(),
				Namespace: f.Ingress.Namespace(),
				Annotations: map[string]string{
					api.UseDNSResolver:         "true",
					api.DNSResolverNameservers: `["8.8.8.8:53", "8.8.4.4:53"]`,
				},
			},
			Spec: core.ServiceSpec{
				Type:         core.ServiceTypeExternalName,
				ExternalName: "google.com",
			},
		}

		_, err = f.KubeClient.CoreV1().Services(svcResolveDNSWithNS.Namespace).Create(svcResolveDNSWithNS)
		Expect(err).NotTo(HaveOccurred())

		svcNotResolvesRedirect = &core.Service{
			ObjectMeta: metav1.ObjectMeta{
				Name:      f.Ingress.UniqueName(),
				Namespace: f.Ingress.Namespace(),
			},
			Spec: core.ServiceSpec{
				Type:         core.ServiceTypeExternalName,
				ExternalName: "google.com",
			},
		}

		_, err = f.KubeClient.CoreV1().Services(svcNotResolvesRedirect.Namespace).Create(svcNotResolvesRedirect)
		Expect(err).NotTo(HaveOccurred())

		svcResolveDNSWithoutNS = &core.Service{
			ObjectMeta: metav1.ObjectMeta{
				Name:      f.Ingress.UniqueName(),
				Namespace: f.Ingress.Namespace(),
			},
			Spec: core.ServiceSpec{
				Type:         core.ServiceTypeExternalName,
				ExternalName: "google.com",
			},
		}

		_, err = f.KubeClient.CoreV1().Services(svcResolveDNSWithoutNS.Namespace).Create(svcResolveDNSWithoutNS)
		Expect(err).NotTo(HaveOccurred())
	})

	JustBeforeEach(func() {
		By("Creating ingress with name " + ing.GetName())
		err := f.Ingress.Create(ing)
		Expect(err).NotTo(HaveOccurred())

		f.Ingress.EventuallyStarted(ing).Should(BeTrue())

		By("Checking generated resource")
		Expect(f.Ingress.IsExistsEventually(ing)).Should(BeTrue())
	})

	AfterEach(func() {
		if options.Cleanup {
			Expect(f.Ingress.Delete(ing)).NotTo(HaveOccurred())
			Expect(f.KubeClient.CoreV1().Services(svcResolveDNSWithNS.Namespace).Delete(svcResolveDNSWithNS.Name, &metav1.DeleteOptions{})).NotTo(HaveOccurred())
			Expect(f.KubeClient.CoreV1().Services(svcResolveDNSWithNS.Namespace).Delete(svcResolveDNSWithNS.Name, &metav1.DeleteOptions{})).NotTo(HaveOccurred())
			Expect(f.KubeClient.CoreV1().Services(svcResolveDNSWithNS.Namespace).Delete(svcResolveDNSWithNS.Name, &metav1.DeleteOptions{})).NotTo(HaveOccurred())
		}
	})

	Describe("ExternalNameResolver", func() {
		BeforeEach(func() {
			ing.Spec = api.IngressSpec{
				Backend: &api.HTTPIngressBackend{
					IngressBackend: api.IngressBackend{
						ServiceName: svcNotResolvesRedirect.Name,
						ServicePort: intstr.FromString("80"),
					}},
				Rules: []api.IngressRule{
					{
						IngressRuleValue: api.IngressRuleValue{
							HTTP: &api.HTTPIngressRuleValue{
								Paths: []api.HTTPIngressPath{
									{
										Path: "/test-dns",
										Backend: api.HTTPIngressBackend{
											IngressBackend: api.IngressBackend{
												ServiceName: svcResolveDNSWithNS.Name,
												ServicePort: intstr.FromString("80"),
											}},
									},
									{
										Path: "/test-no-dns",
										Backend: api.HTTPIngressBackend{
											IngressBackend: api.IngressBackend{
												ServiceName: svcNotResolvesRedirect.Name,
												ServicePort: intstr.FromString("80"),
											}},
									},
									{
										Path: "/test-no-backend-redirect",
										Backend: api.HTTPIngressBackend{
											IngressBackend: api.IngressBackend{
												ServiceName: svcResolveDNSWithoutNS.Name,
												ServicePort: intstr.FromString("80"),
											}},
									},
									{
										Path: "/test-no-backend-rule-redirect",
										Backend: api.HTTPIngressBackend{
											IngressBackend: api.IngressBackend{
												ServiceName: svcNotResolvesRedirect.Name,
												ServicePort: intstr.FromString("80"),
												BackendRules: []string{
													"http-request redirect location https://google.com code 302",
												},
											},
										},
									},
								},
							},
						},
					},
					{
						IngressRuleValue: api.IngressRuleValue{
							HTTP: &api.HTTPIngressRuleValue{
								Paths: []api.HTTPIngressPath{
									{
										Path: "/redirect-rule",
										Backend: api.HTTPIngressBackend{
											IngressBackend: api.IngressBackend{
												BackendRules: []string{
													"http-request redirect location https://github.com/appscode/discuss/issues code 301",
												},
												ServiceName: svcNotResolvesRedirect.Name,
												ServicePort: intstr.FromString("80"),
											},
										},
									},
								},
							},
						},
					},
					{
						IngressRuleValue: api.IngressRuleValue{
							HTTP: &api.HTTPIngressRuleValue{
								Paths: []api.HTTPIngressPath{
									{
										Path: "/redirect",
										Backend: api.HTTPIngressBackend{
											IngressBackend: api.IngressBackend{
												ServiceName: svcNotResolvesRedirect.Name,
												ServicePort: intstr.FromString("80"),
											},
										},
									},
								},
							},
						},
					},
					{
						IngressRuleValue: api.IngressRuleValue{
							HTTP: &api.HTTPIngressRuleValue{
								Paths: []api.HTTPIngressPath{
									{
										Path: "/back-end",
										Backend: api.HTTPIngressBackend{
											IngressBackend: api.IngressBackend{
												ServiceName: f.Ingress.TestServerName(),
												ServicePort: intstr.FromString("8989"),
											},
										},
									},
								},
							},
						},
					},
				},
			}
		})

		It("Should test dns resolvers", func() {
			By("Getting HTTP endpoints")
			eps, err := f.Ingress.GetHTTPEndpoints(ing)
			Expect(err).NotTo(HaveOccurred())
			Expect(len(eps)).Should(BeNumerically(">=", 1))

			By("Calling /test-no-dns")
			err = f.Ingress.DoHTTPTestRedirect(framework.MaxRetry, ing, eps, "GET", "/test-no-dns", func(r *client.Response) bool {
				return Expect(r.Status).Should(Equal(301)) &&
					Expect(r.ResponseHeader.Get("Location")).Should(Equal("http://google.com:80"))
			})
			Expect(err).NotTo(HaveOccurred())

			By("Calling /test-no-backend-redirect")
			err = f.Ingress.DoHTTPTestRedirect(framework.MaxRetry, ing, eps, "GET", "/test-no-backend-redirect", func(r *client.Response) bool {
				return Expect(r.Status).Should(Equal(301)) &&
					Expect(r.ResponseHeader.Get("Location")).Should(Equal("http://google.com:80"))
			})
			Expect(err).NotTo(HaveOccurred())

			By("Calling /test-no-backend-rule-redirect")
			err = f.Ingress.DoHTTPTestRedirect(framework.MaxRetry, ing, eps, "GET", "/test-no-backend-rule-redirect", func(r *client.Response) bool {
				return Expect(r.Status).Should(Equal(302)) &&
					Expect(r.ResponseHeader.Get("Location")).Should(Equal("https://google.com"))
			})
			Expect(err).NotTo(HaveOccurred())

			By("Calling /test-dns")
			err = f.Ingress.DoHTTPStatus(framework.MaxRetry, ing, eps, "GET", "/test-dns", func(r *client.Response) bool {
				return Expect(r.Status).Should(Equal(404))
			})
			Expect(err).NotTo(HaveOccurred())

			By("Calling /default")
			err = f.Ingress.DoHTTPTestRedirect(framework.MaxRetry, ing, eps, "GET", "/default", func(r *client.Response) bool {
				return Expect(r.Status).Should(Equal(301)) &&
					Expect(r.ResponseHeader.Get("Location")).Should(Equal("http://google.com:80"))
			})
			Expect(err).NotTo(HaveOccurred())
		})

		It("Should test dns with backend rules", func() {
			By("Getting HTTP endpoints")
			eps, err := f.Ingress.GetHTTPEndpoints(ing)
			Expect(err).NotTo(HaveOccurred())
			Expect(len(eps)).Should(BeNumerically(">=", 1))

			By("Calling /redirect-rule")
			err = f.Ingress.DoHTTPTestRedirect(framework.MaxRetry, ing, eps, "GET", "/redirect-rule", func(r *client.Response) bool {
				return Expect(r.Status).Should(Equal(301))
			})
			Expect(err).NotTo(HaveOccurred())
		})
	})
})
