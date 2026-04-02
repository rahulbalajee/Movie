package main

import (
	"encoding/json"
	"encoding/xml"
	"fmt"

	"github.com/rahulbalajee/Movie/gen"
	"github.com/rahulbalajee/Movie/metadata/pkg/model"
	"google.golang.org/protobuf/proto"
)

var metadata = &model.Metadata{
	ID:          "1",
	Title:       "Rahul Balajee",
	Description: "New movie metadata",
	Director:    "Balajee",
}

var genMetadata = &gen.Metadata{
	Id:          "1",
	Title:       "Rahul Balajee",
	Description: "New movie metadata",
	Director:    "Balajee",
}

func main() {
	jsonBytes, err := serializeToJSON(metadata)
	if err != nil {
		panic(err)
	}

	xmlBytes, err := serializeToXML(metadata)
	if err != nil {
		panic(err)
	}

	protocBytes, err := serializeToProtoc(genMetadata)
	if err != nil {
		panic(err)
	}

	fmt.Printf("JSON size: \t%dB\n", len(jsonBytes))
	fmt.Printf("XML size: \t%dB\n", len(xmlBytes))
	fmt.Printf("Proto size: \t%dB\n", len(protocBytes))
}

func serializeToJSON(m *model.Metadata) ([]byte, error) {
	return json.Marshal(m)
}

func serializeToXML(m *model.Metadata) ([]byte, error) {
	return xml.Marshal(m)
}

func serializeToProtoc(m *gen.Metadata) ([]byte, error) {
	return proto.Marshal(m)
}
