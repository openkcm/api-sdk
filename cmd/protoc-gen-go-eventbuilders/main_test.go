// SPDX-FileCopyrightText: 2025 SAP SE or an SAP affiliate company and OpenKCM contributors
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"bytes"
	"strings"
	"testing"

	"google.golang.org/protobuf/compiler/protogen"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/types/descriptorpb"
	"google.golang.org/protobuf/types/pluginpb"
)

func TestConsumeVarint(t *testing.T) {
	tests := []struct {
		name     string
		data     []byte
		wantVal  uint64
		wantN    int
		wantFail bool
	}{
		{
			name:    "single byte zero",
			data:    []byte{0x00},
			wantVal: 0,
			wantN:   1,
		},
		{
			name:    "single byte value",
			data:    []byte{0x01},
			wantVal: 1,
			wantN:   1,
		},
		{
			name:    "single byte max",
			data:    []byte{0x7F},
			wantVal: 127,
			wantN:   1,
		},
		{
			name:    "two byte value 128",
			data:    []byte{0x80, 0x01},
			wantVal: 128,
			wantN:   2,
		},
		{
			name:    "two byte value 300",
			data:    []byte{0xAC, 0x02},
			wantVal: 300,
			wantN:   2,
		},
		{
			name:    "larger varint",
			data:    []byte{0x96, 0x01}, // 150
			wantVal: 150,
			wantN:   2,
		},
		{
			name:    "five byte value",
			data:    []byte{0xFF, 0xFF, 0xFF, 0xFF, 0x0F},
			wantVal: 0xFFFFFFFF,
			wantN:   5,
		},
		{
			name:     "empty data",
			data:     []byte{},
			wantVal:  0,
			wantN:    -1,
			wantFail: true,
		},
		{
			name:     "incomplete varint (all continuation bits)",
			data:     []byte{0x80, 0x80, 0x80, 0x80, 0x80, 0x80, 0x80, 0x80, 0x80, 0x80},
			wantVal:  0,
			wantN:    -1,
			wantFail: true,
		},
		{
			name:    "data with trailing bytes",
			data:    []byte{0x05, 0xFF, 0xFF},
			wantVal: 5,
			wantN:   1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotVal, gotN := consumeVarint(tt.data)
			if tt.wantFail {
				if gotN >= 0 {
					t.Errorf("consumeVarint() expected failure, got n=%d", gotN)
				}

				return
			}

			if gotVal != tt.wantVal {
				t.Errorf("consumeVarint() val = %d, want %d", gotVal, tt.wantVal)
			}

			if gotN != tt.wantN {
				t.Errorf("consumeVarint() n = %d, want %d", gotN, tt.wantN)
			}
		})
	}
}

func TestSkipField(t *testing.T) {
	tests := []struct {
		name     string
		data     []byte
		wireType int
		wantLen  int // length of remaining data, -1 for nil
	}{
		// Wire type 0: Varint
		{
			name:     "varint single byte",
			data:     []byte{0x05, 0xAA, 0xBB},
			wireType: 0,
			wantLen:  2,
		},
		{
			name:     "varint two bytes",
			data:     []byte{0x80, 0x01, 0xAA},
			wireType: 0,
			wantLen:  1,
		},
		{
			name:     "varint incomplete",
			data:     []byte{0x80, 0x80, 0x80, 0x80, 0x80, 0x80, 0x80, 0x80, 0x80, 0x80},
			wireType: 0,
			wantLen:  -1,
		},
		// Wire type 1: 64-bit
		{
			name:     "64-bit fixed",
			data:     []byte{0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08, 0xAA},
			wireType: 1,
			wantLen:  1,
		},
		{
			name:     "64-bit insufficient data",
			data:     []byte{0x01, 0x02, 0x03},
			wireType: 1,
			wantLen:  -1,
		},
		// Wire type 2: Length-delimited
		{
			name:     "length-delimited",
			data:     []byte{0x03, 0xAA, 0xBB, 0xCC, 0xDD},
			wireType: 2,
			wantLen:  1,
		},
		{
			name:     "length-delimited empty",
			data:     []byte{0x00, 0xAA},
			wireType: 2,
			wantLen:  1,
		},
		{
			name:     "length-delimited insufficient",
			data:     []byte{0x05, 0xAA, 0xBB},
			wireType: 2,
			wantLen:  -1,
		},
		{
			name:     "length-delimited bad length varint",
			data:     []byte{},
			wireType: 2,
			wantLen:  -1,
		},
		// Wire type 5: 32-bit
		{
			name:     "32-bit fixed",
			data:     []byte{0x01, 0x02, 0x03, 0x04, 0xAA},
			wireType: 5,
			wantLen:  1,
		},
		{
			name:     "32-bit insufficient data",
			data:     []byte{0x01, 0x02},
			wireType: 5,
			wantLen:  -1,
		},
		// Unknown wire type
		{
			name:     "unknown wire type 3",
			data:     []byte{0x01, 0x02, 0x03},
			wireType: 3,
			wantLen:  -1,
		},
		{
			name:     "unknown wire type 4",
			data:     []byte{0x01, 0x02, 0x03},
			wireType: 4,
			wantLen:  -1,
		},
		{
			name:     "unknown wire type 6",
			data:     []byte{0x01, 0x02, 0x03},
			wireType: 6,
			wantLen:  -1,
		},
		{
			name:     "unknown wire type 7",
			data:     []byte{0x01, 0x02, 0x03},
			wireType: 7,
			wantLen:  -1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := skipField(tt.data, tt.wireType)
			if tt.wantLen == -1 {
				if got != nil {
					t.Errorf("skipField() = %v, want nil", got)
				}

				return
			}

			if got == nil {
				t.Errorf("skipField() = nil, want len=%d", tt.wantLen)
				return
			}

			if len(got) != tt.wantLen {
				t.Errorf("skipField() remaining len = %d, want %d", len(got), tt.wantLen)
			}
		})
	}
}

func TestToParamName(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"", ""},
		{"A", "a"},
		{"DataSubjectType", "dataSubjectType"},
		{"ID", "iD"},
		{"UserId", "userId"},
		{"XMLParser", "xMLParser"},
		{"simpleCase", "simpleCase"},
		{"X", "x"},
		{"ABC", "aBC"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got := toParamName(tt.input)
			if got != tt.want {
				t.Errorf("toParamName(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

func TestParseEmbeddedForRequired(t *testing.T) {
	tests := []struct {
		name     string
		data     []byte
		fieldNum int
		want     bool
	}{
		{
			name:     "empty data",
			data:     []byte{},
			fieldNum: 25,
			want:     false,
		},
		{
			name: "field 25 set to true",
			// tag = 25 << 3 | 0 = 200 = 0xC8, value = 1
			data:     []byte{0xC8, 0x01, 0x01},
			fieldNum: 25,
			want:     true,
		},
		{
			name: "field 25 set to false",
			// tag = 25 << 3 | 0 = 200 = 0xC8, value = 0
			data:     []byte{0xC8, 0x01, 0x00},
			fieldNum: 25,
			want:     false,
		},
		{
			name: "different field number",
			// tag = 10 << 3 | 0 = 80 = 0x50, value = 1
			data:     []byte{0x50, 0x01},
			fieldNum: 25,
			want:     false,
		},
		{
			name:     "incomplete tag",
			data:     []byte{0x80, 0x80}, // incomplete varint
			fieldNum: 25,
			want:     false,
		},
		{
			name: "wrong wire type for field 25",
			// tag = 25 << 3 | 2 = 202 = 0xCA (length-delimited instead of varint)
			data:     []byte{0xCA, 0x01, 0x03, 0xAA, 0xBB, 0xCC},
			fieldNum: 25,
			want:     false,
		},
		{
			name: "field found but value varint incomplete",
			// tag = 25 << 3 | 0 = 200, but no value follows
			data:     []byte{0xC8, 0x01},
			fieldNum: 25,
			want:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := parseEmbeddedForRequired(tt.data, tt.fieldNum)
			if got != tt.want {
				t.Errorf("parseEmbeddedForRequired() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestParseUnknownForRequired(t *testing.T) {
	tests := []struct {
		name             string
		data             []byte
		extFieldNum      int
		requiredFieldNum int
		want             bool
	}{
		{
			name:             "empty data",
			data:             []byte{},
			extFieldNum:      1159,
			requiredFieldNum: 25,
			want:             false,
		},
		{
			name: "extension field with required=true",
			// tag = 1159 << 3 | 2 = 9274 (length-delimited)
			// 9274 in varint = 0xBA 0x48
			// length = 3, embedded = [0xC8, 0x01, 0x01] (field 25 = true)
			data:             []byte{0xBA, 0x48, 0x03, 0xC8, 0x01, 0x01},
			extFieldNum:      1159,
			requiredFieldNum: 25,
			want:             true,
		},
		{
			name: "extension field with required=false",
			// same but value = 0
			data:             []byte{0xBA, 0x48, 0x03, 0xC8, 0x01, 0x00},
			extFieldNum:      1159,
			requiredFieldNum: 25,
			want:             false,
		},
		{
			name: "different extension field",
			// tag = 1000 << 3 | 2 = 8002 (length-delimited)
			// 8002 in varint = 0xD2 0x3E
			data:             []byte{0xD2, 0x3E, 0x03, 0xC8, 0x01, 0x01},
			extFieldNum:      1159,
			requiredFieldNum: 25,
			want:             false,
		},
		{
			name: "wrong wire type for extension",
			// tag = 1159 << 3 | 0 = 9272 (varint instead of length-delimited)
			// 9272 in varint = 0xB8 0x48
			data:             []byte{0xB8, 0x48, 0x01},
			extFieldNum:      1159,
			requiredFieldNum: 25,
			want:             false,
		},
		{
			name:             "incomplete tag",
			data:             []byte{0x80, 0x80, 0x80, 0x80, 0x80, 0x80, 0x80, 0x80, 0x80, 0x80},
			extFieldNum:      1159,
			requiredFieldNum: 25,
			want:             false,
		},
		{
			name: "length exceeds data",
			// tag = 1159 << 3 | 2 = 9274, length = 100 but only 3 bytes follow
			data:             []byte{0xBA, 0x48, 0x64, 0x01, 0x02, 0x03},
			extFieldNum:      1159,
			requiredFieldNum: 25,
			want:             false,
		},
		{
			name: "skip varint field then find extension",
			// First: tag=1 wire=0 (varint), value=5
			// Then: extension field with required=true
			data:             []byte{0x08, 0x05, 0xBA, 0x48, 0x03, 0xC8, 0x01, 0x01},
			extFieldNum:      1159,
			requiredFieldNum: 25,
			want:             true,
		},
		{
			name: "skip 64-bit field then find extension",
			// First: tag=1 wire=1 (64-bit)
			// Then: extension field with required=true
			data:             []byte{0x09, 0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08, 0xBA, 0x48, 0x03, 0xC8, 0x01, 0x01},
			extFieldNum:      1159,
			requiredFieldNum: 25,
			want:             true,
		},
		{
			name: "skip 32-bit field then find extension",
			// First: tag=1 wire=5 (32-bit)
			// Then: extension field with required=true
			data:             []byte{0x0D, 0x01, 0x02, 0x03, 0x04, 0xBA, 0x48, 0x03, 0xC8, 0x01, 0x01},
			extFieldNum:      1159,
			requiredFieldNum: 25,
			want:             true,
		},
		{
			name: "skip length-delimited field then find extension",
			// First: tag=1 wire=2 (length-delimited), length=2, data
			// Then: extension field with required=true
			data:             []byte{0x0A, 0x02, 0xAA, 0xBB, 0xBA, 0x48, 0x03, 0xC8, 0x01, 0x01},
			extFieldNum:      1159,
			requiredFieldNum: 25,
			want:             true,
		},
		{
			name: "unable to skip unknown wire type",
			// tag=1 wire=3 (start group - deprecated/unsupported)
			data:             []byte{0x0B, 0x01, 0x02},
			extFieldNum:      1159,
			requiredFieldNum: 25,
			want:             false,
		},
		{
			name: "incomplete length varint in extension",
			// tag = 1159 << 3 | 2, then incomplete varint for length
			data:             []byte{0xBA, 0x48, 0x80, 0x80, 0x80, 0x80, 0x80, 0x80, 0x80, 0x80, 0x80, 0x80},
			extFieldNum:      1159,
			requiredFieldNum: 25,
			want:             false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := parseUnknownForRequired(tt.data, tt.extFieldNum, tt.requiredFieldNum)
			if got != tt.want {
				t.Errorf("parseUnknownForRequired() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGoTypeFromKind(t *testing.T) {
	tests := []struct {
		kind protoreflect.Kind
		want string
	}{
		{protoreflect.StringKind, "string"},
		{protoreflect.BoolKind, "bool"},
		{protoreflect.Int32Kind, "int32"},
		{protoreflect.Sint32Kind, "int32"},
		{protoreflect.Sfixed32Kind, "int32"},
		{protoreflect.Int64Kind, "int64"},
		{protoreflect.Sint64Kind, "int64"},
		{protoreflect.Sfixed64Kind, "int64"},
		{protoreflect.Uint32Kind, "uint32"},
		{protoreflect.Fixed32Kind, "uint32"},
		{protoreflect.Uint64Kind, "uint64"},
		{protoreflect.Fixed64Kind, "uint64"},
		{protoreflect.FloatKind, "float32"},
		{protoreflect.DoubleKind, "float64"},
		{protoreflect.BytesKind, "[]byte"},
		// Message and Enum are handled by goType, not goTypeFromKind
		{protoreflect.MessageKind, "interface{}"},
		{protoreflect.EnumKind, "interface{}"},
		// Unknown kinds
		{protoreflect.GroupKind, "interface{}"},
	}

	for _, tt := range tests {
		t.Run(tt.kind.String(), func(t *testing.T) {
			got := goTypeFromKind(tt.kind)
			if got != tt.want {
				t.Errorf("goTypeFromKind(%v) = %q, want %q", tt.kind, got, tt.want)
			}
		})
	}
}

// createTestRequest creates a CodeGeneratorRequest for testing the plugin.
func createTestRequest(files []*descriptorpb.FileDescriptorProto, filesToGenerate []string) *pluginpb.CodeGeneratorRequest {
	return &pluginpb.CodeGeneratorRequest{
		FileToGenerate: filesToGenerate,
		ProtoFile:      files,
	}
}

func TestGenerateFile_NoCommonMessage(t *testing.T) {
	// Create a file without a Common message - should produce no output
	fileDesc := &descriptorpb.FileDescriptorProto{
		Name:    new("test.proto"),
		Package: new("test"),
		Options: &descriptorpb.FileOptions{
			GoPackage: new("example.com/test;test"),
		},
		MessageType: []*descriptorpb.DescriptorProto{
			{
				Name: new("SomeMessage"),
				Field: []*descriptorpb.FieldDescriptorProto{
					{
						Name:   new("id"),
						Number: proto.Int32(1),
						Type:   descriptorpb.FieldDescriptorProto_TYPE_STRING.Enum(),
						Label:  descriptorpb.FieldDescriptorProto_LABEL_OPTIONAL.Enum(),
					},
				},
			},
		},
	}

	req := createTestRequest([]*descriptorpb.FileDescriptorProto{fileDesc}, []string{"test.proto"})

	// Marshal request to simulate stdin
	reqBytes, err := proto.Marshal(req)
	if err != nil {
		t.Fatalf("failed to marshal request: %v", err)
	}

	// Use Options.New to create plugin from request bytes
	var outputBuf bytes.Buffer

	opts := protogen.Options{} //nolint:exhaustruct_v5

	gen, err := opts.New(req)
	if err != nil {
		t.Fatalf("failed to create plugin: %v", err)
	}

	gen.SupportedFeatures = uint64(pluginpb.CodeGeneratorResponse_FEATURE_PROTO3_OPTIONAL) |
		uint64(pluginpb.CodeGeneratorResponse_FEATURE_SUPPORTS_EDITIONS)
	gen.SupportedEditionsMinimum = descriptorpb.Edition_EDITION_2023
	gen.SupportedEditionsMaximum = descriptorpb.Edition_EDITION_2024

	for _, f := range gen.Files {
		if !f.Generate {
			continue
		}

		generateFile(gen, f)
	}

	resp := gen.Response()

	respBytes, err := proto.Marshal(resp)
	if err != nil {
		t.Fatalf("failed to marshal response: %v", err)
	}

	outputBuf.Write(respBytes)

	// Verify no files were generated (no Common message)
	if len(resp.GetFile()) != 0 {
		t.Errorf("expected no files generated for proto without Common, got %d", len(resp.GetFile()))
	}

	// Verify reqBytes is used
	_ = reqBytes
}

func TestGenerateFile_WithCommonAndEvents(t *testing.T) {
	// Create a file with Common message and event messages
	fileDesc := &descriptorpb.FileDescriptorProto{
		Name:    new("events.proto"),
		Package: new("events"),
		Syntax:  new("proto3"),
		Options: &descriptorpb.FileOptions{
			GoPackage: new("example.com/events;events"),
		},
		MessageType: []*descriptorpb.DescriptorProto{
			{
				Name: new("Common"),
				Field: []*descriptorpb.FieldDescriptorProto{
					{
						Name:   new("timestamp"),
						Number: proto.Int32(1),
						Type:   descriptorpb.FieldDescriptorProto_TYPE_STRING.Enum(),
						Label:  descriptorpb.FieldDescriptorProto_LABEL_OPTIONAL.Enum(),
					},
				},
			},
			{
				Name: new("SimpleEvent"),
				Field: []*descriptorpb.FieldDescriptorProto{
					{
						Name:     new("common"),
						Number:   proto.Int32(1),
						Type:     descriptorpb.FieldDescriptorProto_TYPE_MESSAGE.Enum(),
						TypeName: new(".events.Common"),
						Label:    descriptorpb.FieldDescriptorProto_LABEL_OPTIONAL.Enum(),
					},
				},
			},
			{
				Name: new("EventWithFields"),
				Field: []*descriptorpb.FieldDescriptorProto{
					{
						Name:     new("common"),
						Number:   proto.Int32(1),
						Type:     descriptorpb.FieldDescriptorProto_TYPE_MESSAGE.Enum(),
						TypeName: new(".events.Common"),
						Label:    descriptorpb.FieldDescriptorProto_LABEL_OPTIONAL.Enum(),
					},
					{
						Name:   new("user_id"),
						Number: proto.Int32(2),
						Type:   descriptorpb.FieldDescriptorProto_TYPE_STRING.Enum(),
						Label:  descriptorpb.FieldDescriptorProto_LABEL_OPTIONAL.Enum(),
					},
				},
			},
		},
	}

	req := createTestRequest([]*descriptorpb.FileDescriptorProto{fileDesc}, []string{"events.proto"})

	opts := protogen.Options{} //nolint:exhaustruct_v5

	gen, err := opts.New(req)
	if err != nil {
		t.Fatalf("failed to create plugin: %v", err)
	}

	gen.SupportedFeatures = uint64(pluginpb.CodeGeneratorResponse_FEATURE_PROTO3_OPTIONAL) |
		uint64(pluginpb.CodeGeneratorResponse_FEATURE_SUPPORTS_EDITIONS)
	gen.SupportedEditionsMinimum = descriptorpb.Edition_EDITION_2023
	gen.SupportedEditionsMaximum = descriptorpb.Edition_EDITION_2024

	for _, f := range gen.Files {
		if !f.Generate {
			continue
		}

		generateFile(gen, f)
	}

	resp := gen.Response()

	// Verify at least one file was generated
	if len(resp.GetFile()) == 0 {
		t.Fatal("expected at least one file to be generated")
	}

	// Check the generated content
	var foundContent bool

	for _, f := range resp.GetFile() {
		if f.Content != nil {
			content := f.GetContent()
			foundContent = true

			// Verify the generated content contains expected patterns
			if !strings.Contains(content, "package events") {
				t.Error("generated content should contain 'package events'")
			}

			if !strings.Contains(content, "NewSimpleEventEvent") {
				t.Error("generated content should contain 'NewSimpleEventEvent'")
			}

			if !strings.Contains(content, "NewEventWithFieldsEvent") {
				t.Error("generated content should contain 'NewEventWithFieldsEvent'")
			}

			if !strings.Contains(content, "SetCommon") {
				t.Error("generated content should contain 'SetCommon'")
			}

			if !strings.Contains(content, "Code generated by protoc-gen-go-eventbuilders") {
				t.Error("generated content should contain generator comment")
			}
		}
	}

	if !foundContent {
		t.Error("no content was generated")
	}
}

func TestGenerateFile_NoEventsWithCommon(t *testing.T) {
	// Create a file with Common message but no messages that use it
	fileDesc := &descriptorpb.FileDescriptorProto{
		Name:    new("noevents.proto"),
		Package: new("noevents"),
		Syntax:  new("proto3"),
		Options: &descriptorpb.FileOptions{
			GoPackage: new("example.com/noevents;noevents"),
		},
		MessageType: []*descriptorpb.DescriptorProto{
			{
				Name: new("Common"),
				Field: []*descriptorpb.FieldDescriptorProto{
					{
						Name:   new("timestamp"),
						Number: proto.Int32(1),
						Type:   descriptorpb.FieldDescriptorProto_TYPE_STRING.Enum(),
						Label:  descriptorpb.FieldDescriptorProto_LABEL_OPTIONAL.Enum(),
					},
				},
			},
			{
				Name: new("OtherMessage"),
				Field: []*descriptorpb.FieldDescriptorProto{
					{
						Name:   new("id"),
						Number: proto.Int32(1),
						Type:   descriptorpb.FieldDescriptorProto_TYPE_STRING.Enum(),
						Label:  descriptorpb.FieldDescriptorProto_LABEL_OPTIONAL.Enum(),
					},
				},
			},
		},
	}

	req := createTestRequest([]*descriptorpb.FileDescriptorProto{fileDesc}, []string{"noevents.proto"})

	opts := protogen.Options{} //nolint:exhaustruct_v5

	gen, err := opts.New(req)
	if err != nil {
		t.Fatalf("failed to create plugin: %v", err)
	}

	gen.SupportedFeatures = uint64(pluginpb.CodeGeneratorResponse_FEATURE_PROTO3_OPTIONAL) |
		uint64(pluginpb.CodeGeneratorResponse_FEATURE_SUPPORTS_EDITIONS)
	gen.SupportedEditionsMinimum = descriptorpb.Edition_EDITION_2023
	gen.SupportedEditionsMaximum = descriptorpb.Edition_EDITION_2024

	for _, f := range gen.Files {
		if !f.Generate {
			continue
		}

		generateFile(gen, f)
	}

	resp := gen.Response()

	// Verify no files were generated (no events with Common field)
	if len(resp.GetFile()) != 0 {
		t.Errorf("expected no files generated for proto without events, got %d", len(resp.GetFile()))
	}
}

func TestIsRequiredField_NilOptions(t *testing.T) {
	// Test isRequiredField with nil options by creating a minimal field descriptor
	fileDesc := &descriptorpb.FileDescriptorProto{
		Name:    new("test.proto"),
		Package: new("test"),
		Syntax:  new("proto3"),
		Options: &descriptorpb.FileOptions{
			GoPackage: new("example.com/test;test"),
		},
		MessageType: []*descriptorpb.DescriptorProto{
			{
				Name: new("TestMessage"),
				Field: []*descriptorpb.FieldDescriptorProto{
					{
						Name:   new("field_without_options"),
						Number: proto.Int32(1),
						Type:   descriptorpb.FieldDescriptorProto_TYPE_STRING.Enum(),
						Label:  descriptorpb.FieldDescriptorProto_LABEL_OPTIONAL.Enum(),
						// No options set
					},
				},
			},
		},
	}

	req := createTestRequest([]*descriptorpb.FileDescriptorProto{fileDesc}, []string{"test.proto"})

	opts := protogen.Options{} //nolint:exhaustruct_v5

	gen, err := opts.New(req)
	if err != nil {
		t.Fatalf("failed to create plugin: %v", err)
	}

	// Find the field and test isRequiredField
	for _, f := range gen.Files {
		for _, msg := range f.Messages {
			for _, field := range msg.Fields {
				// Should return false for field without options
				if isRequiredField(field) {
					t.Errorf("isRequiredField should return false for field without options")
				}
			}
		}
	}
}

func TestIsRequiredField_EmptyUnknownFields(t *testing.T) {
	// Test with empty field options (no unknown fields)
	fileDesc := &descriptorpb.FileDescriptorProto{
		Name:    new("test.proto"),
		Package: new("test"),
		Syntax:  new("proto3"),
		Options: &descriptorpb.FileOptions{
			GoPackage: new("example.com/test;test"),
		},
		MessageType: []*descriptorpb.DescriptorProto{
			{
				Name: new("TestMessage"),
				Field: []*descriptorpb.FieldDescriptorProto{
					{
						Name:    new("field_with_empty_options"),
						Number:  proto.Int32(1),
						Type:    descriptorpb.FieldDescriptorProto_TYPE_STRING.Enum(),
						Label:   descriptorpb.FieldDescriptorProto_LABEL_OPTIONAL.Enum(),
						Options: &descriptorpb.FieldOptions{}, // Empty options
					},
				},
			},
		},
	}

	req := createTestRequest([]*descriptorpb.FileDescriptorProto{fileDesc}, []string{"test.proto"})

	opts := protogen.Options{} //nolint:exhaustruct_v5

	gen, err := opts.New(req)
	if err != nil {
		t.Fatalf("failed to create plugin: %v", err)
	}

	for _, f := range gen.Files {
		for _, msg := range f.Messages {
			for _, field := range msg.Fields {
				if isRequiredField(field) {
					t.Errorf("isRequiredField should return false for field with empty options")
				}
			}
		}
	}
}

func TestGenerateFile_WithEnumField(t *testing.T) {
	// Create a file with an enum field to test goType for enums
	fileDesc := &descriptorpb.FileDescriptorProto{
		Name:    new("enum_test.proto"),
		Package: new("enumtest"),
		Syntax:  new("proto3"),
		Options: &descriptorpb.FileOptions{
			GoPackage: new("example.com/enumtest;enumtest"),
		},
		EnumType: []*descriptorpb.EnumDescriptorProto{
			{
				Name: new("Status"),
				Value: []*descriptorpb.EnumValueDescriptorProto{
					{Name: new("STATUS_UNSPECIFIED"), Number: proto.Int32(0)},
					{Name: new("STATUS_ACTIVE"), Number: proto.Int32(1)},
				},
			},
		},
		MessageType: []*descriptorpb.DescriptorProto{
			{
				Name: new("Common"),
				Field: []*descriptorpb.FieldDescriptorProto{
					{
						Name:   new("timestamp"),
						Number: proto.Int32(1),
						Type:   descriptorpb.FieldDescriptorProto_TYPE_STRING.Enum(),
						Label:  descriptorpb.FieldDescriptorProto_LABEL_OPTIONAL.Enum(),
					},
				},
			},
			{
				Name: new("EventWithEnum"),
				Field: []*descriptorpb.FieldDescriptorProto{
					{
						Name:     new("common"),
						Number:   proto.Int32(1),
						Type:     descriptorpb.FieldDescriptorProto_TYPE_MESSAGE.Enum(),
						TypeName: new(".enumtest.Common"),
						Label:    descriptorpb.FieldDescriptorProto_LABEL_OPTIONAL.Enum(),
					},
					{
						Name:     new("status"),
						Number:   proto.Int32(2),
						Type:     descriptorpb.FieldDescriptorProto_TYPE_ENUM.Enum(),
						TypeName: new(".enumtest.Status"),
						Label:    descriptorpb.FieldDescriptorProto_LABEL_OPTIONAL.Enum(),
					},
				},
			},
		},
	}

	req := createTestRequest([]*descriptorpb.FileDescriptorProto{fileDesc}, []string{"enum_test.proto"})

	opts := protogen.Options{} //nolint:exhaustruct_v5

	gen, err := opts.New(req)
	if err != nil {
		t.Fatalf("failed to create plugin: %v", err)
	}

	gen.SupportedFeatures = uint64(pluginpb.CodeGeneratorResponse_FEATURE_PROTO3_OPTIONAL) |
		uint64(pluginpb.CodeGeneratorResponse_FEATURE_SUPPORTS_EDITIONS)
	gen.SupportedEditionsMinimum = descriptorpb.Edition_EDITION_2023
	gen.SupportedEditionsMaximum = descriptorpb.Edition_EDITION_2024

	for _, f := range gen.Files {
		if !f.Generate {
			continue
		}

		generateFile(gen, f)
	}

	resp := gen.Response()

	// Verify at least one file was generated
	if len(resp.GetFile()) == 0 {
		t.Fatal("expected at least one file to be generated")
	}

	// Check the generated content contains the event builder
	for _, f := range resp.GetFile() {
		if f.Content != nil {
			content := f.GetContent()
			if !strings.Contains(content, "NewEventWithEnumEvent") {
				t.Error("generated content should contain 'NewEventWithEnumEvent'")
			}
		}
	}
}

func TestGenerateFile_WithAllScalarTypes(t *testing.T) {
	// Create a file with all scalar types to test goType coverage
	fileDesc := &descriptorpb.FileDescriptorProto{
		Name:    new("scalars.proto"),
		Package: new("scalars"),
		Syntax:  new("proto3"),
		Options: &descriptorpb.FileOptions{
			GoPackage: new("example.com/scalars;scalars"),
		},
		MessageType: []*descriptorpb.DescriptorProto{
			{
				Name: new("Common"),
				Field: []*descriptorpb.FieldDescriptorProto{
					{
						Name:   new("timestamp"),
						Number: proto.Int32(1),
						Type:   descriptorpb.FieldDescriptorProto_TYPE_STRING.Enum(),
						Label:  descriptorpb.FieldDescriptorProto_LABEL_OPTIONAL.Enum(),
					},
				},
			},
			{
				Name: new("EventWithScalars"),
				Field: []*descriptorpb.FieldDescriptorProto{
					{
						Name:     new("common"),
						Number:   proto.Int32(1),
						Type:     descriptorpb.FieldDescriptorProto_TYPE_MESSAGE.Enum(),
						TypeName: new(".scalars.Common"),
						Label:    descriptorpb.FieldDescriptorProto_LABEL_OPTIONAL.Enum(),
					},
					{
						Name:   new("string_field"),
						Number: proto.Int32(2),
						Type:   descriptorpb.FieldDescriptorProto_TYPE_STRING.Enum(),
						Label:  descriptorpb.FieldDescriptorProto_LABEL_OPTIONAL.Enum(),
					},
					{
						Name:   new("bool_field"),
						Number: proto.Int32(3),
						Type:   descriptorpb.FieldDescriptorProto_TYPE_BOOL.Enum(),
						Label:  descriptorpb.FieldDescriptorProto_LABEL_OPTIONAL.Enum(),
					},
					{
						Name:   new("int32_field"),
						Number: proto.Int32(4),
						Type:   descriptorpb.FieldDescriptorProto_TYPE_INT32.Enum(),
						Label:  descriptorpb.FieldDescriptorProto_LABEL_OPTIONAL.Enum(),
					},
					{
						Name:   new("int64_field"),
						Number: proto.Int32(5),
						Type:   descriptorpb.FieldDescriptorProto_TYPE_INT64.Enum(),
						Label:  descriptorpb.FieldDescriptorProto_LABEL_OPTIONAL.Enum(),
					},
					{
						Name:   new("uint32_field"),
						Number: proto.Int32(6),
						Type:   descriptorpb.FieldDescriptorProto_TYPE_UINT32.Enum(),
						Label:  descriptorpb.FieldDescriptorProto_LABEL_OPTIONAL.Enum(),
					},
					{
						Name:   new("uint64_field"),
						Number: proto.Int32(7),
						Type:   descriptorpb.FieldDescriptorProto_TYPE_UINT64.Enum(),
						Label:  descriptorpb.FieldDescriptorProto_LABEL_OPTIONAL.Enum(),
					},
					{
						Name:   new("float_field"),
						Number: proto.Int32(8),
						Type:   descriptorpb.FieldDescriptorProto_TYPE_FLOAT.Enum(),
						Label:  descriptorpb.FieldDescriptorProto_LABEL_OPTIONAL.Enum(),
					},
					{
						Name:   new("double_field"),
						Number: proto.Int32(9),
						Type:   descriptorpb.FieldDescriptorProto_TYPE_DOUBLE.Enum(),
						Label:  descriptorpb.FieldDescriptorProto_LABEL_OPTIONAL.Enum(),
					},
					{
						Name:   new("bytes_field"),
						Number: proto.Int32(10),
						Type:   descriptorpb.FieldDescriptorProto_TYPE_BYTES.Enum(),
						Label:  descriptorpb.FieldDescriptorProto_LABEL_OPTIONAL.Enum(),
					},
				},
			},
		},
	}

	req := createTestRequest([]*descriptorpb.FileDescriptorProto{fileDesc}, []string{"scalars.proto"})

	opts := protogen.Options{} //nolint:exhaustruct_v5

	gen, err := opts.New(req)
	if err != nil {
		t.Fatalf("failed to create plugin: %v", err)
	}

	gen.SupportedFeatures = uint64(pluginpb.CodeGeneratorResponse_FEATURE_PROTO3_OPTIONAL) |
		uint64(pluginpb.CodeGeneratorResponse_FEATURE_SUPPORTS_EDITIONS)
	gen.SupportedEditionsMinimum = descriptorpb.Edition_EDITION_2023
	gen.SupportedEditionsMaximum = descriptorpb.Edition_EDITION_2024

	for _, f := range gen.Files {
		if !f.Generate {
			continue
		}

		generateFile(gen, f)
	}

	resp := gen.Response()

	if len(resp.GetFile()) == 0 {
		t.Fatal("expected at least one file to be generated")
	}

	for _, f := range resp.GetFile() {
		if f.Content != nil {
			content := f.GetContent()
			if !strings.Contains(content, "NewEventWithScalarsEvent") {
				t.Error("generated content should contain 'NewEventWithScalarsEvent'")
			}
		}
	}
}

func TestGoType_WithMessageAndEnumFields(t *testing.T) {
	// Create a file with message and enum fields to test goType coverage
	fileDesc := &descriptorpb.FileDescriptorProto{
		Name:    new("goType_test.proto"),
		Package: new("gotypetest"),
		Syntax:  new("proto3"),
		Options: &descriptorpb.FileOptions{
			GoPackage: new("example.com/gotypetest;gotypetest"),
		},
		EnumType: []*descriptorpb.EnumDescriptorProto{
			{
				Name: new("MyEnum"),
				Value: []*descriptorpb.EnumValueDescriptorProto{
					{Name: new("MY_ENUM_UNSPECIFIED"), Number: proto.Int32(0)},
					{Name: new("MY_ENUM_VALUE"), Number: proto.Int32(1)},
				},
			},
		},
		MessageType: []*descriptorpb.DescriptorProto{
			{
				Name: new("Common"),
				Field: []*descriptorpb.FieldDescriptorProto{
					{
						Name:   new("timestamp"),
						Number: proto.Int32(1),
						Type:   descriptorpb.FieldDescriptorProto_TYPE_STRING.Enum(),
						Label:  descriptorpb.FieldDescriptorProto_LABEL_OPTIONAL.Enum(),
					},
				},
			},
			{
				Name: new("NestedMsg"),
				Field: []*descriptorpb.FieldDescriptorProto{
					{
						Name:   new("value"),
						Number: proto.Int32(1),
						Type:   descriptorpb.FieldDescriptorProto_TYPE_STRING.Enum(),
						Label:  descriptorpb.FieldDescriptorProto_LABEL_OPTIONAL.Enum(),
					},
				},
			},
			{
				Name: new("EventWithMessageAndEnum"),
				Field: []*descriptorpb.FieldDescriptorProto{
					{
						Name:     new("common"),
						Number:   proto.Int32(1),
						Type:     descriptorpb.FieldDescriptorProto_TYPE_MESSAGE.Enum(),
						TypeName: new(".gotypetest.Common"),
						Label:    descriptorpb.FieldDescriptorProto_LABEL_OPTIONAL.Enum(),
					},
					{
						Name:     new("nested"),
						Number:   proto.Int32(2),
						Type:     descriptorpb.FieldDescriptorProto_TYPE_MESSAGE.Enum(),
						TypeName: new(".gotypetest.NestedMsg"),
						Label:    descriptorpb.FieldDescriptorProto_LABEL_OPTIONAL.Enum(),
					},
					{
						Name:     new("my_enum"),
						Number:   proto.Int32(3),
						Type:     descriptorpb.FieldDescriptorProto_TYPE_ENUM.Enum(),
						TypeName: new(".gotypetest.MyEnum"),
						Label:    descriptorpb.FieldDescriptorProto_LABEL_OPTIONAL.Enum(),
					},
				},
			},
		},
	}

	req := createTestRequest([]*descriptorpb.FileDescriptorProto{fileDesc}, []string{"goType_test.proto"})

	opts := protogen.Options{} //nolint:exhaustruct_v5

	gen, err := opts.New(req)
	if err != nil {
		t.Fatalf("failed to create plugin: %v", err)
	}

	gen.SupportedFeatures = uint64(pluginpb.CodeGeneratorResponse_FEATURE_PROTO3_OPTIONAL) |
		uint64(pluginpb.CodeGeneratorResponse_FEATURE_SUPPORTS_EDITIONS)
	gen.SupportedEditionsMinimum = descriptorpb.Edition_EDITION_2023
	gen.SupportedEditionsMaximum = descriptorpb.Edition_EDITION_2024

	// Directly test goType by finding fields
	for _, f := range gen.Files {
		for _, msg := range f.Messages {
			if msg.GoIdent.GoName == "EventWithMessageAndEnum" {
				for _, field := range msg.Fields {
					switch field.GoName {
					case "Nested":
						// Message type
						typ := goType(field)
						if typ != "*NestedMsg" {
							t.Errorf("goType for message field = %q, want %q", typ, "*NestedMsg")
						}
					case "MyEnum":
						// Enum type
						typ := goType(field)
						if typ != "MyEnum" {
							t.Errorf("goType for enum field = %q, want %q", typ, "MyEnum")
						}
					}
				}
			}
		}
	}
}
