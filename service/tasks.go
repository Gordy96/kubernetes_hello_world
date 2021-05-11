package service

import (
	"encoding/json"
	"goinv/domain"
	"log"
	"net/http"
	"strings"

	"context"
	pb "goinv/infrastructure/protobuf"
	"path/filepath"
	"time"

	"github.com/google/uuid"
)

func New(repository domain.Repository, download_service pb.DownloaderClient, logger *log.Logger) *Service {
	return &Service{
		repository: repository,
		logger:     logger,
		downloader: download_service,
	}
}

type Service struct {
	repository domain.Repository
	logger     *log.Logger
	downloader pb.DownloaderClient
}

func (s *Service) Find(w http.ResponseWriter, r *http.Request) {
	parts := strings.Split(r.URL.Path, "/")
	id := uuid.MustParse(parts[len(parts)-1])

	task, err := s.repository.Find(id)
	w.Header().Set("content-type", "application/json")
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte(err.Error()))
	} else {
		json.NewEncoder(w).Encode(task)
	}
}

func (s *Service) FindAll(w http.ResponseWriter, r *http.Request) {
	tasks, err := s.repository.FindAll()
	w.Header().Set("content-type", "application/json")
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(err.Error()))
	} else {
		json.NewEncoder(w).Encode(tasks)
	}
}

func (s *Service) Create(w http.ResponseWriter, r *http.Request) {
	task := &domain.Task{
		ID:     uuid.New(),
		Status: domain.StatusNew,
	}
	temp := map[string]interface{}{}
	json.NewDecoder(r.Body).Decode(&temp)
	task.OriginURL = temp["origin_url"].(string)
	w.Header().Set("content-type", "application/json")
	err := s.repository.Save(task)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(err.Error()))
	} else {
		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()
		dr, err := s.downloader.Download(ctx, &pb.DownloadRequest{Url: task.OriginURL})
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte(err.Error()))
			return
		}
		if dr.Status == pb.DownloadReply_OK {
			task.Status = domain.StatusPending
			task.DownloadURL = filepath.Join(r.Host, dr.Filename)
			err = s.repository.Save(task)
			if err != nil {
				w.WriteHeader(http.StatusBadRequest)
				w.Write([]byte(err.Error()))
				return
			}
		}

		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(task)
	}
}
