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
