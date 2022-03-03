package controllers

import (
	"fmt"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/vmware-tanzu/tanzu-framework/addons/pinniped/config-controller/constants"
	corev1 "k8s.io/api/core/v1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	clusterapiv1beta1 "sigs.k8s.io/cluster-api/api/v1beta1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

var _ = Describe("Controller", func() {
	// TODO: Test cases:
	//   Management cluster addon secret is created/updated/deleted (make sure wlc secrets get updated/deleted? accordingly)
	//   WLC addon secret is created (verify it gets updated w/info from CM and management cluster)
	//   CM is deleted (make sure nothing happens/secrets are not deleted)
	//   CM is updated (make sure secrets get updated)

	// making global var so we don't have to repeat var assignment, but also ok if we hate it
	cluster := &clusterapiv1beta1.Cluster{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: pinnipedNamespace,
			Name:      "some-name",
		},
	}

	Context("Cluster", func() {
		BeforeEach(func() {
			// have to make a copy so we don't edit global var
			clusterCopy := cluster.DeepCopy()
			err := k8sClient.Create(ctx, clusterCopy)
			Expect(err).NotTo(HaveOccurred())
			Eventually(func(g Gomega) {
				err := k8sClient.Get(ctx, client.ObjectKeyFromObject(clusterCopy), &clusterapiv1beta1.Cluster{})
				g.Expect(err).NotTo(HaveOccurred())
			}).Should(Succeed())
		})
		AfterEach(func() {
			if err := k8sClient.Delete(ctx, cluster); err != nil {
				Expect(k8serrors.IsNotFound(err)).To(BeTrue())
			}

			Eventually(func(g Gomega) {
				err := k8sClient.Get(ctx, client.ObjectKeyFromObject(cluster), &clusterapiv1beta1.Cluster{})
				g.Expect(k8serrors.IsNotFound(err)).To(BeTrue())
			}).Should(Succeed())
		})

		When("cluster is created", func() {
			It("creates a secret with identity_management_type set to none", func() {
				Eventually(func(g Gomega) {
					gotSecret := &corev1.Secret{
						ObjectMeta: metav1.ObjectMeta{
							Namespace: pinnipedNamespace,
							Name:      fmt.Sprintf("%s-pinniped-addon", cluster.Name),
						},
					}
					wantSecretLabels := map[string]string{
						constants.TKGAddonLabel:   constants.PinnipedAddonLabel,
						constants.TKGClusterNameLabel: cluster.Name,
					}

					wantSecretData := map[string][]byte{
						"values.yaml": []byte(`#@data/values
#@overlay/match-child-defaults missing_ok=True
---
identity_management_type: none
`),
					}
					err := k8sClient.Get(ctx, client.ObjectKeyFromObject(gotSecret), gotSecret)
					g.Expect(err).NotTo(HaveOccurred())
					g.Expect(gotSecret.Labels).To(Equal(wantSecretLabels))
					g.Expect(gotSecret.Data).To(Equal(wantSecretData))
				}).Should(Succeed())
			})
		})


		When("cluster is updated", func() {
			BeforeEach(func() {
				clusterCopy := cluster.DeepCopy()
				k8sClient.Get(ctx, client.ObjectKeyFromObject(clusterCopy), clusterCopy)
				annotations := clusterCopy.ObjectMeta.Annotations
				if annotations == nil {
					annotations = make(map[string]string)
				}
				annotations["sweetest-cat"] = "lionel"
				clusterCopy.ObjectMeta.Annotations = annotations
				Expect(k8sClient.Update(ctx, clusterCopy)).To(Succeed())
			})
			It("secret remains unchanged", func() {
				Eventually(func(g Gomega) {
					gotSecret := &corev1.Secret{
						ObjectMeta: metav1.ObjectMeta{
							Namespace: pinnipedNamespace,
							Name:      fmt.Sprintf("%s-pinniped-addon", cluster.Name),
						},
					}
					wantSecretLabels := map[string]string{
						constants.TKGAddonLabel:   constants.PinnipedAddonLabel,
						constants.TKGClusterNameLabel: cluster.Name,
					}

					wantSecretData := map[string][]byte{
						"values.yaml": []byte(`#@data/values
#@overlay/match-child-defaults missing_ok=True
---
identity_management_type: none
`),
					}
					err := k8sClient.Get(ctx, client.ObjectKeyFromObject(gotSecret), gotSecret)
					g.Expect(err).NotTo(HaveOccurred())
					g.Expect(gotSecret.Labels).To(Equal(wantSecretLabels))
					g.Expect(gotSecret.Data).To(Equal(wantSecretData))
				}).Should(Succeed())
			})
		})

		When("Cluster is deleted", func() {
			BeforeEach(func() {
				err := k8sClient.Delete(ctx, cluster)
				Expect(err).NotTo(HaveOccurred())
				Eventually(func(g Gomega) {
					err := k8sClient.Get(ctx, client.ObjectKeyFromObject(cluster), &clusterapiv1beta1.Cluster{})
					g.Expect(k8serrors.IsNotFound(err)).To(BeTrue())
				}).Should(Succeed())
			})

			It("deletes the Pinniped addon secret associated with the cluster", func() {
				Eventually(func(g Gomega) {
					gotSecret := &corev1.Secret{
						ObjectMeta: metav1.ObjectMeta{
							Namespace: pinnipedNamespace,
							Name:      fmt.Sprintf("%s-pinniped-addon", cluster.Name),
						},
					}
					err := k8sClient.Get(ctx, client.ObjectKeyFromObject(gotSecret), gotSecret)
					g.Expect(k8serrors.IsNotFound(err)).To(BeTrue())
				}).Should(Succeed())
			})
		})
	})

	Context("Addon Secret", func() {
		BeforeEach(func() {
			// have to make a copy so we don't edit global var
			clusterCopy := cluster.DeepCopy()
			Expect(k8sClient.Create(ctx, clusterCopy)).To(Succeed())
			Eventually(func(g Gomega) {
				err := k8sClient.Get(ctx, client.ObjectKeyFromObject(clusterCopy), &clusterapiv1beta1.Cluster{})
				g.Expect(err).NotTo(HaveOccurred())
			}).Should(Succeed())
		})

		AfterEach(func() {
			err := k8sClient.Delete(ctx, cluster)
			if err != nil {
				Expect(k8serrors.IsNotFound(err)).To(BeTrue())
			}

			Eventually(func(g Gomega) {
				err := k8sClient.Get(ctx, client.ObjectKeyFromObject(cluster), &clusterapiv1beta1.Cluster{})
				g.Expect(k8serrors.IsNotFound(err)).To(BeTrue())
			}).Should(Succeed())
		})
		When("the secret gets deleted", func() {
			gotSecret := &corev1.Secret{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: pinnipedNamespace,
					Name:      fmt.Sprintf("%s-pinniped-addon", cluster.Name),
				},
			}
			BeforeEach(func() {
				Eventually(func(g Gomega) {
					err := k8sClient.Delete(ctx, gotSecret)
					g.Expect(err).NotTo(HaveOccurred())
				}).Should(Succeed())
			})

			It("recreates the secret with identity_management_type set to none", func() {
				Eventually(func(g Gomega) {
					wantSecretLabels := map[string]string{
						constants.TKGAddonLabel:   constants.PinnipedAddonLabel,
						constants.TKGClusterNameLabel: cluster.Name,
					}

					wantSecretData := map[string][]byte{
						"values.yaml": []byte(`#@data/values
#@overlay/match-child-defaults missing_ok=True
---
identity_management_type: none
`),
					}
					err := k8sClient.Get(ctx, client.ObjectKeyFromObject(gotSecret), gotSecret)
					g.Expect(err).NotTo(HaveOccurred())
					g.Expect(gotSecret.Labels).To(Equal(wantSecretLabels))
					g.Expect(gotSecret.Data).To(Equal(wantSecretData))
				}).Should(Succeed())
			})
		})

		When("identity_management_type is changed on the secret", func() {
			gotSecret := &corev1.Secret{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: pinnipedNamespace,
					Name:      fmt.Sprintf("%s-pinniped-addon", cluster.Name),
				},
			}

			BeforeEach(func() {
				Eventually(func(g Gomega) {
					err := k8sClient.Get(ctx, client.ObjectKeyFromObject(gotSecret), gotSecret)
					g.Expect(err).NotTo(HaveOccurred())
				}).Should(Succeed())
				secretCopy := gotSecret.DeepCopy()
				updatedSecretDataValues := []byte(`identity_management_type: fire`)
				secretCopy.Data["values.yaml"] = updatedSecretDataValues
				Expect(k8sClient.Update(ctx, secretCopy)).To(Succeed())
			})

			It("updates the secret with the proper data values", func() {
				Eventually(func(g Gomega) {
					gotSecret := &corev1.Secret{
						ObjectMeta: metav1.ObjectMeta{
							Namespace: pinnipedNamespace,
							Name:      fmt.Sprintf("%s-pinniped-addon", cluster.Name),
						},
					}
					wantSecretLabels := map[string]string{
						constants.TKGAddonLabel:   constants.PinnipedAddonLabel,
						constants.TKGClusterNameLabel: cluster.Name,
					}

					wantSecretData := map[string][]byte{
						"values.yaml": []byte(`#@data/values
#@overlay/match-child-defaults missing_ok=True
---
identity_management_type: none
`),
					}
					err := k8sClient.Get(ctx, client.ObjectKeyFromObject(gotSecret), gotSecret)
					g.Expect(err).NotTo(HaveOccurred())
					g.Expect(gotSecret.Labels).To(Equal(wantSecretLabels))
					g.Expect(gotSecret.Data).To(Equal(wantSecretData))
				}).Should(Succeed())
			})
		})
		When("the secret does not have the Pinniped addon label", func() {
			gotSecret := &corev1.Secret{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: pinnipedNamespace,
					Name:      "another-secret",
					Labels: map[string]string{
						constants.TKGAddonLabel: "pumpkin",
						constants.TKGClusterNameLabel: cluster.Name,
					},
				},
			}

			BeforeEach(func() {
				secretCopy := gotSecret.DeepCopy()
				secretCopy.Type = "tkg.tanzu.vmware.com/addon"
				secretCopy.Data = map[string][]byte{}
				secretCopy.Data["values.yaml"] = []byte("identity_management_type: moses")
				Expect(k8sClient.Create(ctx, secretCopy)).To(Succeed())
			})

			AfterEach(func() {
				Expect(k8sClient.Delete(ctx, gotSecret)).To(Succeed())
			})

			It("does not get updated", func() {
				Eventually(func(g Gomega) {
					wantSecretLabels := map[string]string{
						constants.TKGAddonLabel:   "pumpkin",
						constants.TKGClusterNameLabel: cluster.Name,
					}

					wantSecretData := map[string][]byte{
						"values.yaml": []byte("identity_management_type: moses"),
					}
					err := k8sClient.Get(ctx, client.ObjectKeyFromObject(gotSecret), gotSecret)
					g.Expect(err).NotTo(HaveOccurred())
					g.Expect(gotSecret.Labels).To(Equal(wantSecretLabels))
					g.Expect(gotSecret.Data).To(Equal(wantSecretData))
				}).Should(Succeed())
			})
		})

		When("the secret is not an addon type", func() {
			gotSecret := &corev1.Secret{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: pinnipedNamespace,
					Name:      "newest-secret",
					Labels: map[string]string{
						constants.TKGAddonLabel: constants.PinnipedAddonLabel,
						constants.TKGClusterNameLabel: cluster.Name,
					},
				},
			}

			BeforeEach(func() {
				secretCopy := gotSecret.DeepCopy()
				secretCopy.Type = "not-an-addon"
				secretCopy.Data = map[string][]byte{}
				secretCopy.Data["values.yaml"] = []byte("identity_management_type: moses")
				Expect(k8sClient.Create(ctx, secretCopy)).To(Succeed())
			})

			AfterEach(func() {
				Expect(k8sClient.Delete(ctx, gotSecret)).To(Succeed())
			})

			It("does not get updated", func() {
				Eventually(func(g Gomega) {
					wantSecretLabels := map[string]string{
						constants.TKGAddonLabel:   constants.PinnipedAddonLabel,
						constants.TKGClusterNameLabel: cluster.Name,
					}

					wantSecretData := map[string][]byte{
						"values.yaml": []byte("identity_management_type: moses"),
					}
					err := k8sClient.Get(ctx, client.ObjectKeyFromObject(gotSecret), gotSecret)
					g.Expect(err).NotTo(HaveOccurred())
					g.Expect(gotSecret.Labels).To(Equal(wantSecretLabels))
					g.Expect(gotSecret.Data).To(Equal(wantSecretData))
				}).Should(Succeed())
			})
		})
	})

	Context("pinniped-info configmap", func() {
		cluster2 := &clusterapiv1beta1.Cluster{
			ObjectMeta: metav1.ObjectMeta{
				Namespace: pinnipedNamespace,
				Name:      "another-name",
			},
		}
		clusterList := &clusterapiv1beta1.ClusterList{Items: []clusterapiv1beta1.Cluster{*cluster, *cluster2}}
		BeforeEach(func() {
			for _, c := range clusterList.Items {
				// have to make a copy so we don't edit global var
				clusterCopy := c.DeepCopy()
				Expect(k8sClient.Create(ctx, clusterCopy)).To(Succeed())
				Eventually(func(g Gomega) {
					err := k8sClient.Get(ctx, client.ObjectKeyFromObject(clusterCopy), &clusterapiv1beta1.Cluster{})
					g.Expect(err).NotTo(HaveOccurred())
				}).Should(Succeed())
			}
		})

		AfterEach(func() {
			for _, c := range clusterList.Items {
				err := k8sClient.Delete(ctx, &c)
				if err != nil {
					Expect(k8serrors.IsNotFound(err)).To(BeTrue())
				}
				Eventually(func(g Gomega) {
					err := k8sClient.Get(ctx, client.ObjectKeyFromObject(&c), &clusterapiv1beta1.Cluster{})
					g.Expect(k8serrors.IsNotFound(err)).To(BeTrue())
				}).Should(Succeed())
			}
		})
		When("the configmap gets created", func() {
			configMap := &corev1.ConfigMap{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: "kube-public",
					Name:      "pinniped-info",
				},
			}
			BeforeEach(func() {
				Expect(k8sClient.Create(ctx, configMap)).To(Succeed())
			})

			It("loops through all the addons secrets", func() {
				// TODO: Check CM changes get propagated to secret
				Eventually(func(g Gomega) {
					wantSecretData := map[string][]byte{
						"values.yaml": []byte(`#@data/values
#@overlay/match-child-defaults missing_ok=True
---
identity_management_type: none
`),
					}

					for _, c := range clusterList.Items {
						secret := &corev1.Secret{
							ObjectMeta: metav1.ObjectMeta{
								Namespace: c.Namespace,
								Name:      fmt.Sprintf("%s-pinniped-addon", c.Name),
							},
						}

						wantSecretLabels := map[string]string{
							constants.TKGAddonLabel:   constants.PinnipedAddonLabel,
							constants.TKGClusterNameLabel: c.Name,
						}


						err := k8sClient.Get(ctx, client.ObjectKeyFromObject(secret), secret)
						g.Expect(err).NotTo(HaveOccurred())
						g.Expect(secret.Labels).To(Equal(wantSecretLabels))
						g.Expect(secret.Data).To(Equal(wantSecretData))
					}
				}).Should(Succeed())
			})
		})
		When("a configmap in a different namespace gets created", func() {
			// TODO: Add info to CM and make sure it doesn't get propagated to secrets
			configMap := &corev1.ConfigMap{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: pinnipedNamespace,
					Name:      "pinniped-info",
				},
			}

			BeforeEach(func() {
				Expect(k8sClient.Create(ctx, configMap)).To(Succeed())
			})

			It("does not update addon secrets", func() {
				Eventually(func(g Gomega) {
					wantSecretData := map[string][]byte{
						"values.yaml": []byte(`#@data/values
#@overlay/match-child-defaults missing_ok=True
---
identity_management_type: none
`),
					}

					for _, c := range clusterList.Items {
						secret := &corev1.Secret{
							ObjectMeta: metav1.ObjectMeta{
								Namespace: c.Namespace,
								Name:      fmt.Sprintf("%s-pinniped-addon", c.Name),
							},
						}

						wantSecretLabels := map[string]string{
							constants.TKGAddonLabel:   constants.PinnipedAddonLabel,
							constants.TKGClusterNameLabel: c.Name,
						}


						err := k8sClient.Get(ctx, client.ObjectKeyFromObject(secret), secret)
						g.Expect(err).NotTo(HaveOccurred())
						g.Expect(secret.Labels).To(Equal(wantSecretLabels))
						g.Expect(secret.Data).To(Equal(wantSecretData))
					}
				}).Should(Succeed())
			})
		})
		When("a configmap with a different name gets created", func() {
			// TODO: Add info to CM and make sure it doesn't get propagated to secrets
			configMap := &corev1.ConfigMap{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: pinnipedNamespace,
					Name:      "kitties",
				},
			}

			BeforeEach(func() {
				Expect(k8sClient.Create(ctx, configMap)).To(Succeed())
			})

			It("does not update addon secrets", func() {
				Eventually(func(g Gomega) {
					wantSecretData := map[string][]byte{
						"values.yaml": []byte(`#@data/values
#@overlay/match-child-defaults missing_ok=True
---
identity_management_type: none
`),
					}

					for _, c := range clusterList.Items {
						secret := &corev1.Secret{
							ObjectMeta: metav1.ObjectMeta{
								Namespace: c.Namespace,
								Name:      fmt.Sprintf("%s-pinniped-addon", c.Name),
							},
						}

						wantSecretLabels := map[string]string{
							constants.TKGAddonLabel:   constants.PinnipedAddonLabel,
							constants.TKGClusterNameLabel: c.Name,
						}


						err := k8sClient.Get(ctx, client.ObjectKeyFromObject(secret), secret)
						g.Expect(err).NotTo(HaveOccurred())
						g.Expect(secret.Labels).To(Equal(wantSecretLabels))
						g.Expect(secret.Data).To(Equal(wantSecretData))
					}
				}).Should(Succeed())
			})
		})
	})
})
