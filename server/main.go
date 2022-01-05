package main

import (
	"fmt"
	"log"
	"net"

	"google.golang.org/grpc"

	"price_comparison/model"
	pb "price_comparison/product"
	"price_comparison/sql"
	"price_comparison/worker"
)

// Log setting initialize from /price_comparison/model
// SQL setting initialize from /price_comparison/sql

type Server struct {
}

type ProductGRPC struct {
	Products      chan pb.ProductResponse
	FinishRequest chan int
}

func (s *Server) GetProductInfo(in *pb.ProductRequest, stream pb.ProductService_GetProductInfoServer) error {
	log.Println("Search for", in.KeyWord)

	// Search in the database.
	products, err := sql.Select(in.KeyWord)
	if err != nil {
		return err
	}

	// Output it directly, if there are data in the database.
	if len(products) > 0 {
		// Push the data to client from the database.
		for _, product := range products {
			err := stream.Send(&pb.ProductResponse{
				Name:       product.Name,
				Price:      int32(product.Price),
				ImageURL:   product.ImageURL,
				ProductURL: product.ProductURL,
			})
			if err != nil {
				log.Println("client closed")
				return err
			}
		}
		return nil
	}

	var p ProductGRPC
	p.Products = make(chan pb.ProductResponse, 200)
	p.FinishRequest = make(chan int, 1)

	// Search for keyword in webs, then push the data to grpc output buffer.
	go func() {
		// load worker config
		var workerConfig model.WorkerConfig
		if err := model.OpenJsonEncodeStruct("../config/worker.json", &workerConfig); err != nil {
			log.Println(err, "failed to open worker config")
		}
		worker.Queue(stream.Context(), in.KeyWord, p.Products, workerConfig)
		// Check all products have been send, then finish this grpc request.
		for {
			select {
			case <-stream.Context().Done():
				log.Println("..........ctx canceled...........", stream.Context().Err())
				return
			default:
				if len(p.Products) == 0 {
					p.FinishRequest <- 1
					return
				}
			}
		}
	}()

	// Output the data to the client from buffer.
	for {
		select {
		case product := <-p.Products:
			err := stream.Send(&product)
			if err != nil {
				log.Println("client closed")
				return err
			}
		case <-p.FinishRequest:
			log.Println("Done!")
			return nil
		case <-stream.Context().Done():
			log.Println("Time out")
			return nil
		}
	}
}

func main() {
	log.Println("---------- Service started ---------")

	// Read the grpc config.
	grpcConfig, err := model.OpenJson("../config/grpc.json")
	if err != nil {
		log.Fatal(err)
	}

	// Start the GRPC service.
	grpcServer := grpc.NewServer()
	pb.RegisterProductServiceServer(grpcServer, &Server{})
	listen, err := net.Listen("tcp", fmt.Sprintf(":%v", grpcConfig["port"]))
	if err != nil {
		log.Fatal(err)
	}
	grpcServer.Serve(listen)
}
