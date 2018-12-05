package decoder

import (
	"errors"
	"fmt"
	"math/bits"
)

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

type Person struct {
	Name *Name // tag: 1
	Age  *Age  // tag: 2
}

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
			case 0:
				return errors.New("illegal tag 0")
			case 1:
				p.Name = &Name{}
				p.Name.Unmarshal(v)
			case 2:
				p.Age = &Age{}
				p.Age.Unmarshal(v)
			}
		default: // Person型はwire type 2以外は存在しない
			return fmt.Errorf("unexpected wire type: %d", wire)
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

		switch wire {
		case 2:
			length := int(l.readCurByte())
			v := l.readBytes(length)

			switch tag {
			case 0:
				return errors.New("illegal tag 0")
			case 1:
				n.Value = string(v)
			}
		default: // Name型はwire type 2以外は存在しない
			return fmt.Errorf("unexpected wire type: %d", wire)
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

		switch wire {
		case 0:
			switch tag {
			case 0:
				return errors.New("illegal tag 0")
			case 1:
				i, err := l.decodeVarint()
				if err != nil {
					return err
				}
				a.Value = int32(i)
			}
		default: // Age型はwire type 1以外は存在しない
			return fmt.Errorf("unexpected wire type: %d", wire)
		}

	}
	return nil
}

type Age struct {
	Value int32 // tag: 1
}

func (l *Lexer) decodeVarint() (uint64, error) {
	if len(l.b) == l.position {
		return 0, errors.New("unexpected EOF")
	}

	bs := []byte{}
	b := l.readCurByte()
	for bits.LeadingZeros8(b) == 0 { // 最上位bitが1のとき
		bs = append(bs, b&0x7f)
		b = l.readCurByte()
	}

	// 最上位bitが0のとき = 最後の1byte
	x := uint64(b)
	for i := 0; i < len(bs); i++ {
		x = x<<7 + uint64(bs[len(bs)-1-i])
	}

	return x, nil
}
