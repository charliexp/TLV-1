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
	// "bytes"
	// "encoding/binary"
	"fmt"
	"testing"
)

func TestTLVPkg(t *testing.T) {

	command := TLVPkg{
		DataType: DATA_TYPE_PRIMITVIE,
		TagValue: 0,
		Value:    []byte{10},
	}
	command.Build()

	sid := TLVPkg{
		DataType: DATA_TYPE_PRIMITVIE,
		TagValue: 2,
		Value:    []byte{20},
	}
	sid.Build()

	var rootValue []byte
	rootValue = append(rootValue, command.getBytes()...)
	rootValue = append(rootValue, sid.getBytes()...)
	rootPkg := TLVPkg{
		DataType: DATA_TYPE_STRUCT,
		TagValue: 3,
		Value:    rootValue,
	}
	rootPkg.Build()
	tlvBytes := rootPkg.getBytes()

	//fmt.Printf("tlvBytes = %v\n", tlvBytes)

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
	rawFrameType := []byte{FRAME_TYPE_PRIMITVIE, FRAME_TYPE_PRIVATE}
	rawDataType := []byte{DATA_TYPE_PRIMITVIE, DATA_TYPE_STRUCT}
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
