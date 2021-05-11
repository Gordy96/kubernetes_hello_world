package main

import (
	"errors"
	"goinv/domain"
	"goinv/service"
	"net/http"

	"github.com/google/uuid"
	"github.com/gorilla/mux"

	pb "goinv/infrastructure/protobuf"

	"google.golang.org/grpc"

	"context"
	"log"
	"os"
	"os/signal"
	"time"
)

type repo struct {
	tasks map[string]*domain.Task
}

func (r *repo) Find(id uuid.UUID) (*domain.Task, error) {
	task, has := r.tasks[id.String()]
	if !has {
		return nil, errors.New("not found")
	}
	return task, nil
}

func (r *repo) Save(task *domain.Task) error {
	r.tasks[task.ID.String()] = task
	return nil
}
func (r *repo) FindAll() (domain.Tasks, error) {
	result := make(domain.Tasks, 0)
	for _, task := range r.tasks {
		result = append(result, task)
	}
	return result, nil
}

func start(ctx context.Context) error {
	address := os.Getenv("DOWNLOADER_HOST")
	if address == "" {
		address = "downloader"
	}
	conn, err := grpc.Dial(address+":50051", grpc.WithInsecure(), grpc.WithBlock())
	if err != nil {
		log.Fatalf("did not connect: %v", err)
	}
	defer conn.Close()
	c := pb.NewDownloaderClient(conn)

	uuidRegex := "[a-fA-F0-9]{8}-[a-fA-F0-9]{4}-[a-fA-F0-9]{4}-[a-fA-F0-9]{4}-[a-fA-F0-9]{12}"
	reppository := &repo{
		tasks: make(map[string]*domain.Task),
	}
	service := service.New(reppository, c, nil)
	main := mux.NewRouter()
	router := main.PathPrefix("/api/v1/tasks").Subrouter()
	router.Handle("/", http.HandlerFunc(service.FindAll)).Methods("GET")
	router.Handle("/", http.HandlerFunc(service.Create)).Methods("POST")
	router.Handle("/{id:"+uuidRegex+"}", http.HandlerFunc(service.Find)).Methods("GET")
	srv := &http.Server{
		Addr:    "0.0.0.0:8000",
		Handler: router,
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
