package formats

import (
	"testing"

	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/descriptorpb"
)

func testPersonDescriptorSetBytes(t *testing.T) string {
	t.Helper()

	optional := descriptorpb.FieldDescriptorProto_LABEL_OPTIONAL
	repeated := descriptorpb.FieldDescriptorProto_LABEL_REPEATED
	stringType := descriptorpb.FieldDescriptorProto_TYPE_STRING
	int32Type := descriptorpb.FieldDescriptorProto_TYPE_INT32

	descriptorSet := &descriptorpb.FileDescriptorSet{
		File: []*descriptorpb.FileDescriptorProto{
			{
				Name:    proto.String("person.proto"),
				Package: proto.String("test"),
				Syntax:  proto.String("proto3"),
				MessageType: []*descriptorpb.DescriptorProto{
					{
						Name: proto.String("Person"),
						Field: []*descriptorpb.FieldDescriptorProto{
							{
								Name:   proto.String("name"),
								Number: proto.Int32(1),
								Label:  &optional,
								Type:   &stringType,
							},
							{
								Name:     proto.String("lucky_numbers"),
								Number:   proto.Int32(2),
								Label:    &repeated,
								Type:     &int32Type,
								JsonName: proto.String("luckyNumbers"),
								Options: &descriptorpb.FieldOptions{
									Packed: proto.Bool(true),
								},
							},
						},
					},
				},
			},
		},
	}

	encoded, err := proto.Marshal(descriptorSet)
	if err != nil {
		t.Fatalf("failed to marshal descriptor set: %v", err)
	}
	return string(encoded)
}

func testMapDescriptorSetBytes(t *testing.T) string {
	t.Helper()

	optional := descriptorpb.FieldDescriptorProto_LABEL_OPTIONAL
	repeated := descriptorpb.FieldDescriptorProto_LABEL_REPEATED
	messageType := descriptorpb.FieldDescriptorProto_TYPE_MESSAGE
	int32Type := descriptorpb.FieldDescriptorProto_TYPE_INT32
	mapEntry := true

	descriptorSet := &descriptorpb.FileDescriptorSet{
		File: []*descriptorpb.FileDescriptorProto{
			{
				Name:    proto.String("m.proto"),
				Package: proto.String("t"),
				Syntax:  proto.String("proto3"),
				MessageType: []*descriptorpb.DescriptorProto{
					{
						Name: proto.String("M"),
						Field: []*descriptorpb.FieldDescriptorProto{
							{
								Name:     proto.String("kv"),
								Number:   proto.Int32(1),
								Label:    &repeated,
								Type:     &messageType,
								TypeName: proto.String(".t.M.KvEntry"),
							},
						},
						NestedType: []*descriptorpb.DescriptorProto{
							{
								Name:    proto.String("KvEntry"),
								Options: &descriptorpb.MessageOptions{MapEntry: &mapEntry},
								Field: []*descriptorpb.FieldDescriptorProto{
									{
										Name:   proto.String("key"),
										Number: proto.Int32(1),
										Label:  &optional,
										Type:   &int32Type,
									},
									{
										Name:   proto.String("value"),
										Number: proto.Int32(2),
										Label:  &optional,
										Type:   &int32Type,
									},
								},
							},
						},
					},
				},
			},
		},
	}

	encoded, err := proto.Marshal(descriptorSet)
	if err != nil {
		t.Fatalf("failed to marshal map descriptor set: %v", err)
	}
	return string(encoded)
}
