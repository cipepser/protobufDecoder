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

## 挙動チェック

`tag`を`0`にしたとき

```sh
panic: proto: person.Person: illegal tag 0 (wire type 2)
```

このとき`tag`は`1` or `2`だけど、`3`になるようにバイナリを作ると該当フィールドは`nil`になる。

`table_unmarshal.go`に以下のように書かれている。

> Explicitly disallow tag 0. This will ensure we flag an error
> when decoding a buffer of all zeros. Without this code, we
> would decode and skip an all-zero buffer of even length.
> [0 0] is [tag=0/wiretype=varint varint-encoded-0].

## protoを変えてみる

```proto
syntax = "proto3";

package person;

message Person {
    string name = 1;
    int32 age = 2;
}
```

上記`.proto`を用いて、以下をエンコードする。

```go
p := &pb.Person{
  Name: "Alice",
  Age:  20,
}
```

結果、`0a 05 41 6c 69 63 65 10 14`

`0a`でタグ`name(1)`,`Length-delimited(type 2)`で`05`byte読み込む。
`05`バイトが`41 6c 69 63 65`で`Alice`となる。
続いて`10`がタグ`age(2)`,`Varint(type 0)`となるので`14`を`Varint`として読み込む。

127を超えると`Varint`のバイト数が大きくなるので以下に変えて再度エンコードする。

```go
p := &pb.Person{
  Name: "Alice",
  Age:  131, // 20から131に変更
}
```

結果、`0a 05 41 6c 69 63 65 10 83 01`

前半`0a 05 41 6c 69 63 65`は上記と同様。

続いて`10`がタグ`age(2)`,`Varint(type 0)`となるので`83`をVarintとして読み込む。

ただし、`0x83` = `0b1000 0011`となり、先頭1bitが`1`のため、次の1byteも読み込む必要がある。
(実装するときは、128以上ならもう1byte読み込む処理にする)
次の1byteを読み込むと`0x01` = `0b0000 0001`となる。

`0x83`と`0x01`を組み合わせてVarintを読み込めばいいが、仕様に以下のように書かれているので、
リトルエンディアンで読んでいく。

> varints store numbers with the least significant group first

なお、先頭1bitは無視することにも注意する。

Varint = `0x01`(`0b0000 0001`)から先頭1bit落としたもの ++ `0x83`(`0b1000 0011`)から先頭1bit落としたもの
= `000 0001` ++ `000 0011`
= `0b1000 0011`
`0d131`

となりVarintが読み込める。


## References
* [Protocol Buffers](https://developers.google.com/protocol-buffers/)
* [Go support for Protocol Buffers - Google's data interchange format](https://github.com/golang/protobuf/)