// $ go run main.go iphone
package main

import (
	"context"
	"encoding/json"
	"io"
	"log"
	"os"

	"google.golang.org/grpc"

	pb "price_comparison/product"
)

func main() {
	// Get parameter
	if len(os.Args) < 2 {
		log.Fatal("no parameter")
	}
	searchKeyWord := os.Args[1]
	log.Println("Your input:", searchKeyWord)


	// Open the config.
	jsonFile, err := os.Open("config.json")
	if err != nil {
		log.Fatal(err)
	}
	defer jsonFile.Close()

	// Parse
	var config map[string]string
	if err = json.NewDecoder(jsonFile).Decode(&config); err != nil {
		log.Fatal(err)
	}

	// / Connect to GRPC service
	conn, err := grpc.Dial(config["ip"], grpc.WithInsecure())
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()

	client := pb.NewProductServiceClient(conn)
	stream, err := client.GetProductInfo(context.Background(), &pb.ProductRequest{KeyWord: searchKeyWord})
	if err != nil {
		log.Fatal(err)
	}

	// Receive
	for {
		reply, err := stream.Recv()
		if err == io.EOF {
			log.Println("Done")
			return
		}
		if err != nil {
			log.Fatal(err)
		}
		log.Printf("reply : %v\n", reply)
	}
}
