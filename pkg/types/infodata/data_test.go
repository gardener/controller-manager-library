/*
 * Copyright 2020 SAP SE or an SAP affiliate company. All rights reserved. This file is licensed under the Apache Software License, v. 2 except as noted otherwise in the LICENSE file
 *
 *  Licensed under the Apache License, Version 2.0 (the "License");
 *  you may not use this file except in compliance with the License.
 *  You may obtain a copy of the License at
 *
 *       http://www.apache.org/licenses/LICENSE-2.0
 *
 *  Unless required by applicable law or agreed to in writing, software
 *  distributed under the License is distributed on an "AS IS" BASIS,
 *  WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 *  See the License for the specific language governing permissions and
 *  limitations under the License.
 *
 *
 */

package infodata_test

import (
	"encoding/json"
	"fmt"

	. "github.com/onsi/ginkgo"
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
			Expect(err).To(BeNil())
			list.Set("cert", c)
			Expect(len(list)).To(Equal(1))
			s, err := json.Marshal(list)
			Expect(err).To(BeNil())
			fmt.Printf("%s\n", s)
		})

		It("adds two cert info", func() {

			list := infodata.InfoDataList{}

			c1, err := certdata.NewCertificate(key1, cert1)
			Expect(err).To(BeNil())
			c2, err := certdata.NewCertificate(key2, cert2)
			Expect(err).To(BeNil())
			list.Set("cert1", c1)
			list.Set("cert2", c2)
			Expect(len(list)).To(Equal(2))

		})
		It("adds two cert info", func() {

			list := infodata.InfoDataList{}

			c1, err := certdata.NewCertificate(key1, cert1)
			Expect(err).To(BeNil())
			c2, err := certdata.NewCertificate(key2, cert2)
			Expect(err).To(BeNil())
			list.Set("cert1", c1)
			list.Set("cert2", c2)
			list.Set("cert1", c2)
			Expect(len(list)).To(Equal(2))

		})

		It("reads cert info", func() {

			list := infodata.InfoDataList{}

			c1, err := certdata.NewCertificate(key1, cert1)
			Expect(err).To(BeNil())
			c2, err := certdata.NewCertificate(key2, cert2)
			Expect(err).To(BeNil())
			err = list.Set("cert1", c1)
			Expect(err).To(BeNil())
			err = list.Set("cert2", c2)
			Expect(err).To(BeNil())
			Expect(len(list)).To(Equal(2))

			info, err := list.Get("cert2")
			Expect(err).To(BeNil())
			Expect(info).NotTo(BeNil())
			Expect(info.(certdata.Certificate).PrivateKey()).To(Equal(c2.PrivateKey()))
			Expect(info.(certdata.Certificate).Certificates()).To(Equal(c2.Certificates()))
		})

	})
	Context("strings", func() {
		It("adds and reads strings", func() {
			list := infodata.InfoDataList{}
			err := list.Set("test", simple.String("bla"))
			Expect(err).To(BeNil())
			err = list.Set("other", simple.String("blub"))
			Expect(err).To(BeNil())

			Expect(list.Get("test")).To(Equal(simple.String(("bla"))))
		})

		It("adds and reads string arrays", func() {
			list := infodata.InfoDataList{}
			err := list.Set("test", simple.StringArray([]string{"bla", "blub"}))
			Expect(err).To(BeNil())
			err = list.Set("other", simple.StringArray([]string{"alice", "bob"}))
			Expect(err).To(BeNil())

			Expect(list.Get("test")).To(Equal(simple.StringArray([]string{"bla", "blub"})))

			s, err := json.Marshal(list)
			Expect(err).To(BeNil())
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
			Expect(err).To(BeNil())

			Expect(list.Get("test")).To(Equal(v))

			s, err := json.Marshal(list)
			Expect(err).To(BeNil())
			fmt.Printf("%s\n", s)

		})
		It("adds and reads list values", func() {
			list := infodata.InfoDataList{}
			v := simple.ValueList{
				26.0,
				27.0,
			}
			err := list.Set("test", v)
			Expect(err).To(BeNil())

			Expect(list.Get("test")).To(Equal(v))
		})

	})
})
