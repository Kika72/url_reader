package service

import (
	"context"
	"encoding/base64"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"sync"

	"url_reader/config"
)

type WorkerFunc func (ctx context.Context, in chan string, out chan UrlResponse, wg *sync.WaitGroup)
type ReaderOption func(*urlReader)

type urlReader struct {
	cfg config.Config
	worker WorkerFunc
}

func NewUrlReader(cfg config.Config, options ...ReaderOption) UrlReader {
	u := urlReader{
		cfg: cfg,
	}

	for _, option := range options {
		option(&u)
	}
	return u
}

func WithWorker(w WorkerFunc) ReaderOption {
	return func(u *urlReader) { u.worker = w }
}

func (u urlReader) Read(ctx context.Context, data UrlRequests) (UrlResponses, error) {
	workersCount := u.cfg.MaxOutRequests
	if workersCount > len(data.Requests) {
		workersCount = len(data.Requests)
	}

	// channel for writing results
	destChan := make(chan UrlResponse, workersCount)
	// channel for sending tasks to workers
	sourceChan := make(chan string)

	wg := sync.WaitGroup{}
	wg.Add(workersCount)

	// context for cancellation process in case of error
	// parent context is request context
	workerCtx, workerCancel := context.WithCancel(ctx)
	defer workerCancel()

	// execute workers
	for i := 0; i < workersCount; i++ {
		worker := u.worker
		if worker == nil {
			worker = u.executeWorker
		}
		go worker(workerCtx, sourceChan, destChan, &wg)
	}

	// write to source channel
	go func() {
		defer func() {
			if err := recover(); err != nil {
				fmt.Printf("panic in requests source: %v\n", err)
			}
		}()

		defer close(sourceChan)

		reqLoop:
		for _, req := range data.Requests {
			select {
			case sourceChan <- req:
				continue
			case <- workerCtx.Done():
				// exit sender if process is canceled
				break reqLoop
			}
		}
	}()

	// execute results receiver
	var err error
	wgReceiver := sync.WaitGroup{}
	wgReceiver.Add(1)
	result := UrlResponses {
		Responses: make([]UrlResponse, 0, len(data.Requests)),
	}
	go func() {
		defer func() {
			if err := recover(); err != nil {
				fmt.Printf("panic in requests source: %v\n", err)
			}
		}()
		defer wgReceiver.Done()

		for resp := range destChan {
			if resp.Error != nil {
				err = resp.Error
				workerCancel()
				break
			}
			result.Responses = append(result.Responses, resp)
		}
	}()

	// wait until all workers are done. Then close destination channel
	wg.Wait()
	close(destChan)

	// wait until receiver writes all data to response
	wgReceiver.Wait()

	if err != nil {
		return UrlResponses{}, err
	}

	return result, nil
}

func (u urlReader) executeWorker(ctx context.Context, in chan string, out chan UrlResponse, wg *sync.WaitGroup) {
	go func() {
		defer func() {
			if err := recover(); err != nil {
				fmt.Printf("panic in requests source: %v\n", err)
			}
		}()
		defer wg.Done()

		cln := http.Client {
			Timeout: u.cfg.Timeout,
		}
		receiverLoop:
		for {
			select {
			case req, ok := <- in:
				if !ok {
					break receiverLoop
				}

				resp, err := cln.Get(req)
				if err != nil {
					out <- UrlResponse{
						Url:     req,
						Content: "",
						Error:   err,
					}
					return
				}

				body, err := func () ([]byte, error) {
					defer resp.Body.Close()
					return ioutil.ReadAll(resp.Body)
				}()

				result := UrlResponse{
					Url:     req,
				}

				switch {
				case resp.StatusCode >= http.StatusBadRequest:
					result.Error = errors.New(string(body))
				default:
					result.Content = base64.StdEncoding.EncodeToString(body)
				}

				out <- result
			case <-ctx.Done():
				break receiverLoop
			}
		}
	}()
}
