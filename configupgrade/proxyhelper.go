// Copyright (c) 2024 Tulir Asokan
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at http://mozilla.org/MPL/2.0/.

package configupgrade

type ProxyHelper struct {
	Prefix []string
	Target Helper
}

var _ Helper = (*ProxyHelper)(nil)

func (p *ProxyHelper) Copy(allowedTypes YAMLType, path ...string) {
	p.Target.Copy(allowedTypes, append(p.Prefix, path...)...)
}

func (p *ProxyHelper) Get(tag YAMLType, path ...string) (string, bool) {
	return p.Target.Get(tag, append(p.Prefix, path...)...)
}

func (p *ProxyHelper) GetBase(path ...string) string {
	return p.Target.GetBase(append(p.Prefix, path...)...)
}

func (p *ProxyHelper) GetNode(path ...string) *YAMLNode {
	return p.Target.GetNode(append(p.Prefix, path...)...)
}

func (p *ProxyHelper) GetBaseNode(path ...string) *YAMLNode {
	return p.Target.GetBaseNode(append(p.Prefix, path...)...)
}

func (p *ProxyHelper) Set(tag YAMLType, value string, path ...string) {
	p.Target.Set(tag, value, append(p.Prefix, path...)...)
}

func (p *ProxyHelper) SetMap(value YAMLMap, path ...string) {
	p.Target.SetMap(value, append(p.Prefix, path...)...)
}

func (p *ProxyHelper) AddSpaceBeforeComment(path ...string) {
	p.Target.AddSpaceBeforeComment(append(p.Prefix, path...)...)
}
