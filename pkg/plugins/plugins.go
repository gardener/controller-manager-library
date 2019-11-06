/*
 * Copyright 2019 SAP SE or an SAP affiliate company. All rights reserved. This file is licensed under the Apache Software License, v. 2 except as noted otherwise in the LICENSE file
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *      http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 *
 */

package plugins

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"plugin"
	"strings"
	"sync"

	"github.com/gardener/controller-manager-library/pkg/logger"
)

type Plugin struct {
	*plugin.Plugin
	name string
	path string
}

func (this *Plugin) GetName() string {
	return this.name
}
func (this *Plugin) GetPath() string {
	return this.path
}

type Handler interface {
	HandlePlugin(p *Plugin) error
}

var handlers = []Handler{}
var lock sync.Mutex

func AddHandler(h Handler) {
	lock.Lock()
	defer lock.Unlock()

	handlers = append(handlers, h)
}

var plugins = map[string]*Plugin{}

func addPlugin(name string, p *plugin.Plugin, path string) error {
	lock.Lock()
	defer lock.Unlock()

	logger.Infof("loaded plugin %s from %s", name, path)
	pl := &Plugin{p, name, path}
	plugins[name] = pl
	for _, h := range handlers {
		err := h.HandlePlugin(pl)
		if err != nil {
			return err
		}
	}
	return nil
}

func LoadPlugins(dir string) error {
	files, err := ioutil.ReadDir(dir)
	if err != nil {
		return fmt.Errorf("cannot read plugin file '%s': %s", dir, err)
	}
	logger.Infof("scanning %s for plugins", dir)
	for _, f := range files {
		path := fmt.Sprintf("%s%c%s", dir, filepath.Separator, f.Name())
		if ok, _ := IsFile(path); ok {
			p, err := plugin.Open(path)
			if err != nil {
				logger.Errorf("cannot load plugin %s: %s", path, err)
			} else {
				v, err := p.Lookup("Name")
				if err != nil {
					logger.Errorf("loaded plugin %s has no name", path)
				} else {
					name, ok := v.(*string)
					if ok {
						err := addPlugin(*name, p, path)
						if err != nil {
							return err
						}

					} else {
						logger.Errorf("loaded plugin %s has invalid variable Name", path)
					}
				}
			}
		}
	}
	return nil
}

func HandleCommandLine(opt string, args []string) error {
	for i := range args {
		path := ""
		if strings.HasPrefix(args[i], opt+"=") {
			path = args[1][len(opt)+1:]
		} else {
			if args[i] == opt {
				if len(args) <= i+1 {
					return fmt.Errorf("missing argument for %s", opt)
				}
				path = args[i+1]
			}
		}
		if path != "" {
			if err := LoadPlugins(path); err != nil {
				return err
			}
			break
		}
	}
	return nil
}

func IsFile(path string) (bool, error) {
	fi, err := os.Stat(path)
	if err == nil {
		if fi.Mode().IsRegular() {
			return true, nil
		}
	}
	return false, err
}
