package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	modulehandler "github.com/mxab/tf-registry/internal/module/handler"

	"github.com/labstack/echo/v4"
	"github.com/mxab/tf-registry/internal/discovery"
	"github.com/mxab/tf-registry/internal/upload"
)

func server() *echo.Echo {
	e := echo.New()

	b := discovery.DiscoveryResponse{
		ModulesV1: "/v1/modules",
	}
	discovery.NewController(b)
	modulehandler.RegisterModuleControllerGroup(e.Group(b.ModulesV1), nil)

	return e
}
func main() {

	serverCmd := flag.NewFlagSet("server", flag.ExitOnError)
	serverConfig := serverCmd.String("config", "", "config file")

	uploadCmd := flag.NewFlagSet("upload", flag.ExitOnError)
	uploadModuleDir := uploadCmd.String("module-dir", ".", "module directory")

	if len(os.Args) < 2 {
		fmt.Println("expected 'foo' or 'bar' subcommands")
		os.Exit(1)
	}
	switch os.Args[1] {
	case "server":
		fmt.Printf("server %s", *serverConfig)
		serverCmd.Parse(os.Args[2:])
		//create a new s3 client
		launchServer()
	case "upload":
		uploadCmd.Parse(os.Args[2:])
		upload.UploadDir(*uploadModuleDir, "http://localhost:1323", "test", "test", "test", "test")

	default:
		flag.PrintDefaults()
		os.Exit(1)
	}

}

func launchServer() bool {
	sess, err := session.NewSession(&aws.Config{
		Region: aws.String("us-east-1")},
	)
	if err != nil {
		fmt.Printf("failed to create session, %v", err)
		return true
	}

	svc := s3.New(sess)
	svc.ListBuckets(nil)

	e := server()
	e.Logger.Fatal(e.Start(":1323"))
	return false
}
