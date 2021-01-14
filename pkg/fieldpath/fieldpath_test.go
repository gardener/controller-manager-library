/*
 * SPDX-FileCopyrightText: 2020 SAP SE or an SAP affiliate company and Gardener contributors
 *
 * SPDX-License-Identifier: Apache-2.0
 *
 *
 */

package fieldpath

import (
	"fmt"
	"reflect"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

type A struct {
	FieldAB B
	FieldAC C
}
type B struct {
	FieldBaC  []C
	FieldBaaC [][]C
	FieldBaI  []interface{}
	FieldBapA []*A
	FieldBS   string
}

type C struct {
	FieldC1   string
	FieldC2   interface{}
	FieldCpB  *B
	FieldCppB **B
}

var _ = Describe("Fieldpath", func() {

	Context("for dynamic objects", func() {
		Context("and valid expressions", func() {
			It("simple field", func() {
				v := MAP{
					"A": "A",
				}
				p := MustFieldPath(".A")
				Expect(p.Get(v)).To(Equal("A"))
			})

			It("nested field", func() {
				v := MAP{
					"FieldAC": MAP{
						"FieldC1": "test",
					},
				}
				p := MustFieldPath(".FieldAC.FieldC1")
				Expect(p.Get(v)).To(Equal("test"))
			})

			It("projection", func() {
				v := MAP{
					"FieldAB": MAP{
						"FieldBaC": ARRAY{
							MAP{
								"FieldC1": "test",
							},
						},
					},
				}
				p := MustFieldPath(".FieldAB.FieldBaC[].FieldC1")
				Expect(p.Get(v)).To(Equal([]interface{}{"test"}))
			})

			It("empty projection", func() {
				v := MAP{
					"FieldAB": MAP{
						"FieldBaC": ARRAY{},
					},
				}
				p := MustFieldPath(".FieldAB.FieldBaC[].FieldC1")
				Expect(p.Get(v)).To(BeNil())
			})

			It("double projection", func() {
				v := MAP{
					"FieldAB": MAP{
						"FieldBaaC": []ARRAY{
							ARRAY{
								MAP{"FieldC1": "alice"},
							},
							ARRAY{
								MAP{"FieldC1": "bob"},
							},
						},
					},
				}
				p := MustFieldPath(".FieldAB.FieldBaaC[][].FieldC1")
				Expect(p.Get(v)).To(Equal([]ARRAY{ARRAY{"alice"}, ARRAY{"bob"}}))
			})

			It("nested double projection", func() {
				v := MAP{
					"FieldAB": MAP{
						"FieldBapA": ARRAY{
							MAP{
								"FieldAB": MAP{
									"FieldBaC": ARRAY{
										MAP{
											"FieldC1": "alice",
										},
										MAP{
											"FieldC1": "bob",
										},
									},
								},
							},
							MAP{
								"FieldAB": MAP{
									"FieldBaC": ARRAY{
										MAP{
											"FieldC1": "peter",
										},
									},
								},
							},
						},
					},
				}
				p := MustFieldPath(".FieldAB.FieldBapA[].FieldAB.FieldBaC[].FieldC1")
				Expect(p.Get(v)).To(Equal(ARRAY{ARRAY{"alice", "bob"}, ARRAY{"peter"}}))
			})
		})

		//////////////////////////

		Context("and invalid expressions", func() {
			It("simple field", func() {
				v := MAP{}
				p := MustFieldPath(".A")
				Expect(p.Get(v)).To(Equal(Unknown))
			})
		})

	})

	////////////////////////////////////////////////////////////////////////////////

	Context("Static", func() {
		Context("Type", func() {
			It("simple field", func() {
				p := MustFieldPath(".FieldC1")

				v := &C{}
				Expect(p.VType(v)).To(Equal(reflect.TypeOf(v.FieldC1)))
			})
			It("nested field", func() {
				p := MustFieldPath(".FieldCpB.FieldBS")

				v := &C{}
				Expect(p.VType(v)).To(Equal(reflect.TypeOf("")))
				Expect(v.FieldCpB).To(BeNil())
			})
			It("index", func() {
				p := MustFieldPath(".FieldAB.FieldBaC[1].FieldC1")

				v := &A{}
				Expect(p.VType(v)).To(Equal(reflect.TypeOf("")))
				Expect(v.FieldAB.FieldBaC).To(BeNil())
			})
			It("index interface", func() {
				p := MustFieldPath(".FieldAB.FieldBaI[1].FieldC1")

				v := &A{}
				t, err := p.VType(v)
				//fmt.Printf("type %s\n", t)
				Expect(t, err).To(Equal(Unknown))
				Expect(v.FieldAB.FieldBaC).To(BeNil())
			})
		})

		Context("and valid expressions", func() {
			It("simple field", func() {
				p := MustFieldPath(".FieldC1")

				v := &C{
					FieldC1: "fieldc1",
				}
				Expect(p.Get(v)).To(Equal("fieldc1"))
			})

			It("nested field", func() {
				v := &A{
					FieldAC: C{
						FieldC1: "test",
					},
				}
				p := MustFieldPath(".FieldAC.FieldC1")
				Expect(p.Get(v)).To(Equal("test"))
			})

			It("projection", func() {
				v := &A{
					FieldAB: B{
						FieldBaC: []C{
							C{
								FieldC1: "test",
							},
						},
					},
				}
				p := MustFieldPath(".FieldAB.FieldBaC[].FieldC1")
				Expect(p.Get(v)).To(Equal([]string{"test"}))
			})

			It("empty projection", func() {
				v := &A{
					FieldAB: B{
						FieldBaC: []C{},
					},
				}
				p := MustFieldPath(".FieldAB.FieldBaC[].FieldC1")
				Expect(p.Get(v)).To(BeNil())
			})

			It("double projection", func() {
				v := &A{
					FieldAB: B{
						FieldBaaC: [][]C{
							[]C{
								C{FieldC1: "alice"},
							},
							[]C{
								C{FieldC1: "bob"},
							},
						},
					},
				}
				p := MustFieldPath(".FieldAB.FieldBaaC[][].FieldC1")
				Expect(p.Get(v)).To(Equal([][]string{[]string{"alice"}, []string{"bob"}}))
			})

			It("nested double projection", func() {
				v := &A{
					FieldAB: B{
						FieldBapA: []*A{
							&A{
								FieldAB: B{
									FieldBaC: []C{
										C{
											FieldC1: "alice",
										},
										C{
											FieldC1: "bob",
										},
									},
								},
							},
							&A{
								FieldAB: B{
									FieldBaC: []C{
										C{
											FieldC1: "peter",
										},
									},
								},
							},
						},
					},
				}
				p := MustFieldPath(".FieldAB.FieldBapA[].FieldAB.FieldBaC[].FieldC1")
				Expect(p.Get(v)).To(Equal([][]string{[]string{"alice", "bob"}, []string{"peter"}}))
			})
		})

		//////////////////////////

		Context("and invalid expressions", func() {
			It("simple field", func() {
				v := &C{
					FieldC2: MAP{
						"A": "A",
					},
				}
				p := MustFieldPath(".FieldC2.B")
				Expect(p.Get(v)).To(Equal(Unknown))
			})
		})

		It("nested field", func() {
			v := &A{
				FieldAC: C{},
			}
			p := MustFieldPath(".FieldAC.FieldCpB.FieldBaC")
			Expect(p.Get(v)).To(Equal(Unknown))
		})

		It("projection", func() {
			v := &A{
				FieldAB: B{
					// FieldBaC: []C{ C{FieldC1: "test"}},
				},
			}
			p := MustFieldPath(".FieldAB.FieldBaC[].FieldC1")
			Expect(p.Get(v)).To(BeNil())
		})

		//////////////////////////

		Context("and undefined expressions", func() {
			It("simple field", func() {
				v := &C{}
				p := MustFieldPath(".FieldC3")
				r, err := p.Get(v)
				Expect(r).To(BeNil())
				Expect(err).To(Equal(fmt.Errorf("<object> has no field \"FieldC3\"")))
			})
		})

		It("nested field", func() {
			v := &A{}
			p := MustFieldPath(".FieldAC.FieldC3")
			r, err := p.Get(v)
			Expect(r).To(BeNil())
			Expect(err).To(Equal(fmt.Errorf("<object>.FieldAC has no field \"FieldC3\"")))
		})

		It("nested field2", func() {
			v := &A{
				FieldAC: C{},
			}
			p := MustFieldPath(".FieldAC.FieldCpB.FieldX")
			r, err := p.Get(v)
			Expect(r).To(BeNil())
			Expect(err).To(Equal(fmt.Errorf("<object>.FieldAC.FieldCpB has no field \"FieldX\"")))
		})

		It("projection", func() {
			v := &A{
				FieldAB: B{
					FieldBaC: []C{
						C{
							FieldC1: "test",
						},
					},
				},
			}
			p := MustFieldPath(".FieldAB.FieldBaC[].FieldC3")
			r, err := p.Get(v)
			Expect(r).To(BeNil())
			Expect(err).To(Equal(fmt.Errorf("<object>.FieldAB.FieldBaC[] has no field \"FieldC3\"")))
		})
		It("double projection", func() {
			v := &A{
				FieldAB: B{},
			}
			p := MustFieldPath(".FieldAB.FieldBaaC[][].FieldC3")
			r, err := p.Get(v)
			Expect(r).To(BeNil())
			Expect(err).To(Equal(fmt.Errorf("<object>.FieldAB.FieldBaaC[][] has no field \"FieldC3\"")))
		})
	})

	Context("values", func() {
		It("simple field", func() {
			v := &C{
				FieldC1: "alice",
			}
			p := MustFieldPath(".FieldC1")
			Expect(Values(p, v)).To(Equal([]interface{}{"alice"}))
		})

		It("projection", func() {
			v := &A{
				FieldAB: B{
					FieldBaC: []C{
						C{
							FieldC1: "bob",
						},
					},
				},
			}
			p := MustFieldPath(".FieldAB.FieldBaC[].FieldC1")
			Expect(Values(p, v)).To(Equal([]interface{}{"bob"}))
		})

		It("empty projection", func() {
			v := &A{
				FieldAB: B{
					FieldBaC: []C{},
				},
			}
			p := MustFieldPath(".FieldAB.FieldBaC[].FieldC1")
			Expect(Values(p, v)).To(BeNil())
		})

		It("double projection", func() {
			v := &A{
				FieldAB: B{
					FieldBaaC: [][]C{
						[]C{
							C{FieldC1: "alice"},
						},
						[]C{
							C{FieldC1: "bob"},
						},
					},
				},
			}
			p := MustFieldPath(".FieldAB.FieldBaaC[][].FieldC1")
			Expect(Values(p, v)).To(Equal([]interface{}{"alice", "bob"}))
		})

		It("nested double projection", func() {
			v := &A{
				FieldAB: B{
					FieldBapA: []*A{
						&A{
							FieldAB: B{
								FieldBaC: []C{
									C{
										FieldC1: "alice",
									},
									C{
										FieldC1: "bob",
									},
								},
							},
						},
						&A{
							FieldAB: B{
								FieldBaC: []C{
									C{
										FieldC1: "peter",
									},
								},
							},
						},
						&A{
							FieldAB: B{},
						},
					},
				},
			}
			p := MustFieldPath(".FieldAB.FieldBapA[].FieldAB.FieldBaC[].FieldC1")
			Expect(Values(p, v)).To(Equal([]interface{}{"alice", "bob", "peter"}))
		})
	})

	Context("set", func() {
		Context("dynamic", func() {
			It("simple field", func() {
				v := MAP{}
				p := MustFieldPath(".FieldAB")
				Expect(p.Set(v, "alice")).To(Succeed())
				Expect(p.Get(v)).To(Equal("alice"))
			})
			It("nested field (simple type)", func() {
				v := MAP{}
				p := MustFieldPath(".FieldAB.FieldBapA")
				Expect(p.Set(v, "alice")).To(Succeed())
				Expect(p.Get(v)).To(Equal("alice"))
			})
			It("nested field (array entry)", func() {
				v := MAP{}
				p := MustFieldPath(".FieldAB.FieldBapA[0]")
				Expect(p.Set(v, "alice")).To(Succeed())
				Expect(p.Get(v)).To(Equal("alice"))
			})
			It("nested array field", func() {
				v := MAP{}
				p := MustFieldPath(".FieldBapA[0].FieldAB")
				Expect(p.Set(v, "alice")).To(Succeed())
				Expect(p.Get(v)).To(Equal("alice"))
			})
		})
		Context("static", func() {
			It("nested index", func() {
				v := &A{}
				p := MustFieldPath(".FieldAB.FieldBapA[0].FieldAB.FieldBS")
				Expect(p.Set(v, "alice")).To(Succeed())
				Expect(v.FieldAB.FieldBapA[0].FieldAB.FieldBS).To(Equal("alice"))
				Expect(p.Get(v)).To(Equal("alice"))
			})
			It("nested double index", func() {
				v := &A{}
				p := MustFieldPath(".FieldAB.FieldBapA[0].FieldAB.FieldBaC[0].FieldC1")
				Expect(p.Set(v, "alice")).To(Succeed())
				Expect(v.FieldAB.FieldBapA[0].FieldAB.FieldBaC[0].FieldC1).To(Equal("alice"))
				Expect(p.Get(v)).To(Equal("alice"))
			})
			It("nested field", func() {
				v := &C{}
				p := MustFieldPath(".FieldCpB.FieldBS")
				Expect(p.Set(v, "alice")).To(Succeed())
				Expect(v.FieldCpB.FieldBS).To(Equal("alice"))
			})
			It("double pointer field", func() {
				v := &C{}
				p := MustFieldPath(".FieldCppB.FieldBS")
				Expect(p.Set(v, "alice")).To(Succeed())
				Expect((*v.FieldCppB).FieldBS).To(Equal("alice"))
			})
		})
	})
})
