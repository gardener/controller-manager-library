/*
 * SPDX-FileCopyrightText: 2020 SAP SE or an SAP affiliate company and Gardener contributors
 *
 * SPDX-License-Identifier: Apache-2.0
 *
 */

package misc

import (
	"context"
	"fmt"
	"time"

	"github.com/gardener/controller-manager-library/pkg/controllermanager/extension"
	"github.com/gardener/controller-manager-library/pkg/utils"
	"github.com/gardener/controller-manager-library/pkg/wait"
)

type Elem struct {
	name  string
	after []string
}

func (this *Elem) Name() string {
	return this.name
}

func (this *Elem) After() []string {
	return this.after
}

func (this *Elem) Before() []string {
	return nil
}

func (this *Elem) String() string {
	return this.name
}

func MiscMain() {
	A := &Elem{"A", []string{"B", "C"}}
	B := &Elem{"B", []string{"D"}}
	C := &Elem{"C", []string{"D"}}
	D := &Elem{"D", []string{"B"}}

	o, _, err := extension.Order([]*Elem{
		A, B, C, D,
	})
	utils.Must(err)
	fmt.Printf("%+v\n", o)
}

func MiscBackoff() {
	b := wait.Backoff{
		Duration: time.Second,
		Factor:   1.1,
		Steps:    -1,
	}

	ctx, cancel := context.WithCancel(context.Background())
	fmt.Printf("start\n")

	timer := time.NewTimer(20 * time.Second)
	go func() {
		fmt.Printf("%s\n", wait.ExponentialBackoff(ctx, b, func() (bool, error) {
			fmt.Printf("%s tick...\n", time.Now())
			return false, nil
		}))
	}()
	fmt.Printf("wait\n")

	<-timer.C
	fmt.Printf("shutdown\n")
	cancel()

	timer.Reset(5 * time.Second)
	<-timer.C
	fmt.Printf("done\n")
}
