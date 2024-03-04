/*
 * SPDX-FileCopyrightText: 2020 SAP SE or an SAP affiliate company and Gardener contributors
 *
 * SPDX-License-Identifier: Apache-2.0
 *
 *
 */

package infodata_test

import (
	"encoding/json"
	"fmt"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"k8s.io/client-go/util/cert"

	"github.com/gardener/controller-manager-library/pkg/types/infodata"
	"github.com/gardener/controller-manager-library/pkg/types/infodata/certdata"
	"github.com/gardener/controller-manager-library/pkg/types/infodata/simple"
	"github.com/gardener/controller-manager-library/pkg/utils/pkiutil"
)

var _ = Describe("InfoData test", func() {
	Context("certs", func() {
		key1, _ := pkiutil.NewPrivateKey()
		cert1, _ := cert.NewSelfSignedCACert(cert.Config{CommonName: "test"}, key1)

		key2, _ := pkiutil.NewPrivateKey()
		cert2, _ := cert.NewSelfSignedCACert(cert.Config{CommonName: "other"}, key2)
		It("adds cert info", func() {

			list := infodata.InfoDataList{}

			c, err := certdata.NewCertificate(key1, cert1)
			Expect(err).ToNot(HaveOccurred())
			Expect(list.Set("cert", c)).ToNot(HaveOccurred())
			Expect(list).To(HaveLen(1))
			s, err := json.Marshal(list)
			Expect(err).ToNot(HaveOccurred())
			fmt.Printf("%s\n", s)
		})

		It("adds two cert info", func() {

			list := infodata.InfoDataList{}

			c1, err := certdata.NewCertificate(key1, cert1)
			Expect(err).ToNot(HaveOccurred())
			c2, err := certdata.NewCertificate(key2, cert2)
			Expect(err).ToNot(HaveOccurred())
			Expect(list.Set("cert1", c1)).ToNot(HaveOccurred())
			Expect(list.Set("cert2", c2)).ToNot(HaveOccurred())
			Expect(list).To(HaveLen(2))

		})
		It("adds two cert info", func() {

			list := infodata.InfoDataList{}

			c1, err := certdata.NewCertificate(key1, cert1)
			Expect(err).ToNot(HaveOccurred())
			c2, err := certdata.NewCertificate(key2, cert2)
			Expect(err).ToNot(HaveOccurred())
			Expect(list.Set("cert1", c1)).ToNot(HaveOccurred())
			Expect(list.Set("cert2", c2)).ToNot(HaveOccurred())
			Expect(list.Set("cert1", c2)).ToNot(HaveOccurred())
			Expect(list).To(HaveLen(2))

		})

		It("reads cert info", func() {

			list := infodata.InfoDataList{}

			c1, err := certdata.NewCertificate(key1, cert1)
			Expect(err).ToNot(HaveOccurred())
			c2, err := certdata.NewCertificate(key2, cert2)
			Expect(err).ToNot(HaveOccurred())
			err = list.Set("cert1", c1)
			Expect(err).ToNot(HaveOccurred())
			err = list.Set("cert2", c2)
			Expect(err).ToNot(HaveOccurred())
			Expect(list).To(HaveLen(2))

			info, err := list.Get("cert2")
			Expect(err).ToNot(HaveOccurred())
			Expect(info).NotTo(BeNil())
			Expect(info.(certdata.Certificate).PrivateKey()).To(Equal(c2.PrivateKey()))
			Expect(info.(certdata.Certificate).Certificates()).To(Equal(c2.Certificates()))
		})

	})
	Context("strings", func() {
		It("adds and reads strings", func() {
			list := infodata.InfoDataList{}
			err := list.Set("test", simple.String("bla"))
			Expect(err).ToNot(HaveOccurred())
			err = list.Set("other", simple.String("blub"))
			Expect(err).ToNot(HaveOccurred())

			Expect(list.Get("test")).To(Equal(simple.String(("bla"))))
		})

		It("adds and reads string arrays", func() {
			list := infodata.InfoDataList{}
			err := list.Set("test", simple.StringArray([]string{"bla", "blub"}))
			Expect(err).ToNot(HaveOccurred())
			err = list.Set("other", simple.StringArray([]string{"alice", "bob"}))
			Expect(err).ToNot(HaveOccurred())

			Expect(list.Get("test")).To(Equal(simple.StringArray([]string{"bla", "blub"})))

			s, err := json.Marshal(list)
			Expect(err).ToNot(HaveOccurred())
			fmt.Printf("%s\n", s)

		})
	})
	Context("values", func() {
		It("adds and reads values", func() {
			list := infodata.InfoDataList{}
			v := simple.Values{
				"alice": 25.0,
				"bob": []interface{}{
					26.0,
					27.0,
				},
			}
			err := list.Set("test", v)
			Expect(err).ToNot(HaveOccurred())

			Expect(list.Get("test")).To(Equal(v))

			s, err := json.Marshal(list)
			Expect(err).ToNot(HaveOccurred())
			fmt.Printf("%s\n", s)

		})
		It("adds and reads list values", func() {
			list := infodata.InfoDataList{}
			v := simple.ValueList{
				26.0,
				27.0,
			}
			err := list.Set("test", v)
			Expect(err).ToNot(HaveOccurred())

			Expect(list.Get("test")).To(Equal(v))
		})

	})
})
