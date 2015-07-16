package golang

import (
	"encoding/binary"
	"fmt"
)

/**
TLV构建对象
*/
type TLVObject struct {
	Pkg TLVPkg

	node []*TLVObject //该tlv结构下的数据
}

/**
添加一个TLV对象
*/
func (this *TLVObject) addNode(node *TLVObject) {
	this.node = append(this.node, node)
}

func (this TLVObject) String() (ret string) {
	return traversalField(this.node)
}

/**
递归遍历节点数据
*/
func traversalField(node []*TLVObject) (ret string) {
	for i := 0; i < len(node); i++ {
		ret += fmt.Sprintf("%v", node[i].Pkg)
		ret += traversalField(node[i].node)
	}

	return ret
}

/**
通过二进制字节，得到TLV对象
*/
func (this *TLVObject) FromBytes(tlvBytes []byte) {
	parseTLVPkg(this, tlvBytes)
}

/**
解析出TLV对象
*/
func parseTLVPkg(node *TLVObject, tlvBytes []byte) {

	tagByteCount := findTagByteCount(tlvBytes)
	lenByteCount := findLenByteCount(tlvBytes, tagByteCount)
	length := parseLength(tlvBytes[tagByteCount : tagByteCount+lenByteCount])

	fmt.Printf("tagByteCount = %v, lenByteCount = %v, length = %v\n", tagByteCount, lenByteCount, length)

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
			tagByteCount := findTagByteCount(tlvBytes[offset:])
			lenByteCount := findLenByteCount(tlvBytes[offset:], tagByteCount)
			length := parseLength(tlvBytes[offset+tagByteCount : offset+tagByteCount+lenByteCount])

			consumeLen := tagByteCount + lenByteCount + length

			parseTLVPkg(&newNode, tlvBytes[offset:offset+consumeLen])

			//fmt.Printf("remainLen = %v, consumeLen = %v\n", remainLen, consumeLen)

			offset += consumeLen
			remainLen -= consumeLen
			if remainLen <= 0 {
				break
			}
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

/**
获取TLVObject下的一个TLVObject
*/
func (this *TLVObject) Get(key int) (tlvObject *TLVObject, ok bool) {
	findObject, ok := findTLVObject(this, key)
	return findObject, ok
}

/**
获取一个
*/
func (this *TLVObject) GetBool(key int) (ret bool, ok bool) {
	findObject, ok := findTLVObject(this, key)
	if ok == false {
		fmt.Errorf("不存在该字段, key:%v\n", key)
		return false, false
	}

	value := findObject.Pkg.Value
	if len(value) != 1 {
		fmt.Errorf("该字段不是bool类型, key:%v, value:%v\n", key, value)
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
		fmt.Errorf("不存在该字段, key:%v\n", key)
		return 0, false
	}

	value := findObject.Pkg.Value
	if len(value) != 1 {
		fmt.Errorf("该字段不是int8类型, key:%v, value:%v\n", key, value)
		return 0, false
	}

	return int8(value[0]), true
}

func (this *TLVObject) GetUint8(key int) (ret uint8, ok bool) {
	findObject, ok := findTLVObject(this, key)
	if ok == false {
		fmt.Errorf("不存在该字段, key:%v\n", key)
		return 0, false
	}

	value := findObject.Pkg.Value
	if len(value) != 1 {
		fmt.Errorf("该字段不是int8类型, key:%v, value:%v\n", key, value)
		return 0, false
	}

	return uint8(value[0]), true
}

/**
获取指定位数
*/
func (this *TLVObject) getIntWithDigit(key int, digit int) (ret int64, ok bool) {
	findObject, ok := findTLVObject(this, key)
	if ok == false {
		fmt.Errorf("不存在该字段, key:%v\n", key)
		return 0, false
	}

	value := findObject.Pkg.Value
	if len(value) != digit {
		fmt.Errorf("该字段不是int8类型, key:%v, value:%v\n", key, value)
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
		fmt.Errorf("digit不合法, digit:%v\n", digit)
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

func (this *TLVObject) GetString(key int) (ret string, ok bool) {
	findObject, ok := findTLVObject(this, key)
	if ok == false {
		fmt.Errorf("不存在该字段, key:%v\n", key)
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

/**
添加基本数据节点
*/
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
	valueBytes := []byte{byte(value)}

	this.addPrimitiveNode(key, valueBytes)
	return nil
}

func (this *TLVObject) PutInt16(key int, value int16) error {
	valueBytes := make([]byte, 2)
	binary.BigEndian.PutUint16(valueBytes, uint16(value))

	this.addPrimitiveNode(key, valueBytes)
	return nil
}

func (this *TLVObject) PutInt32(key int, value int32) error {
	valueBytes := make([]byte, 4)
	binary.BigEndian.PutUint32(valueBytes, uint32(value))

	this.addPrimitiveNode(key, valueBytes)
	return nil
}

func (this *TLVObject) PutInt64(key int, value int64) error {
	valueBytes := make([]byte, 8)
	binary.BigEndian.PutUint64(valueBytes, uint64(value))

	this.addPrimitiveNode(key, valueBytes)
	return nil
}

func (this *TLVObject) PutString(key int, value string) error {
	valueBytes := []byte(value)

	this.addPrimitiveNode(key, valueBytes)
	return nil
}

func (this *TLVObject) build() {
	this.Pkg.Value = buildNode(this.node)
}

/**
构建TLV嵌套结构的节点数据
*/
func buildNode(node []*TLVObject) (nodeBytes []byte) {
	for i := 0; i < len(node); i++ {
		if node[i].Pkg.DataType == DataTypeStruct {
			node[i].Pkg.Value = buildNode(node[i].node)
			node[i].Pkg.Build()
		}
		nodeBytes = append(nodeBytes, node[i].Pkg.Bytes()...)
	}

	return nodeBytes
}

/**
获取TLV的字节数据
*/
func (this *TLVObject) Bytes() []byte {
	return this.Pkg.Value
}
