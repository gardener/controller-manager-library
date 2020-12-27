/*
 * SPDX-FileCopyrightText: 2019 SAP SE or an SAP affiliate company and Gardener contributors
 *
 * SPDX-License-Identifier: Apache-2.0
 *
 */

package recover

import (
	"fmt"

	"github.com/gardener/controller-manager-library/pkg/controllermanager/controller/reconcile"
	"github.com/gardener/controller-manager-library/pkg/logger"
)

func RecoverMain() {

	call1()
}

func call1() {

	result := SaveAction()
	fmt.Printf("found result: %v\n", result)
}

func SaveAction() (result reconcile.Status) {

	defer func() {
		if r := recover(); r != nil {
			if res, ok := r.(reconcile.Status); ok {
				result = res
			} else {
				panic(r)
			}
		}
	}()

	return function(2)
}

func AbortAndDelayOnError(err error) {
	if err != nil {
		panic(reconcile.Delay(logger.New(), err))
	}
}

func function(mode int) reconcile.Status {
	var err error
	switch mode {
	case 0:
	case 1:
		err = fmt.Errorf("this was an error")
	default:
		panic("fatal error")
	}
	AbortAndDelayOnError(err)
	fmt.Printf("normal processing\n")
	return reconcile.Succeeded(nil)
}
