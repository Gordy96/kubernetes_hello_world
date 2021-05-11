package main

import (
	"log"
	"net"
	"net/http"
	"path/filepath"
	"strings"

	"google.golang.org/grpc"

	pb "goinv/infrastructure/protobuf"

	"context"
	"io"
	"os"
	"os/signal"
	"time"
)

type downloader struct {
	storagePath string
	pb.UnimplementedDownloaderServer
}

func (d *downloader) Download(ctx context.Context, in *pb.DownloadRequest) (*pb.DownloadReply, error) {
	parts := strings.Split(in.GetUrl(), "/")
	filename := parts[len(parts)-1]
	go d.DownloadFile(in.GetUrl(), filename)
	return &pb.DownloadReply{Status: pb.DownloadReply_OK, Filename: filename}, nil
}

func (d *downloader) DownloadFile(url string, filename string) {
	response, err := http.Get(url)
	if err != nil {
		return
	}
	defer response.Body.Close()
	out, err := os.Create(filepath.Join(d.storagePath, filename))
	if err != nil {
		return
	}
	defer out.Close()
	io.Copy(out, response.Body)
}

func start(ctx context.Context) error {
	cwd, err := os.Getwd()
	if err != nil {
		panic(err)
	}
	path := filepath.Join(cwd, "storage", "files")
	if _, err := os.Stat(path); os.IsNotExist(err) {
		os.MkdirAll(path, os.ModePerm)
	}
	dl_service := downloader{
		storagePath: path,
	}

	lis, err := net.Listen("tcp", "0.0.0.0:50051")
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	s := grpc.NewServer()
	defer s.Stop()
	pb.RegisterDownloaderServer(s, &dl_service)
	go func() {
		if err := s.Serve(lis); err != nil {
			log.Fatalf("failed to serve: %v", err)
		}
	}()

	fs := http.FileServer(http.Dir(path))
	srv := &http.Server{
		Addr:    "0.0.0.0:8080",
		Handler: fs,
	}
	go func() {
		if err = srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("listen:%+s\n", err)
		}
	}()
	log.Printf("server started")
	<-ctx.Done()
	log.Printf("server stopped")
	return shutDownServerGracefully(srv)
}

func shutDownServerGracefully(srv *http.Server) (err error) {
	ctxShutDown, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer func() {
		cancel()
	}()
	if err = srv.Shutdown(ctxShutDown); err != nil {
		log.Fatalf("graceful shutdown Failed:%+s", err)
	}
	log.Printf("server exited properly")
	if err == http.ErrServerClosed {
		err = nil
	}
	return err
}

func makeGracefulShutdownContext() context.Context {
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)

	ctx, cancel := context.WithCancel(context.Background())

	go func() {
		oscall := <-c
		log.Printf("system call:%+v", oscall)
		cancel()
	}()
	return ctx
}

func main() {
	ctx := makeGracefulShutdownContext()
	if err := start(ctx); err != nil {
		log.Printf("failed to start:+%v\n", err)
	}
}
