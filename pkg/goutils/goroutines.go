/*
 * SPDX-FileCopyrightText: 2019 SAP SE or an SAP affiliate company and Gardener contributors
 *
 * SPDX-License-Identifier: Apache-2.0
 *
 */

package goutils

import (
	"fmt"
	"runtime"
	"strconv"
	"strings"
)

var NumberOfGoRoutines = runtime.NumGoroutine

type Function struct {
	Name     string
	Location string
}
type GoRoutine struct {
	Id      int64
	Status  string
	Current Function
	First   Function
	Creator Function
	Stack   Stack
}

type Stack []*Function

type GoRoutines map[int64]*GoRoutine

func (this GoRoutines) String() string {
	return this.ToString("", true)
}

func (this GoRoutines) ToString(msg string, withStack bool) string {
	result := ""
	for _, r := range this {
		result = msg + ":"
		result = fmt.Sprintf("%s\n  %-4d [%s]", result, r.Id, r.Status)
		result = fmt.Sprintf("%s\n      creator  %s", result, r.Creator.Location)
		result = fmt.Sprintf("%s\n      current  %s", result, r.Current.Location)
		if withStack {
			for _, s := range r.Stack {
				result = fmt.Sprintf("%s\n      stack    %s", result, s.Location)
			}
		}
	}
	return result
}

func setFirst(prev *GoRoutine, index int, lines []string) {
	if prev != nil {
		off := index - 3
		if strings.HasPrefix(lines[off], "created by ") {
			prev.Creator.Name = lines[off][11:]
			prev.Creator.Location = strings.TrimSpace(lines[off+1])
			off -= 2
		}
		if off >= 0 {
			prev.First.Name = lines[off]
			prev.First.Location = strings.TrimSpace(lines[off+1])
		}
	}
}

func ListGoRoutines(withStack bool) GoRoutines {
	m := GoRoutines{}
	buf := [100000]byte{}
	n := runtime.Stack(buf[:], true)
	//	fmt.Printf("---\n%s\n---\n", buf[:n])
	lines := strings.Split(string(buf[:n]), "\n")
	start := 0
	var prev *GoRoutine
	for i, l := range lines {
		if strings.HasPrefix(l, "goroutine ") {
			setFirst(prev, i, lines)
			r := &GoRoutine{}
			idx := strings.Index(l, "[")
			if idx >= 0 {
				r.Id, _ = strconv.ParseInt(strings.TrimSpace(l[9:idx]), 10, 64)
				idx2 := strings.Index(l[idx:], "]")
				if idx2 >= 0 {
					r.Status = l[idx+1 : idx+idx2]
				}
			}
			r.Current.Name = lines[i+1]
			r.Current.Location = strings.TrimSpace(lines[i+2])
			start = i + 1
			m[r.Id] = r
			prev = r
		} else {
			if withStack && i == start && prev != nil && l != "" && !strings.HasPrefix(l, "created by ") {
				prev.Stack = append(prev.Stack, &Function{
					Name:     lines[i],
					Location: strings.TrimSpace(lines[i+1]),
				})
				start = i + 2
			}
		}
	}
	setFirst(prev, len(lines), lines)
	return m
}

func GoRoutineDiff(a, b GoRoutines) (add GoRoutines, del GoRoutines) {
	add = GoRoutines{}
	del = GoRoutines{}

	for id, r := range a {
		if b[id] == nil {
			del[id] = r
		}
	}
	for id, r := range b {
		if a[id] == nil {
			add[id] = r
		}
	}
	return
}
