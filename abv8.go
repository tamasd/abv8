// Copyright 2016 Tamás Demeter-Haludka
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package abv8

import (
	"encoding/json"
	"io"
	"io/ioutil"
	"net/http"
	"sync"

	"github.com/ry/v8worker"
	"github.com/tamasd/ab"
)

type PageData struct {
}

type PageReply struct {
	Status int    `json:"status"`
	Page   string `json:"page"`
}

type v8message struct {
	data  PageData
	reply chan PageReply
}

type WorkerPool struct {
	workernum        uint8
	threadsPerWorker uint8
	ch               chan v8message
	clientJS         string
}

func NewWorkerPool(workers, threadsPerWorker uint8, clientPath string) (*WorkerPool, error) {
	data, err := ioutil.ReadFile(clientPath)
	if err != nil {
		return nil, err
	}
	return &WorkerPool{
		workernum:        workers,
		threadsPerWorker: threadsPerWorker,
		ch:               make(chan v8message),
		clientJS:         string(data),
	}, nil
}

func (wp *WorkerPool) spawn() {
	var once sync.Once
	worker := v8worker.New(discardSend, discardSendSync)
	worker.Load("client.js", wp.clientJS)
	for i := uint8(0); i < wp.threadsPerWorker; i++ {
		go func() {
			for msg := range wp.ch {
				payload := marshalJSON(msg.data)
				reply := worker.SendSync(payload)
				pageReply := PageReply{}
				unmarshalJSON(reply, &pageReply)
				msg.reply <- pageReply
			}

			once.Do(func() {
				worker.TerminateExecution()
				worker.Dispose()
			})
		}()
	}
}

func (wp *WorkerPool) Send(p PageData) PageReply {
	msg := v8message{
		data:  p,
		reply: make(chan PageReply),
	}

	wp.ch <- msg

	return <-msg.reply
}

func (wp *WorkerPool) Render(r *ab.Renderer, p PageData) *ab.Renderer {
	reply := wp.Send(p)
	r.SetCode(reply.Status).AddOffer("text/html", func(w http.ResponseWriter) {
		io.WriteString(w, reply.Page)
	})

	return r
}

func (wp *WorkerPool) Start() {
	for i := uint8(0); i < wp.workernum; i++ {
		wp.spawn()
	}
}

func (wp *WorkerPool) Stop() {
	close(wp.ch)
}

func marshalJSON(v interface{}) string {
	b, _ := json.Marshal(v)
	return string(b)
}

func unmarshalJSON(s string, v interface{}) {
	json.Unmarshal([]byte(s), v)
}

func discardSendSync(msg string) string {
	return ""
}

func discardSend(msg string) {
}
