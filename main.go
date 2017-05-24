package main

import (
	"bufio"
	"flag"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
)

func usage() {
	fmt.Fprintf(os.Stderr, "Usage of %s:\n\n", os.Args[0])
	fmt.Fprintf(os.Stderr, "clearbucket [flags] <bucket-name>\n\n")
	flag.PrintDefaults()
}

func main() {
	var region = flag.String("region", "us-west-1", "The region the bucket is in.")
	var verbose = flag.Bool("verbose", false, "If we should noisily print items deleted.")
	flag.Parse()

	if flag.NArg() < 1 {
		usage()
		os.Exit(1)
	}
	bname := flag.Arg(0)
	if strings.Contains(bname, "production") {
		reader := bufio.NewReader(os.Stdin)
		fmt.Println("Looks like this might be a production bucket...")
		fmt.Println("Since this deletes all versions of everything\nirrecoverably, please confirm.")
		fmt.Print("Enter the full bucket name: ")
		verify, err := reader.ReadString('\n')
		if err != nil {
			log.Fatalln(err)
		}
		if verify != bname {
			fmt.Println("Input didn't match, exiting")
			os.Exit(1)
		}
	}

	cfg := &aws.Config{Region: region}
	sess := session.Must(session.NewSession(cfg))
	svc := s3.New(sess)

	bucket := aws.String(bname)

	params := &s3.ListObjectVersionsInput{
		Bucket: bucket,
	}

	err := svc.ListObjectVersionsPages(params,
		func(page *s3.ListObjectVersionsOutput, lastPage bool) bool {
			var oids []*s3.ObjectIdentifier
			for _, p := range page.Versions {
				oids = append(oids, &s3.ObjectIdentifier{
					Key:       p.Key,
					VersionId: p.VersionId})
			}
			out, err := svc.DeleteObjects(&s3.DeleteObjectsInput{
				Bucket: bucket,
				Delete: &s3.Delete{
					Objects: oids,
				},
			})
			if err != nil {
				fmt.Fprintln(os.Stderr, err.Error())
			}
			if *verbose {
				fmt.Println(out)
			}

			return true
		})
	if err != nil {
		log.Fatalln(err.Error())
	}
}
