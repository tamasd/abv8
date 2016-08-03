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
	"sync"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestSyncProcessing(t *testing.T) {
	Convey("Given a v8 worker pool, an operation should work", t, func(c C) {
		wp, err := NewWorkerPool(2, 2, "synctest.js")
		So(err, ShouldBeNil)

		wp.Start()
		defer wp.Stop()

		const max = 0x100
		var wg sync.WaitGroup
		wg.Add(max)
		for i := 0; i < max; i++ {
			go func() {
				defer wg.Done()
				reply := wp.Send(PageData{})

				c.So(reply.Status, ShouldEqual, 200)
				c.So(reply.Page, ShouldEqual, "ok")
			}()
		}

		wg.Wait()
	})
}
