package main

import (
	"context"
	"fmt"
	"log"
	"os"
	sync "sync"
	"time"

	storage "cloud.google.com/go/bigquery/storage/apiv1beta2"
	"cloud.google.com/go/bigquery/storage/managedwriter/adapt"
	"github.com/kennycoder/sku-crawler/crawlers"
	storagepb "google.golang.org/genproto/googleapis/cloud/bigquery/storage/v1beta2"
	"google.golang.org/protobuf/proto"
)

var (
	project = os.Getenv("PROJECT_ID")
	dataset = os.Getenv("BQ_DATASET")
	table   = os.Getenv("BQ_TABLE")
	trace   = "price_craweler"
)

var wg sync.WaitGroup

func scrape() {
	crawlerList := []crawlers.Crawler{crawlers.Mediamarkt{}, crawlers.Mindfactory{}}

	for _, crawler := range crawlerList {
		fetch(crawler)
	}
}

func fetch(crawler crawlers.Crawler) {
	// the data we will stream to bigquery

	var rows = []Row{}

	select {
	case products := <-crawler.Fetch(&wg):
		for _, product := range products {
			fmt.Printf("%s: - %f - [%s]\n", product.Name, product.Price, product.Source)
			rows = append(rows, Row{
				Name:      product.Name,
				Price:     product.Price,
				Source:    product.Source,
				Timestamp: time.Now().UnixMicro(),
			})
		}
	default:
		return

	}

	wg.Wait()
	fmt.Println("Done!")

	ctx := context.Background()

	// create the bigquery client
	log.Println("creating the bigquery client...")
	client, err := storage.NewBigQueryWriteClient(ctx)
	if err != nil {
		log.Fatal(err)
	}
	defer client.Close()

	// get the protobuf descriptor for our row type
	log.Println("creating descriptor...")
	var row Row
	descriptor, err := adapt.NormalizeDescriptor(row.ProtoReflect().Descriptor())
	if err != nil {
		log.Fatal("NormalizeDescriptor: ", err)
	}

	// create the write stream
	// a COMMITTED write stream inserts data immediately into bigquery
	log.Println("creating the write stream...")
	resp, err := client.CreateWriteStream(ctx, &storagepb.CreateWriteStreamRequest{
		Parent: fmt.Sprintf("projects/%s/datasets/%s/tables/%s", project, dataset, table),
		WriteStream: &storagepb.WriteStream{
			Type: storagepb.WriteStream_COMMITTED,
		},
	})
	if err != nil {
		log.Fatal("CreateWriteStream: ", err)
	}

	// get the stream by calling AppendRows
	log.Println("calling AppendRows...")
	stream, err := client.AppendRows(ctx)
	if err != nil {
		log.Fatal("AppendRows: ", err)
	}

	// serialize the rows
	log.Println("marshalling the rows...")
	var opts proto.MarshalOptions
	var data [][]byte
	for _, row := range *&rows {
		buf, err := opts.Marshal(&row)
		if err != nil {
			log.Fatal("protobuf.Marshal: ", err)
		}
		data = append(data, buf)
	}

	// send the rows to bigquery
	log.Println("sending the data...")
	err = stream.Send(&storagepb.AppendRowsRequest{
		WriteStream: resp.Name,
		TraceId:     trace, // identifies this client
		Rows: &storagepb.AppendRowsRequest_ProtoRows{
			ProtoRows: &storagepb.AppendRowsRequest_ProtoData{
				// protocol buffer schema
				WriterSchema: &storagepb.ProtoSchema{
					ProtoDescriptor: descriptor,
				},
				// protocol buffer data
				Rows: &storagepb.ProtoRows{
					SerializedRows: data, // serialized protocol buffer data
				},
			},
		},
	})
	if err != nil {
		log.Fatal("AppendRows.Send: ", err)
	}

	// get the response, which will tell us whether it worked
	log.Println("waiting for response...")
	r, err := stream.Recv()
	if err != nil {
		log.Fatal("AppendRows.Recv: ", err)
	}

	if rErr := r.GetError(); rErr != nil {
		log.Printf("result was error: %v", rErr)
	} else if rResult := r.GetAppendResult(); rResult != nil {
		log.Printf("now stream offset is %d", rResult.Offset.Value)
	}

	log.Println("done")

}

func main() {
	scrape()
}
