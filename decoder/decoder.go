package decoder

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

func (l *Lexer) readBytes(n int) []byte {
	bs := l.b[l.position : l.position+n]
	for i := 0; i < n; i++ {
		l.next()
	}
	return bs
}

func (l *Lexer) hasNext() bool {
	return l.readPosition < len(l.b)
}

func (l *Lexer) next() {
	l.position++
	l.readPosition = l.position + 1
}

//func (l *Lexer) peekByte() byte {
//	return l.b[l.readPosition]
//}

type Person struct {
	Name *Name // tag: 1
	Age  *Age  // tag: 2
}

// TODO: tagのslice or mapが必要
//  それぞれの型ごとに必要はなず。でもそれぞれのtag番号はgivenとする
//  field名を保持するものが必要。reflectで取ってくる？既知なので与えてしまってもよいはず
//  公式でもreflect.TypeOfを使っているのでそこは必要になるはず

//func (p *Person) getTags() {
//
//}

func (p *Person) Unmarshal(b []byte) error {
	l := New(b)
	for l.hasNext() {
		key := uint64(l.readCurByte())
		tag := key >> 3
		wire := int(key) & 7

		switch wire {
		case 2:
			length := int(l.readCurByte())
			v := l.readBytes(length)

			switch tag {
			case 1:
				p.Name = &Name{}
				p.Name.Unmarshal(v)
			case 2:
				p.Age = &Age{}
				p.Age.Unmarshal(v)
			}

		// TODO: case 0は別に分ける必要がある
		//  先頭1bitを切り落とさないといけない(7bitで以内なら1byteで済むので)
		case 0, 1, 5:
			length := types[wire]

			v := l.readBytes(length)
			_ = v // TODO
		default:
			l.next()
		}
	}

	return nil
}

type Name struct {
	Value string // tag: 1
}

func (n *Name) Unmarshal(b []byte) error {
	l := New(b)
	for l.hasNext() {
		key := uint64(l.readCurByte())
		tag := key >> 3
		wire := int(key) & 7

		// TODO: この処理Person/Name/Ageで同じになってしまう？
		switch wire {
		case 2:
			length := int(l.readCurByte())
			v := l.readBytes(length)

			switch tag {
			case 1:
				n.Value = string(v)
			}

		// TODO: case 0は別に分ける必要がある
		//  先頭1bitを切り落とさないといけない(7bitで以内なら1byteで済むので)
		case 0, 1, 5:
			length := types[wire]

			v := l.readBytes(length)
			_ = v // TODO
		default:
			l.next()
		}

	}
	return nil
}

func (a *Age) Unmarshal(b []byte) error {
	l := New(b)
	for l.hasNext() {
		key := uint64(l.readCurByte())
		tag := key >> 3
		wire := int(key) & 7

		// TODO: この処理Person/Name/Ageで同じになってしまう？
		switch wire {
		case 2:
			length := int(l.readCurByte())
			v := l.readBytes(length)
			_ = v // TODO: やっぱり処理が型によって異なる？想定だけcaesを書いてdefaultでerrorを返す？

		// TODO: case 0は別に分ける必要がある
		//  先頭1bitを切り落とさないといけない(7bitで以内なら1byteで済むので)
		case 0, 1, 5:
			length := types[wire]
			v := l.readBytes(length)

			switch tag {
			case 1:
				_, i := decodeVarint(v)
				a.Value = int32(i)
			}
		default:
			l.next()
		}

	}
	return nil
}

type Age struct {
	Value int32 // tag: 1
}

func decodeVarint(bs []byte) (uint64, int) {
	// TODO: unimplemented
	if len(bs) == 1 {
		return uint64(bs[0]), int(bs[0])
	}

	return 0, 0
}
