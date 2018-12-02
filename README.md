# protobufDecoder

[Protocol Buffers](https://developers.google.com/protocol-buffers/)のバイナリをデコードしたい。

本内容は、あくまでprotobufの勉強を目的としたもので仕様には完璧に添えていません。
実運用する際には[公式](https://github.com/golang/protobuf/)を利用してください。

## バイナリの生成

```proto
syntax = "proto3";

package person;

message Person {
  Name name = 1;
  Age age = 2;
}

message Name {
  string value = 1;
}

message Age {
  int32 value = 1;
}
```

```sh
❯ protoc -I=./ --go_out=./ person.proto
```

```go
package main

import (
	"io/ioutil"
	"log"

	pb "github.com/cipepser/protobufDecoder/Person"
	"github.com/golang/protobuf/proto"
)

func main() {
	p := &pb.Person{
		Name: &pb.Name{
			Value: "Alice",
		},
		Age: &pb.Age{
			Value: 20,
		},
	}

	if err := write("./person/alice.bin", p); err != nil {
		log.Fatal(err)
	}
}

func write(file string, p *pb.Person) error {
	out, err := proto.Marshal(p)
	if err != nil {
		return err
	}

	if err := ioutil.WriteFile(file, out, 0644); err != nil {
		return err
	}

	return nil
}
```

```sh
❯ go run main.go
❯ vim person/alice.bin
:%!xxd
```

```bin
00000000: 0a07 0a05 416c 6963 6512 0208 14         ....Alice....
```

こちらの図と同じやつだった。

https://github.com/cipepser/protobuf-sample/blob/master/img/protobuf.png

```sh
❯ hexdump person/alice.bin
0000000 0a 07 0a 05 41 6c 69 63 65 12 02 08 14
000000d
```

## 前のやつからコピー

[Protocol Buffers のエンコーディング仕様の解説](https://qiita.com/aiueo4u/items/38195248a29e9ff719c7)にあるように以下が基本となる。

> key = タグナンバー * 8 + タイプ値

タイプ値は、[公式ドキュメント](https://developers.google.com/protocol-buffers/docs/encoding)でwire typesとして以下のように定義されている。

|  Type | Meaning | Used For |
|  ------ | ------ | ------ |
|  0 | Varint | int32, int64, uint32, uint64, sint32, sint64, bool, enum |
|  1 | 64-bit | fixed64, sfixed64, double |
|  2 | Length-delimited | string, bytes, embedded messages, packed repeated fields |
|  3 | Start group | groups (deprecated) |
|  4 | End group | groups (deprecated) |
|  5 | 32-bit | fixed32, sfixed32, float |


改めて以下をパースしていく。

```
0a 07 0a 05 41 6c 69 63 65 12 02 08 14
```

`()`内の数字は何進数で表記しているかを表す。

まず初めの`0a(16)`は、  
`10(10)` = タグ`name(1)` * 8 + `Length-delimited(type 2)`  
となる。  
※`Name`は自身で定義したmessageであり、表中の`embedded message`が該当し、`Length-delimited`となる。

続く`07(16)`は、`name`のlengthなので、`0a 05 41 6c 69 63 65`を`Name`として読んでいく。

なので、`Name`最初の`0a(16)`は、  
`10(10)` = タグ`value(1)` * 8 + `Length-delimited(type 2)`  
となる。

続く`05(16)`は、`value`のlengthなので、`41 6c 69 63 65`を`string`として読んでいく。
utf8(この文字範囲ならASCIIと同じだけど)として読むと`41 6c 69 63 65`は`Alice`となる。

このまま残りの`12 02 08 14`も読んでいく。

`0a(12)`は、  
`18(10)` = タグ`age(2)` * 8 + `Length-delimited(type 2)`  
となる。

続く`08(16)`は、`value(1)`のlengthなので、`14`を`int32`として読んで`14(16)` = `20(10)`

![protobuf](https://github.com/cipepser/protobuf-sample/blob/master/img/protobuf.png)



## References
* [Protocol Buffers](https://developers.google.com/protocol-buffers/)
* [Go support for Protocol Buffers - Google's data interchange format](https://github.com/golang/protobuf/)