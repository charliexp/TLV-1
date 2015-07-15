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

/**
实现TLV数据的解码功能
*/
package golang

import (
	"fmt"
)

/**
TLV网络数据解码器
*/
type StreamDecoder struct {
	buf    []byte //缓冲区
	bufLen int    //缓冲区数据长度

	beforeCursor int //之前解析到数据位置
	curCursor    int //当前解析到数据位置

	isFindTag bool
	isFindLen bool

	valueLen int //数据段的长度

	callBack TLVCallBack
}

type TLVCallBack interface {
	onFinish(tlvObject *TLVObject) (err error)
}

/**
从网络流数据中解析出TLV结构数据
*/
func (this *StreamDecoder) Parse(request []byte, requestLen int) (err error) {
	this.buf = append(this.buf, request[:requestLen]...)
	this.bufLen += requestLen

	for ; this.curCursor < this.bufLen; this.curCursor++ {

		//fmt.Printf("curCursor = %v, bufLen = %d\n", this.curCursor, this.bufLen)

		//计算tag
		if this.isFindTag == false {
			if this.buf[this.curCursor]&0x80 == 0 {
				this.isFindTag = true
				this.beforeCursor = this.curCursor + 1

				//fmt.Printf("findTag-->curCursor = %v\n", this.curCursor)
			}
			continue
		}

		//计算length
		if this.isFindLen == false {
			if this.buf[this.curCursor]&0x80 == 0 {
				this.isFindLen = true
				this.valueLen = parseLength(this.buf[this.beforeCursor : this.curCursor+1])

				this.beforeCursor = this.curCursor

				//fmt.Printf("findLen-->curCursor = %v, valueLen = %v\n", this.curCursor, this.valueLen)
			}
			continue
		}

		if this.curCursor-this.beforeCursor == this.valueLen {
			//已经完整的获取到一个tlv包数据，开始解析整个tlv包
			//fmt.Printf("tlvBytes1 = %v\n", this.buf[:this.curCursor+1])

			tlvObject := TLVObject{}
			tlvObject.FromBytes(this.buf[:this.curCursor+1])
			fmt.Printf("%v\n", tlvObject)

			if this.callBack != nil {
				this.callBack.onFinish(&tlvObject)
			}

			this.reset()
		}
	}

	return nil
}

/**
解析完一个TLV结构后，重置解码器
*/
func (this *StreamDecoder) reset() {
	//遗弃已经解析完的数据包
	this.buf = this.buf[this.curCursor+1:]
	this.bufLen = this.bufLen - this.curCursor - 1

	this.isFindLen = false
	this.isFindTag = false

	this.beforeCursor = 0
	this.curCursor = -1
}

/**
查找tag部分占多少字节
*/
func findTagByteCount(tlvBytes []byte) (tagByteCount int) {
	for i := 0; i < len(tlvBytes); i++ {
		tagByteCount++
		if tlvBytes[i]&0x80 == 0 {
			break
		}
	}

	return tagByteCount
}

/**
查找length部分占多少字节
*/
func findLenByteCount(tlvBytes []byte, lenStartPos int) (lenByteCount int) {
	for i := lenStartPos; i < len(tlvBytes); i++ {
		lenByteCount++
		if tlvBytes[i]&0x80 == 0 {
			break
		}
	}

	return lenByteCount
}

/**
解析数据类型
*/
func parseTag(tagBytes []byte) (frameType byte, dataType byte, tagValue int) {
	frameType = tagBytes[0] & FRAME_TYPE_PRIVATE
	dataType = tagBytes[0] & DATA_TYPE_STRUCT

	tempByte := tagBytes[0] //临时保存，计算完成tagValue恢复

	tagBytes[0] = tagBytes[0] & 0x1f
	tagValue = 0
	power := 1
	byteCount := len(tagBytes)
	for i := 0; i < byteCount; i++ {
		digit := tagBytes[i]
		tagValue += int(digit&0x7f) * power
		power *= 128
	}

	tagBytes[0] = tempByte

	return frameType, dataType, tagValue
}

/**
解析数据长度
*/
func parseLength(lenBytes []byte) (length int) {
	length = 0
	power := 1
	byteCount := len(lenBytes)
	for i := 0; i < byteCount; i++ {
		digit := lenBytes[i]
		length += int(digit&0x7f) * power
		power *= 128
	}

	return length
}
