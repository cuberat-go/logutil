// This program creates an error logger that writes error messages to a log
// file in the ~/log/custom_marshal/ directory. The log file is created on the
// first write. The log file name is generated based on the program name,
// current timestamp, and process ID.
package main

import (
	// Built-in/core modules.
	"log/slog"

	// Third-party modules.
	"github.com/cuberat-go/jsonutil"
	"github.com/cuberat-go/logutil"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"

	// First-party modules.
	"github.com/cuberat-go/logutil/examples/custom_marshal/proto_stuff/my_proto_stuff"
)

// Marshals protobuf messages to JSON using the protojson package. Other data
// types get marshaled using the builtin JSON package. This is an example of a
// custom marshal function that can be used with the
// logutil.HandlerOptions.MarshalFunc field.
func marshalFunc(v any) ([]byte, error) {
	protoMarshal := func(v proto.Message) ([]byte, error) {
		return protojson.MarshalOptions{
			UseProtoNames: true,
		}.Marshal(v)
	}
	enc := jsonutil.WithMarshalers(jsonutil.MarshalFunc(protoMarshal))
	return enc.Marshal(v)
}

func main() {
	opts := &logutil.HandlerOptions{
		MarshalFunc: marshalFunc,
	}
	errorHandler, closer := logutil.NewErrorHandler(opts)
	defer closer.Close()

	logger := slog.New(errorHandler)

	logger.Debug("This is a debug message", "key", "value")
	logger.Info("This is an info message", "key", "value")
	logger.Error("This is an example error message", "foo", "bar",
		"proto_message", &my_proto_stuff.ProtoWithOneof{
			TopField: "top_value",
			TestOneof: &my_proto_stuff.ProtoWithOneof_Field2{
				Field2: 42,
			},
		},
	)

}
