#1. 通信协议
在基于TCP进行网络开发时，我们需要按照一定的规则对要传输的数据进行编码，这样客户端和服务器端才能从TCP数据流中有效的解析出传输的数据，TLV就是这样中编码规则。

> 本文对TLV的描述参考于这位仁兄的博文:[自定义通信协议设计之TLV编码应用](http://my.oschina.net/maxid/blog/206546)，感谢他如此详细的介绍了TLV，就如他文中结束时的预告一样，本项目将实现一个go语言版本的TLV编码及解析库。

#2. TLV编码介绍
TLV是指由数据的类型Tag，数据的长度Length，数据的值Value组成的结构体，几乎可以描任意数据类型，TLV的Value也可以是一个TLV结构，正因为这种嵌套的特性，可以让我们用来包装协议的实现。

![TLV图片](https://ahq02g.dm1.livefilestore.com/y2pGVRwLpxg1OC06Gsg7iN_tG0sHAzb83R45u43PaLfZDshVot43rvfzX62n89oZmaKkjhNFoH0DShGWgsBC6qKHqr5WBmY6XpAfjkOvpxyPIw/TLV.png?psid=1)

## 2.1. Tag 描述
### 2.1.1 Tag首字节描述
![Tag图片](https://ahq02g.dm2303.livefilestore.com/y2pE8maaJOVi2hTlZv13O7S6LxqLsbTzFf7HCG-J-Rnxhg2UWvmKHMTT2tvFMs3zjJGEb7WIdgQE3d8Wu6HroKynVJG2n1j_yFr4ckHlad1-7w/TLV_DISC.png?psid=1)

- **第6~7位**
表示TLV的类型，00表示TLV描述的是基本数据类型(Primitive Frame, int,string,long…)，01表示用户自定义类型(Private Frame，常用于描述协议中的消息)。

- **第5位**
表示Value的编码方式，分别支持Primitive及Constructed两种编码方式, Primitive指以原始数据类型进行编码，Constructed指以TLV方式进行编码，0表示以Primitive方式编码，1表示以Constructed方式编码。

- **第0~4位**
当Tag Value小于0x1F(31)时，首字节0～4位用来描述Tag Value，否则0~4位全部置1，作为存在后续字节的标志，Tag Value将采用后续字节进行描述。

### 2.1.2 Tag后续描述
后续字节采用每个字节的0～6位（即7bit）来存储Tag Value, 第7位用来标识是否还有后续字节。
- **第7位**
描述是否还有后续字节，1表示有后续字节，0表示没有后续字节，即结束字节。
- **第0~6位**
填充Tag Value的对应bit(从低位到高位开始填充)，如：Tag Value为：0000001 11111111 11111111 (10进制：131071), 填充后实际字节内容为：10000111 11111111 01111111。

![Tag后续字节图片](https://ahq02g.dm1.livefilestore.com/y2p0tGpzvc24_EpddfBEZsdEqUBHaRIMFSom4izyJN4ryrf2boD7g4FfqyVtiSqmd5UOc9TuNxHwmsCmkm2JFD8hL-HlOYIcixa6BMgc9_RbgY/TAG_NB.png?psid=1)

## 2.1. Length 描述
描述Value部分所占字节的个数，编码格式分两类：定长方式（DefiniteForm）和不定长方式（IndefiniteForm），其中定长方式又包括短形式与长形式。

### 2.1.1 定长方式
定长方式中，按长度是否超过一个八位，又分为短、长两种形式，编码方式如下：

- **短形式**
字节第7位为0，表示Length使用1个字节即可满足Value类型长度的描述，范围在0~127之间的。

![短形式图片](https://ahq02g.dm2303.livefilestore.com/y2pTUflngsP_OT4c4ReAdXLsOjD88bI2HcMc1nKHN6bouH9UDCYdGhKk33-EVJ-66Ms2zFv56R724HjvFb1OwB1_DBt1HxA40dtO6qKAfzTpJI/LENGTH-S.png?psid=1)

- **长形式**
即Value类型的长度大于127时，Length需要多个字节来描述，这时第一个字节的第7位置为1，0~6位用来描述Length值占用的字节数，然后直将Length值转为byte后附在其后，如： Value大小占234个字节（11101010）,由于大于127，这时Length需要使用两个字节来描述，10000001 11101010。

![长形式图片](https://ahq02g.dm2302.livefilestore.com/y2pPaMKjeIKEYAljAyvYv2qXf-zukGgyLXdqTgHVOp3e-J7PyObfa_uLeTJPHa7Ny5gPMEeE-LB-_AnOE1YVIC_gA08rP8vfh17yQuw7ngjow8/LENGTH-L.png?psid=1)

### 2.1.2 不定长方式
Length所在八位组固定编码为0x80，但在Value编码结束后以两个0x00结尾。这种方式使得可以在编码没有完全结束的情况下，可以先发送部分数据给对方。

![不定长图片](https://ahq02g.dm2301.livefilestore.com/y2p8bAu4O1EEq4cCoORp0uogPl7-CCyC2k31Rdimj1MyNQHVFp47GgO-0oJdsMhshg8zZND53TsNP6lcigss-FvdC8OD_zu4icx49H5NyCzU8w/LENGHT-D.png?psid=1)

# 3. Value 描述
由一个或多个值组成 ，值可以是一个原始数据类型(Primitive Data)，也可以是一个TLV结构(Constructed Data)。

## 3.1. Primitive Data 编码
![Primitive Data图片](https://ahq02g.dm2303.livefilestore.com/y2pr4FIH8dqYIpQvawEBEajtRXuWPFHRb9zS3EeMttlyi_TJjWTIQgg9MQw2v_qVr740-w6kcn_e6RseACqeUlIeYXiTozKo6lT-1HYuv6rdYY/P-DATA.png?psid=1)

## 3.2. Constructed Data 编码
![Constructed Data图片](https://ahq02g.dm2304.livefilestore.com/y2pafCW8TjTzhOrF86tdHK7Qrfl_01j4lZFrKYObH_Y1ACBcMmo1dat9Eohp30bJKLuDVxo_Y_nwN1wy93gddHzgVh_SbJcXTQD48At8DE2SQI/C-DATA.png?psid=1)







