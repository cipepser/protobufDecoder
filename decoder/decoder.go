package decoder

import "fmt"

type Lexer struct {
	b            []byte
	position     int
	readPosition int
}

func New(input []byte) *Lexer {
	l := &Lexer{b: input}
	l.readPosition = l.position + 1
	return l
}

func (l *Lexer) readCurByte() byte {
	b := l.b[l.position]
	l.next()
	return b
}

func (l *Lexer) readByte(n int) []byte {
	bs := l.b[l.position : l.position+n]
	for i := 0; i < n; i++ {
		l.next()
	}
	return bs
}

func (l *Lexer) next() {
	l.position++
	l.readPosition = l.position + 1
}

type Person struct {
	Name *Name // tag: 1
	Age  *Age  // tag: 2
}

type Name struct {
	Value string // tag: 1
}

type Age struct {
	Value int32 // tag: 1
}

// tagのslice or mapが必要
// それぞれの型ごとに必要はなず。でもそれぞれのtag番号はgivenとする
// field名を保持するものが必要。reflectで取ってくる？既知なので与えてしまってもよいはず

var (
	// 各wire typeごとに、あとに続くbyte数を保持
	types = map[int]int{
		0: 1, // Varint:	int32, int64, uint32, uint64, sint32, sint64, bool, enum
		1: 2, // 64-bit:	fixed64, sfixed64, double
		// unimplemented 2: , // Length-delimited:	string, bytes, embedded messages, packed repeated fields
		// unimplemented 3: , // Start: group	groups (deprecated)
		// unimplemented 4: , // End: group	groups (deprecated)
		5: 1, // 32-bit:	fixed32, sfixed32, float
	}
)

func (p *Person) Unmarshal(b []byte) error {
	l := New(b)
	for l.readPosition < len(l.b) {
		key := uint64(l.readCurByte())
		tag := key >> 3
		wire := int(key) & 7
		fmt.Println("-------------")
		fmt.Printf("tag: %x\n", tag)
		fmt.Printf("wire: %x\n", wire)

		switch wire {
		case 2:
			length := int(l.readCurByte())
			v := l.readByte(length)
			fmt.Printf("value: % x\n", v)
			// TODO: ここでは再帰的に呼ぶ必要がある
		case 0, 1, 5:
			length := types[wire]

			v := l.readByte(length)
			fmt.Printf("value: % x\n", v)
		default:
			l.next()
		}
	}

	p.Name = &Name{"Alice"}
	p.Age = &Age{20}
	return nil
}
