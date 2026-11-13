// SPDX-FileCopyrightText: 2025 SAP SE or an SAP affiliate company and OpenKCM contributors
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"strings"
	"testing"

	"google.golang.org/protobuf/compiler/protogen"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/descriptorpb"
	"google.golang.org/protobuf/types/pluginpb"
)

func TestToSnakeCase(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"", ""},
		{"a", "a"},
		{"A", "a"},
		{"DataSubjectType", "data_subject_type"},
		{"ID", "i_d"},
		{"UserId", "user_id"},
		{"XMLParser", "x_m_l_parser"},
		{"simpleCase", "simple_case"},
		{"simple", "simple"},
		{"ALLCAPS", "a_l_l_c_a_p_s"},
		{"TenantID", "tenant_i_d"},
		{"CMKId", "c_m_k_id"},
		{"firstName", "first_name"},
		{"FirstName", "first_name"},
		{"X", "x"},
		{"ABC", "a_b_c"},
		{"camelCase", "camel_case"},
		{"PascalCase", "pascal_case"},
		{"snake_case", "snake_case"},
		{"MixedCASE", "mixed_c_a_s_e"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got := toSnakeCase(tt.input)
			if got != tt.want {
				t.Errorf("toSnakeCase(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

func TestWarnEvents(t *testing.T) {
	// Verify warnEvents map contains expected entries
	expectedWarnEvents := []string{
		"LoginFailed",
		"CMKUnavailable",
		"CMKDisabled",
		"UnauthenticatedRequest",
		"UnauthorizedRequest",
	}

	for _, event := range expectedWarnEvents {
		if !warnEvents[event] {
			t.Errorf("warnEvents should contain %q", event)
		}
	}

	// Verify some events that should NOT be warnings
	nonWarnEvents := []string{
		"LoginSucceeded",
		"TenantCreated",
		"CMKCreated",
		"CMKEnabled",
		"AccessGranted",
	}

	for _, event := range nonWarnEvents {
		if warnEvents[event] {
			t.Errorf("warnEvents should NOT contain %q", event)
		}
	}
}

func TestWarnEventsCount(t *testing.T) {
	// Ensure we have exactly the expected number of warn events
	expectedCount := 5
	if len(warnEvents) != expectedCount {
		t.Errorf("warnEvents has %d entries, want %d", len(warnEvents), expectedCount)
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
		Syntax:  new("proto3"),
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

	// Verify no files were generated (no Common message)
	if len(resp.GetFile()) != 0 {
		t.Errorf("expected no files generated for proto without Common, got %d", len(resp.GetFile()))
	}
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
					{
						Name:   new("description"),
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

			if !strings.Contains(content, "ToLogRecord") {
				t.Error("generated content should contain 'ToLogRecord'")
			}

			if !strings.Contains(content, "FromLogRecord") {
				t.Error("generated content should contain 'FromLogRecord'")
			}

			if !strings.Contains(content, "ParseAuditEvent") {
				t.Error("generated content should contain 'ParseAuditEvent'")
			}

			if !strings.Contains(content, "AuditEvent") {
				t.Error("generated content should contain 'AuditEvent'")
			}

			if !strings.Contains(content, "parseTraceID") {
				t.Error("generated content should contain 'parseTraceID'")
			}

			if !strings.Contains(content, "parseSpanID") {
				t.Error("generated content should contain 'parseSpanID'")
			}

			if !strings.Contains(content, "formatTraceID") {
				t.Error("generated content should contain 'formatTraceID'")
			}

			if !strings.Contains(content, "formatSpanID") {
				t.Error("generated content should contain 'formatSpanID'")
			}

			if !strings.Contains(content, "Code generated by protoc-gen-go-otlpconverters") {
				t.Error("generated content should contain generator comment")
			}

			// Check imports
			if !strings.Contains(content, `"encoding/hex"`) {
				t.Error("generated content should contain hex import")
			}

			if !strings.Contains(content, `"go.opentelemetry.io/collector/pdata/plog"`) {
				t.Error("generated content should contain plog import")
			}
		}
	}

	if !foundContent {
		t.Error("no content was generated")
	}
}

func TestGenerateFile_WithWarnEvent(t *testing.T) {
	// Create a file with an event that should have WARN severity
	fileDesc := &descriptorpb.FileDescriptorProto{
		Name:    new("warn_events.proto"),
		Package: new("warnevents"),
		Syntax:  new("proto3"),
		Options: &descriptorpb.FileOptions{
			GoPackage: new("example.com/warnevents;warnevents"),
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
				Name: new("LoginFailed"), // Should be WARN
				Field: []*descriptorpb.FieldDescriptorProto{
					{
						Name:     new("common"),
						Number:   proto.Int32(1),
						Type:     descriptorpb.FieldDescriptorProto_TYPE_MESSAGE.Enum(),
						TypeName: new(".warnevents.Common"),
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

	req := createTestRequest([]*descriptorpb.FileDescriptorProto{fileDesc}, []string{"warn_events.proto"})

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
			// Verify WARN severity is used for LoginFailed
			if !strings.Contains(content, "SeverityNumberWarn") {
				t.Error("generated content should contain 'SeverityNumberWarn' for LoginFailed event")
			}

			if !strings.Contains(content, `"WARN"`) {
				t.Error("generated content should contain 'WARN' severity text")
			}
		}
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

func TestGenerateFile_WithEnumField(t *testing.T) {
	// Create a file with an enum field to test enum parsing generation
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

	if len(resp.GetFile()) == 0 {
		t.Fatal("expected at least one file to be generated")
	}

	for _, f := range resp.GetFile() {
		if f.Content != nil {
			content := f.GetContent()
			// Verify enum parser is generated
			if !strings.Contains(content, "parseStatus") {
				t.Error("generated content should contain 'parseStatus' for enum parsing")
			}

			// Verify enum is used in ToLogRecord
			if !strings.Contains(content, "String()") {
				t.Error("generated content should contain '.String()' for enum conversion")
			}
		}
	}
}

func TestGenerateFile_WithAllScalarTypes(t *testing.T) {
	// Create a file with all scalar types to test type handling
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
			// Verify different attribute types are used
			if !strings.Contains(content, "PutStr") {
				t.Error("generated content should contain 'PutStr' for string fields")
			}

			if !strings.Contains(content, "PutBool") {
				t.Error("generated content should contain 'PutBool' for bool fields")
			}

			if !strings.Contains(content, "PutInt") {
				t.Error("generated content should contain 'PutInt' for int fields")
			}

			if !strings.Contains(content, "PutDouble") {
				t.Error("generated content should contain 'PutDouble' for float/double fields")
			}

			if !strings.Contains(content, "hex.EncodeToString") {
				t.Error("generated content should contain 'hex.EncodeToString' for bytes fields")
			}

			// Verify FromLogRecord type conversions
			if !strings.Contains(content, "v.Str()") {
				t.Error("generated content should contain 'v.Str()' for string parsing")
			}

			if !strings.Contains(content, "v.Bool()") {
				t.Error("generated content should contain 'v.Bool()' for bool parsing")
			}

			if !strings.Contains(content, "v.Int()") {
				t.Error("generated content should contain 'v.Int()' for int parsing")
			}

			if !strings.Contains(content, "v.Double()") {
				t.Error("generated content should contain 'v.Double()' for float/double parsing")
			}

			if !strings.Contains(content, "hex.DecodeString") {
				t.Error("generated content should contain 'hex.DecodeString' for bytes parsing")
			}
		}
	}
}

func TestGenerateFile_WithNestedMessage(t *testing.T) {
	// Create a file with a nested message field (should generate TODO comment)
	fileDesc := &descriptorpb.FileDescriptorProto{
		Name:    new("nested.proto"),
		Package: new("nested"),
		Syntax:  new("proto3"),
		Options: &descriptorpb.FileOptions{
			GoPackage: new("example.com/nested;nested"),
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
				Name: new("NestedData"),
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
				Name: new("EventWithNestedMessage"),
				Field: []*descriptorpb.FieldDescriptorProto{
					{
						Name:     new("common"),
						Number:   proto.Int32(1),
						Type:     descriptorpb.FieldDescriptorProto_TYPE_MESSAGE.Enum(),
						TypeName: new(".nested.Common"),
						Label:    descriptorpb.FieldDescriptorProto_LABEL_OPTIONAL.Enum(),
					},
					{
						Name:     new("nested_data"),
						Number:   proto.Int32(2),
						Type:     descriptorpb.FieldDescriptorProto_TYPE_MESSAGE.Enum(),
						TypeName: new(".nested.NestedData"),
						Label:    descriptorpb.FieldDescriptorProto_LABEL_OPTIONAL.Enum(),
					},
				},
			},
		},
	}

	req := createTestRequest([]*descriptorpb.FileDescriptorProto{fileDesc}, []string{"nested.proto"})

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
			// Verify TODO comment for nested message
			if !strings.Contains(content, "TODO: handle nested message field") {
				t.Error("generated content should contain TODO comment for nested message")
			}
		}
	}
}

func TestGenerateFile_MultipleEvents(t *testing.T) {
	// Create a file with multiple events to test ParseAuditEvent switch
	fileDesc := &descriptorpb.FileDescriptorProto{
		Name:    new("multi.proto"),
		Package: new("multi"),
		Syntax:  new("proto3"),
		Options: &descriptorpb.FileOptions{
			GoPackage: new("example.com/multi;multi"),
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
				Name: new("EventA"),
				Field: []*descriptorpb.FieldDescriptorProto{
					{
						Name:     new("common"),
						Number:   proto.Int32(1),
						Type:     descriptorpb.FieldDescriptorProto_TYPE_MESSAGE.Enum(),
						TypeName: new(".multi.Common"),
						Label:    descriptorpb.FieldDescriptorProto_LABEL_OPTIONAL.Enum(),
					},
				},
			},
			{
				Name: new("EventB"),
				Field: []*descriptorpb.FieldDescriptorProto{
					{
						Name:     new("common"),
						Number:   proto.Int32(1),
						Type:     descriptorpb.FieldDescriptorProto_TYPE_MESSAGE.Enum(),
						TypeName: new(".multi.Common"),
						Label:    descriptorpb.FieldDescriptorProto_LABEL_OPTIONAL.Enum(),
					},
				},
			},
			{
				Name: new("EventC"),
				Field: []*descriptorpb.FieldDescriptorProto{
					{
						Name:     new("common"),
						Number:   proto.Int32(1),
						Type:     descriptorpb.FieldDescriptorProto_TYPE_MESSAGE.Enum(),
						TypeName: new(".multi.Common"),
						Label:    descriptorpb.FieldDescriptorProto_LABEL_OPTIONAL.Enum(),
					},
				},
			},
		},
	}

	req := createTestRequest([]*descriptorpb.FileDescriptorProto{fileDesc}, []string{"multi.proto"})

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
			// Verify ParseAuditEvent switch contains all events
			if !strings.Contains(content, `case "EventA":`) {
				t.Error("generated content should contain 'case \"EventA\":' in switch")
			}

			if !strings.Contains(content, `case "EventB":`) {
				t.Error("generated content should contain 'case \"EventB\":' in switch")
			}

			if !strings.Contains(content, `case "EventC":`) {
				t.Error("generated content should contain 'case \"EventC\":' in switch")
			}

			// Verify default case
			if !strings.Contains(content, "unknown audit event type") {
				t.Error("generated content should contain 'unknown audit event type' in default case")
			}
		}
	}
}
