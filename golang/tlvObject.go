package golang

import (
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
func (this *TLVObject) AddNode(node *TLVObject) {
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
	node.AddNode(&newNode)

	if dataType == DATA_TYPE_STRUCT {
		tlvBytes = tlvBytes[tagByteCount+lenByteCount:]
		remainLen := len(tlvBytes)
		offset := 0

		for {
			tagByteCount := findTagByteCount(tlvBytes)
			lenByteCount := findLenByteCount(tlvBytes, tagByteCount)
			length := parseLength(tlvBytes[tagByteCount : tagByteCount+lenByteCount])

			consumeLen := tagByteCount + lenByteCount + length

			parseTLVPkg(&newNode, tlvBytes[offset:offset+consumeLen])

			offset += consumeLen
			remainLen -= consumeLen
			if remainLen <= 0 {
				break
			}
		}

	}
}

func (this *TLVObject) Get(key int) *TLVObject {
	return nil
}

func (this *TLVObject) GetBool(key int) bool {
	return false
}

func (this *TLVObject) GetInt8(key int) int8 {
	return 0
}

func (this *TLVObject) GetInt16(key int) int16 {
	return 0
}

func (this *TLVObject) GetInt32(key int) int32 {
	return 0
}

func (this *TLVObject) GetInt64(key int) int64 {
	return 0
}

func (this *TLVObject) GetString(key int) string {
	return ""
}

func (this *TLVObject) Put(key int, tlvObject TLVObject) {

}

func (this *TLVObject) PutBool(key int) {

}

func (this *TLVObject) PutInt8(key int) {

}

func (this *TLVObject) PutInt16(key int) {

}

func (this *TLVObject) PutInt32(key int) {

}

func (this *TLVObject) PutInt64(key int) {

}
