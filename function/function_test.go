package function_test

import (
	"fmt"
	"log"
	"os"
	"strconv"
	"testing"
	"time"

	"github.com/baidu/openedge/runtime"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
)

func IgnoreTestFunctionRuntime(t *testing.T) {
	address := "127.0.0.1:65283"
	defaultName := "sayhi"

	// Set up a connection to the server.
	conn, err := grpc.Dial(address, grpc.WithInsecure())
	if err != nil {
		log.Fatalf("did not connect: %v", err)
	}
	defer conn.Close()
	c := runtime.NewRuntimeClient(conn)

	// Contact the server and print out its response.
	id := "1"
	if len(os.Args) > 1 {
		id = os.Args[1]
	}
	payload := fmt.Sprintf("{\"id\":%s}", id)

	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()

	msg := &runtime.Message{
		QOS:              1,
		Topic:            "test",
		Payload:          []byte(payload),
		FunctionName:     defaultName,
		FunctionInvokeID: strconv.Itoa(time.Now().Nanosecond()),
	}
	fmt.Println("[REQUEST]", msg)
	msg, err = c.Handle(ctx, msg)
	fmt.Println("[RESPONSE]", msg, err)
}
