package controllers

import (
	"context"
	"errors"
	"fmt"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"

	hellov1beta1 "github.com/bells17/k8s-sample-controller/api/v1beta1"
)

var _ = Describe("Message controller", func() {
	Context("when creating Message resource", func() {
		It("Should set .status.message", func() {
			ctx := context.Background()
			msg := hellov1beta1.Message{
				TypeMeta: metav1.TypeMeta{
					APIVersion: hellov1beta1.GroupVersion.String(),
					Kind:       "Message",
				},
				ObjectMeta: metav1.ObjectMeta{
					Name:      "sample",
					Namespace: "test",
				},
				Spec: hellov1beta1.MessageSpec{
					Message: "foo",
				},
			}

			err := k8sClient.Create(ctx, &msg)
			Expect(err).Should(Succeed())

			createdMsg := hellov1beta1.Message{}
			Eventually(func() error {
				err := k8sClient.Get(ctx, client.ObjectKey{Name: "sample", Namespace: "test"}, &createdMsg)
				if err != nil {
					return err
				}

				if createdMsg.Status.Message != fmt.Sprintf(messageTemplate, createdMsg.Spec.Message) {
					return errors.New("status is invalid")
				}

				return nil
			}).Should(Succeed())
		})
	})
})
