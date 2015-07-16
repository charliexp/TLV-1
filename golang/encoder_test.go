// Copyright (c) 2014 Dataence, LLC. All rights reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package golang

import (
	"encoding/binary"
	"fmt"
	"strings"
	"testing"
)

func TestTLVPkg(t *testing.T) {

	var int8Value int8 = 10
	int8ValueBytes := make([]byte, 1)
	int8ValueBytes[0] = byte(int8Value)
	int8Field := TLVPkg{
		DataType: DateTypePrimitive,
		TagValue: 0,
		Value:    int8ValueBytes,
	}
	int8Field.Build()

	var int16Value int16 = -300
	int16ValueBytes := make([]byte, 2)
	binary.BigEndian.PutUint16(int16ValueBytes, uint16(int16Value))
	int16Field := TLVPkg{
		DataType: DateTypePrimitive,
		TagValue: 1,
		Value:    int16ValueBytes,
	}
	int16Field.Build()

	var int32Value int32 = 655354
	int32ValueBytes := make([]byte, 4)
	binary.BigEndian.PutUint32(int32ValueBytes, uint32(int32Value))
	int32Field := TLVPkg{
		DataType: DateTypePrimitive,
		TagValue: 2,
		Value:    int32ValueBytes,
	}
	int32Field.Build()

	var int64Value int64 = 65535400
	int64ValueBytes := make([]byte, 8)
	binary.BigEndian.PutUint64(int64ValueBytes, uint64(int64Value))
	int64Field := TLVPkg{
		DataType: DateTypePrimitive,
		TagValue: 3,
		Value:    int64ValueBytes,
	}
	int64Field.Build()

	stringValue := "zhoujunhua"
	stringField := TLVPkg{
		DataType: DateTypePrimitive,
		TagValue: 4,
		Value:    []byte(stringValue),
	}
	stringField.Build()

	var rootValue []byte
	rootValue = append(rootValue, int8Field.Bytes()...)
	rootValue = append(rootValue, int16Field.Bytes()...)
	rootValue = append(rootValue, int32Field.Bytes()...)
	rootValue = append(rootValue, int64Field.Bytes()...)
	rootValue = append(rootValue, stringField.Bytes()...)
	rootPkg := TLVPkg{
		DataType: DataTypeStruct,
		TagValue: 0,
		Value:    rootValue,
	}
	rootPkg.Build()
	tlvBytes := rootPkg.Bytes()

	//数据序列化完成，进行反序列化
	tlvObject := TLVObject{}
	tlvObject.FromBytes(tlvBytes)

	fmt.Printf("%v\n", tlvObject)

	//fmt.Printf("tlvBytes = %v\n", tlvBytes)

	mutiTLVBytes := tlvBytes
	mutiTLVBytes = append(mutiTLVBytes, tlvBytes...)
	mutiTLVBytes = append(mutiTLVBytes, tlvBytes...)
	mutiTLVBytes = append(mutiTLVBytes, tlvBytes...)

	streamDecoder := StreamDecoder{}
	streamDecoder.Parse(mutiTLVBytes[:5], len(mutiTLVBytes[:5]))
	streamDecoder.Parse(mutiTLVBytes[5:6], len(mutiTLVBytes[5:6]))
	streamDecoder.Parse(mutiTLVBytes[6:10], len(mutiTLVBytes[6:10]))
	streamDecoder.Parse(mutiTLVBytes[10:], len(mutiTLVBytes[10:]))

	findObject, ok := tlvObject.Get(0)
	if ok {
		int8ValueParse, ok := findObject.GetInt8(0)
		if ok {
			fmt.Printf("int8ValueParse = %v\n", int8ValueParse)
		} else {
			t.Errorf("没有找到int8Field\n")
		}

		int16ValueParse, ok := findObject.GetInt16(1)
		if ok {
			fmt.Printf("int16ValueParse = %v\n", int16ValueParse)
		} else {
			t.Errorf("没有找到int16Field\n")
		}

		int32ValueParse, ok := findObject.GetInt32(2)
		if ok {
			fmt.Printf("int32ValueParse = %v\n", int32ValueParse)
		} else {
			t.Errorf("没有找到int32Field\n")
		}

		int64ValueParse, ok := findObject.GetInt64(3)
		if ok {
			fmt.Printf("int64ValueParse = %v\n", int64ValueParse)
		} else {
			t.Errorf("没有找到int64Field\n")
		}

		stringParse, ok := findObject.GetString(4)
		if ok {
			fmt.Printf("stringParse = %v\n", stringParse)
		} else {
			t.Errorf("没有找到stringField\n")
		}

		if int8ValueParse != int8Value ||
			int16ValueParse != int16Value ||
			int32ValueParse != int32Value ||
			int64ValueParse != int64Value ||
			strings.EqualFold(stringParse, stringValue) == false {
			t.Errorf("测试失败\n")
		}

	} else {
		t.Errorf("没有找到findObject\n")
	}
}

/**
测试长度编码和解码是否正确
*/
func TestBuildLength(t *testing.T) {
	rawLength := []int{0x00, 0x7f, 0x81, 0x7fff, 0x8001}

	for i := 0; i < len(rawLength); i++ {
		lenBytes := buildLength(rawLength[i])
		parseLength := parseLength(lenBytes)

		if rawLength[i] != parseLength {
			fmt.Errorf("rawLength[%d] = %d, parseLength = %d\n", i, rawLength[i], parseLength)
		}
	}

}

/**
测试类型编码和解码是否正确
*/
func TestBuildTag(t *testing.T) {
	rawFrameType := []byte{FarmeTypePrimitive, FarmeTypePrivate}
	rawDataType := []byte{DateTypePrimitive, DataTypeStruct}
	rawTagValue := []int{0x1f, 0x81, 0x3FFF, 0x3FFFF}

	for i := 0; i < len(rawFrameType); i++ {
		for j := 0; j < len(rawDataType); j++ {
			for k := 0; k < len(rawTagValue); k++ {
				tagBytes := buildTag(rawFrameType[i], rawDataType[j], rawTagValue[k])
				frameType, dataType, tagValue := parseTag(tagBytes)

				if tagValue != rawTagValue[k] || frameType != rawFrameType[i] || dataType != rawDataType[j] {
					fmt.Errorf("rawdata--> rawTagValue=%d, rawFrameType=%d, rawDataType=%d\n", rawTagValue[k], rawFrameType[i], rawDataType[j])
					fmt.Errorf("parseResult--> tagValue=%d, frameType=%d, dataType=%d\n", tagValue, frameType, dataType)
				}
			}
		}
	}

}
