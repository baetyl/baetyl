/*
 *
 * Copyright 2015 gRPC authors.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 *
 */

package main

import (
	"fmt"
	"log"
	"os"
	"strconv"
	"time"

	"github.com/baidu/openedge/module/function/runtime"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
)

const (
	address     = "127.0.0.1:65283"
	defaultName = "sayhi"
)

func main() {
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
