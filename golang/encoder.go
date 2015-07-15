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

/*
实现TLV编码功能
*/
package golang

import (
	"fmt"
)

// 帧类型
const (
	FRAME_TYPE_PRIMITVIE = 0x00 //基本类型
	FRAME_TYPE_PRIVATE   = 0x40 //私有类型
)

// 数据类型
const (
	DATA_TYPE_PRIMITVIE = 0x00 //基本数据编码
	DATA_TYPE_STRUCT    = 0x20 //TLV嵌套
)

type TLVPkg struct {
	tagByteCount  int //类型字段占用的字节数
	lenByteCount  int //长度字段占用的字节数
	dataByteCount int //数据字段暂用的字节数

	FrameType byte   //帧类型，0-基本类型，1-私有类型
	DataType  byte   //数据类型，0-基本数据，1-TLV数据
	TagValue  int    //tag类型值
	Value     []byte //实际数据

	data []byte //数据包字节数据
}

/**
构建tlv对象数据
*/
func (this *TLVPkg) Build() {
	this.dataByteCount = len(this.Value)

	tagBytes := buildTag(this.FrameType, this.DataType, this.TagValue)
	lenBytes := buildLength(this.dataByteCount)

	this.tagByteCount = len(tagBytes)
	this.lenByteCount = len(lenBytes)

	this.data = append(this.data, tagBytes...)
	this.data = append(this.data, lenBytes...)
	this.data = append(this.data, this.Value...)
}

/**
获取TLV数据包大小
*/
func (this *TLVPkg) Size() int {
	return this.tagByteCount + this.lenByteCount + this.dataByteCount
}

/**
获取TLV数据包的字节数据
*/
func (this *TLVPkg) getBytes() []byte {
	return this.data
}

func (this TLVPkg) String() (ret string) {
	ret = fmt.Sprintf("FrameType = %d, DataType = %d, TagValue = %d, Vaule = %v\n", this.FrameType, this.DataType, this.TagValue, this.Value)
	return ret
}

/**
生成TLV的Tag字节数据
*/
func buildTag(frameType byte, dataType byte, tagValue int) (tagBytes []byte) {

	tagValueBytes := buildLength(tagValue)
	count := len(tagValueBytes)

	if tagValueBytes[count-1] > 0x1f {
		tagBytes = append(tagBytes, 0x80)
	}
	tagBytes = append(tagBytes, tagValueBytes...)

	tagBytes[0] = tagBytes[0] | frameType | dataType

	return tagBytes
}

/**
生成TLV的数据长度字节数据
*/
func buildLength(length int) (lenBytes []byte) {

	if length < 0 {
		fmt.Errorf("长度不能为负数 length = %d\n", length)
	}

	if length == 0 {
		return []byte{0}
	}

	for {
		if length <= 0 {
			break
		}

		digit := length % 128
		length = length / 128
		if length > 0 {
			digit = digit | 0x80
		}

		lenBytes = append(lenBytes, byte(digit))
	}
	return lenBytes
}
