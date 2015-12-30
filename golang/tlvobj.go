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
	"errors"
	"fmt"
	"math"
)

var (
	ErrInvalidParam = errors.New("输入参数非法")
)

// TLV构建对象
type TLVObject struct {
	Pkg TLVPkg

	node []*TLVObject //该tlv结构下的数据
}

// 添加一个TLV对象
func (this *TLVObject) addNode(node *TLVObject) {
	this.node = append(this.node, node)
}

func (this TLVObject) String() (ret string) {
	return traversalField(this.node)
}

// 递归遍历节点数据
func traversalField(node []*TLVObject) (ret string) {
	for i := 0; i < len(node); i++ {
		ret += fmt.Sprintf("%v", node[i].Pkg)
		ret += traversalField(node[i].node)
	}

	return ret
}

// 通过二进制字节，得到TLV对象
func (this *TLVObject) FromBytes(tlvBytes []byte) {
	parseTLVPkg(this, tlvBytes)
}

// 解析出TLV对象
func parseTLVPkg(node *TLVObject, tlvBytes []byte) {

	tagByteCount := findTagByteCount(tlvBytes)
	lenByteCount := findLenByteCount(tlvBytes, tagByteCount)
	length := parseLength(tlvBytes[tagByteCount : tagByteCount+lenByteCount])

	//fmt.Printf("tagByteCount = %v, lenByteCount = %v, length = %v\n", tagByteCount, lenByteCount, length)

	var value []byte
	frameType, dataType, tagValue := parseTag(tlvBytes[:tagByteCount])
	value = tlvBytes[tagByteCount+lenByteCount : tagByteCount+lenByteCount+length]

	//fmt.Printf("frameType = %v, dataType = %v, tagValue = %v, value = %v\n", frameType, dataType, tagValue, value)

	pkg := TLVPkg{
		FrameType: frameType,
		DataType:  dataType,
		TagValue:  tagValue,
		Value:     value,
	}

	newNode := TLVObject{
		Pkg: pkg,
	}
	node.addNode(&newNode)

	if dataType == DataTypeStruct {
		tlvBytes = tlvBytes[tagByteCount+lenByteCount:]
		remainLen := len(tlvBytes)
		offset := 0

		for {
			if remainLen <= 0 {
				//fmt.Printf("exit, remainLen = %v\n", remainLen)
				break
			}

			tagByteCount := findTagByteCount(tlvBytes[offset:])
			lenByteCount := findLenByteCount(tlvBytes[offset:], tagByteCount)
			length := parseLength(tlvBytes[offset+tagByteCount : offset+tagByteCount+lenByteCount])

			consumeLen := tagByteCount + lenByteCount + length

			parseTLVPkg(&newNode, tlvBytes[offset:offset+consumeLen])

			//fmt.Printf("remainLen = %v, consumeLen = %v\n", remainLen, consumeLen)

			offset += consumeLen
			remainLen -= consumeLen
		}

	}
}

func findTLVObject(rawObject *TLVObject, key int) (retObject *TLVObject, ok bool) {
	ok = false
	for i := 0; i < len(rawObject.node); i++ {
		tagValue := rawObject.node[i].Pkg.TagValue
		if tagValue == key {
			retObject = rawObject.node[i]
			ok = true
			break
		}
	}
	return retObject, ok
}

// 获取TLVObject的key
func (this *TLVObject) GetKey() int {
	return this.node[0].Pkg.TagValue
}

// 获取TLVObject下的一个TLVObject
func (this *TLVObject) Get(key int) (tlvObject *TLVObject, ok bool) {
	findObject, ok := findTLVObject(this, key)
	return findObject, ok
}

func (this *TLVObject) GetBool(key int) (ret bool, ok bool) {
	findObject, ok := findTLVObject(this, key)
	if ok == false {
		//fmt.Printf("don't exist this field, key:%v\n", key)
		return false, false
	}

	value := findObject.Pkg.Value
	if len(value) != 1 {
		fmt.Printf("该字段不是bool类型, key:%v, value:%v\n", key, value)
		return false, false
	}

	if value[0]&0x01 > 0 {
		ret = true
	} else {
		ret = false
	}

	return ret, false
}

func (this *TLVObject) GetInt8(key int) (ret int8, ok bool) {
	findObject, ok := findTLVObject(this, key)
	if ok == false {
		//fmt.Printf("don't exist this field, key:%v\n", key)
		return 0, false
	}

	value := findObject.Pkg.Value
	if len(value) != 1 {
		fmt.Printf("该字段不是int8类型, key:%v, value:%v\n", key, value)
		return 0, false
	}

	return int8(value[0]), true
}

func (this *TLVObject) GetUint8(key int) (ret uint8, ok bool) {
	findObject, ok := findTLVObject(this, key)
	if ok == false {
		//fmt.Printf("don't exist this field, key:%v\n", key)
		return 0, false
	}

	value := findObject.Pkg.Value
	if len(value) != 1 {
		fmt.Printf("该字段不是int8类型, key:%v, value:%v\n", key, value)
		return 0, false
	}

	return uint8(value[0]), true
}

// 获取指定位数
func (this *TLVObject) getIntWithDigit(key int, digit int) (ret int64, ok bool) {
	findObject, ok := findTLVObject(this, key)
	if ok == false {
		//fmt.Printf("don't exist this field, key:%v\n", key)
		return 0, false
	}

	value := findObject.Pkg.Value
	if len(value) != digit {
		fmt.Printf("该字段不是int8类型, key:%v, value:%v\n", key, value)
		return 0, false
	}

	switch digit {
	case 2:
		ret = int64(binary.BigEndian.Uint16(value))
	case 4:
		ret = int64(binary.BigEndian.Uint32(value))
	case 8:
		ret = int64(binary.BigEndian.Uint64(value))
	default:
		ok = false
		fmt.Printf("int number of bytes invalid: %v\n", digit)
	}

	return ret, ok
}

func (this *TLVObject) GetInt16(key int) (int16, bool) {
	ret, ok := this.getIntWithDigit(key, 2)
	return int16(ret), ok
}

func (this *TLVObject) GetInt32(key int) (int32, bool) {
	ret, ok := this.getIntWithDigit(key, 4)
	return int32(ret), ok
}

func (this *TLVObject) GetInt64(key int) (int64, bool) {
	ret, ok := this.getIntWithDigit(key, 8)
	return ret, ok
}

func (this *TLVObject) GetUint16(key int) (uint16, bool) {
	ret, ok := this.getIntWithDigit(key, 2)
	return uint16(ret), ok
}

func (this *TLVObject) GetUint32(key int) (uint32, bool) {
	ret, ok := this.getIntWithDigit(key, 4)
	return uint32(ret), ok
}

func (this *TLVObject) GetUint64(key int) (uint64, bool) {
	ret, ok := this.getIntWithDigit(key, 8)
	return uint64(ret), ok
}

func (this *TLVObject) GetVarUint(key int) (ret uint64, ok bool) {
	findObject, ok := findTLVObject(this, key)
	if ok == false {
		//fmt.Printf("don't exist this field, key:%v\n", key)
		return 0, false
	}

	value := findObject.Pkg.Value
	digit := len(value)

	switch digit {
	case 1:
		ret = uint64(value[0])
	case 2:
		ret = uint64(binary.BigEndian.Uint16(value))
	case 4:
		ret = uint64(binary.BigEndian.Uint32(value))
	case 8:
		ret = uint64(binary.BigEndian.Uint64(value))
	default:
		ok = false
		fmt.Printf("int number of bytes invalid: %v\n", digit)
	}

	return ret, ok
}

func (this *TLVObject) GetVarInt(key int) (ret int64, ok bool) {
	findObject, ok := findTLVObject(this, key)
	if ok == false {
		//fmt.Printf("don't exist this field, key:%v\n", key)
		return 0, false
	}

	value := findObject.Pkg.Value
	digit := len(value)

	switch digit {
	case 1:
		ret = int64(value[0])
	case 2:
		ret = int64(binary.BigEndian.Uint16(value))
	case 4:
		ret = int64(binary.BigEndian.Uint32(value))
	case 8:
		ret = int64(binary.BigEndian.Uint64(value))
	default:
		ok = false
		fmt.Printf("int number of bytes invalid: %v\n", digit)
	}

	return ret, ok
}

func (this *TLVObject) GetBytes(key int) ([]byte, bool) {
	findObject, ok := findTLVObject(this, key)
	if ok == false {
		//fmt.Printf("don't exist this field, key:%v\n", key)
		return nil, false
	}

	return findObject.Pkg.Value, true
}

func (this *TLVObject) GetString(key int) (ret string, ok bool) {
	findObject, ok := findTLVObject(this, key)
	if ok == false {
		//fmt.Printf("don't exist this field, key:%v\n", key)
		return "", false
	}
	ret = string(findObject.Pkg.Value)
	return ret, true
}

func (this *TLVObject) Put(key int, tlvObject *TLVObject) error {
	tlvObject.Pkg.FrameType = FarmeTypePrimitive
	tlvObject.Pkg.DataType = DataTypeStruct
	tlvObject.Pkg.TagValue = key
	this.addNode(tlvObject)
	return nil
}

// 添加基本数据节点
func (this *TLVObject) addPrimitiveNode(key int, valueBytes []byte) {
	pkg := TLVPkg{
		FrameType: FarmeTypePrimitive,
		DataType:  DataTypePrimitive,
		TagValue:  key,
		Value:     valueBytes,
	}
	pkg.Build()

	newNode := TLVObject{
		Pkg: pkg,
	}
	this.addNode(&newNode)
}

func (this *TLVObject) PutBool(key int, value bool) error {
	valueBytes := []byte{0}
	if value {
		valueBytes[0] = 1
	}

	this.addPrimitiveNode(key, valueBytes)
	return nil
}

func (this *TLVObject) PutInt8(key int, value int8) error {
	return this.PutUint8(key, uint8(value))
}

func (this *TLVObject) PutUint8(key int, value uint8) error {
	valueBytes := []byte{byte(value)}

	this.addPrimitiveNode(key, valueBytes)
	return nil
}

func (this *TLVObject) PutInt16(key int, value int16) error {
	return this.PutUint16(key, uint16(value))
}

func (this *TLVObject) PutUint16(key int, value uint16) error {
	valueBytes := make([]byte, 2)
	binary.BigEndian.PutUint16(valueBytes, value)

	this.addPrimitiveNode(key, valueBytes)
	return nil
}

func (this *TLVObject) PutInt32(key int, value int32) error {
	return this.PutUint32(key, uint32(value))
}

func (this *TLVObject) PutUint32(key int, value uint32) error {
	valueBytes := make([]byte, 4)
	binary.BigEndian.PutUint32(valueBytes, value)

	this.addPrimitiveNode(key, valueBytes)
	return nil
}

func (this *TLVObject) PutInt64(key int, value int64) error {
	return this.PutUint64(key, uint64(value))
}

// 写入任意长度的整形数据
// 范围为int8到int64之间，根据数值大小，自动计算
func (this *TLVObject) PutVarInt(key int, value int64) (err error) {
	if value >= math.MinInt8 && value <= math.MaxInt8 {
		err = this.PutInt8(key, int8(value))
	} else if value >= math.MinInt16 && value <= math.MaxInt16 {
		err = this.PutInt16(key, int16(value))
	} else if value >= math.MinInt32 && value <= math.MaxInt32 {
		err = this.PutInt32(key, int32(value))
	} else {
		err = this.PutInt64(key, int64(value))
	}

	return err
}

func (this *TLVObject) PutVarUint(key int, value uint64) (err error) {
	if value >= 0 && value <= math.MaxUint8 {
		err = this.PutUint8(key, uint8(value))
	} else if value > math.MaxUint8 && value <= math.MaxUint16 {
		err = this.PutUint16(key, uint16(value))
	} else if value > math.MaxUint16 && value <= math.MaxUint32 {
		err = this.PutUint32(key, uint32(value))
	} else {
		err = this.PutUint64(key, uint64(value))
	}

	return err
}

func (this *TLVObject) PutUint64(key int, value uint64) error {
	valueBytes := make([]byte, 8)
	binary.BigEndian.PutUint64(valueBytes, value)

	this.addPrimitiveNode(key, valueBytes)
	return nil
}

func (this *TLVObject) PutBytes(key int, value []byte) error {
	this.addPrimitiveNode(key, value)
	return nil
}

func (this *TLVObject) PutString(key int, value string) error {
	if value == "" {
		return ErrInvalidParam
	}

	valueBytes := []byte(value)
	this.addPrimitiveNode(key, valueBytes)
	return nil
}

// 构建TLV嵌套结构的节点数据
func buildNode(node []*TLVObject) (nodeBytes []byte) {
	//fmt.Printf("node count:%v\n", len(node))
	for i := 0; i < len(node); i++ {
		if node[i].Pkg.DataType == DataTypeStruct {
			node[i].Pkg.Value = buildNode(node[i].node)
			node[i].Pkg.Build()
		}
		//fmt.Printf("append pkg:%v", node[i].Pkg)
		nodeBytes = append(nodeBytes, node[i].Pkg.Bytes()...)
		//fmt.Printf("nodeBytes:%v\n\n", nodeBytes)
	}

	return nodeBytes
}

// 获取TLV的字节数据
func (this *TLVObject) Bytes() []byte {
	if this.Pkg.Value == nil {
		this.Pkg.Value = buildNode(this.node)
	}
	return this.Pkg.Value
}
