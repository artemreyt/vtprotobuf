// Copyright (c) 2021 PlanetScale Inc. All rights reserved.
// Copyright (c) 2013, The GoGo Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package marshal

import (
	"sort"

	"github.com/artemreyt/vtprotobuf/generator"
	"google.golang.org/protobuf/compiler/protogen"
	"google.golang.org/protobuf/reflect/protoreflect"
)

func init() {
	generator.RegisterFeature("flat_oneofs", func(gen *generator.GeneratedFile) generator.FeatureGenerator {
		return &flat{GeneratedFile: gen, marshal: &marshal{GeneratedFile: gen, Stable: false, flat: true}}
	})
}

// flat generation feature generates additional method
type flat struct {
	*marshal
	*generator.GeneratedFile
	once bool
}

var _ generator.FeatureGenerator = (*flat)(nil)

func (p *flat) Name() string {
	return "flat_oneofs"
}

func (p *flat) GenerateFile(file *protogen.File) bool {
	proto3 := file.Desc.Syntax() == protoreflect.Proto3
	for _, message := range file.Messages {
		p.message(proto3, message)
	}
	return p.once
}

func (p *flat) GenerateHelpers() {
}

func (p *flat) message(proto3 bool, message *protogen.Message) {
	for _, nested := range message.Messages {
		p.message(proto3, nested)
	}

	if message.Desc.IsMapEntry() {
		return
	}

	p.once = true

	var numGen counter
	ccTypeName := message.GoIdent

	p.P(`func (m *`, ccTypeName, `) MarshalVTFlat() (dAtA []byte, err error) {`)
	p.P(`if m == nil {`)
	p.P(`return nil, nil`)
	p.P(`}`)
	p.P(`size := m.SizeVT()`)
	p.P(`dAtA = make([]byte, size)`)
	p.P(`n, err := m.MarshalToSizedBufferVTFlat(dAtA[:size])`)
	p.P(`if err != nil {`)
	p.P(`return nil, err`)
	p.P(`}`)
	p.P(`return dAtA[:n], nil`)
	p.P(`}`)
	p.P(``)
	p.P(`func (m *`, ccTypeName, `) MarshalToVTFlat(dAtA []byte) (int, error) {`)
	p.P(`size := m.SizeVT()`)
	p.P(`return m.MarshalToSizedBufferVTFlat(dAtA[:size])`)
	p.P(`}`)
	p.P(``)
	p.P(`func (m *`, ccTypeName, `) MarshalToSizedBufferVTFlat(dAtA []byte) (int, error) {`)
	p.P(`if m == nil {`)
	p.P(`return 0, nil`)
	p.P(`}`)
	p.P(`i := len(dAtA)`)
	p.P(`_ = i`)
	p.P(`var l int`)
	p.P(`_ = l`)
	p.P(`if m.unknownFields != nil {`)
	p.P(`i -= len(m.unknownFields)`)
	p.P(`copy(dAtA[i:], m.unknownFields)`)
	p.P(`}`)

	sort.Slice(message.Fields, func(i, j int) bool {
		return message.Fields[i].Desc.Number() < message.Fields[j].Desc.Number()
	})

	for i := len(message.Fields) - 1; i >= 0; i-- {
		field := message.Fields[i]
		oneof := field.Oneof != nil && !field.Oneof.Desc.IsSynthetic()
		if !oneof {
			p.field(proto3, false, &numGen, field)
		} else {
			p.P(`if msg, ok := m.`, field.Oneof.GoName, `.(*`, field.GoIdent.GoName, `); ok {`)
			p.P(`size, err := msg.MarshalToSizedBufferVTFlat(dAtA[:i])`)
			p.P(`if err != nil {`)
			p.P(`return 0, err`)
			p.P(`}`)
			p.P(`i -= size`)
			p.P(`}`)
		}
	}
	p.P(`return len(dAtA) - i, nil`)
	p.P(`}`)
	p.P()

	//Generate MarshalToVTFlat methods for oneof fields
	for _, field := range message.Fields {
		if field.Oneof == nil || field.Oneof.Desc.IsSynthetic() {
			continue
		}
		ccTypeName := field.GoIdent
		p.P(`func (m *`, ccTypeName, `) MarshalToVTFlat(dAtA []byte) (int, error) {`)
		p.P(`size := m.SizeVT()`)
		p.P(`return m.MarshalToSizedBufferVTFlat(dAtA[:size])`)
		p.P(`}`)
		p.P(``)
		p.P(`func (m *`, ccTypeName, `) MarshalToSizedBufferVTFlat(dAtA []byte) (int, error) {`)
		p.P(`i := len(dAtA)`)
		p.field(proto3, true, &numGen, field)
		p.P(`return len(dAtA) - i, nil`)
		p.P(`}`)
	}
}
