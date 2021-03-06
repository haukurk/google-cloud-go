/*
Copyright 2017 Google LLC

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package spanner

import (
	"testing"
	"time"

	"github.com/golang/protobuf/proto"
	"github.com/golang/protobuf/ptypes"
	"github.com/googleapis/gax-go/v2"
	edpb "google.golang.org/genproto/googleapis/rpc/errdetails"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

func TestRetryInfo(t *testing.T) {
	b, _ := proto.Marshal(&edpb.RetryInfo{
		RetryDelay: ptypes.DurationProto(time.Second),
	})
	trailers := map[string]string{
		retryInfoKey: string(b),
	}
	gotDelay, ok := extractRetryDelay(toSpannerErrorWithMetadata(status.Errorf(codes.Aborted, ""), metadata.New(trailers), true))
	if !ok || !testEqual(time.Second, gotDelay) {
		t.Errorf("<ok, retryDelay> = <%t, %v>, want <true, %v>", ok, gotDelay, time.Second)
	}
}

func TestRetryerRespectsServerDelay(t *testing.T) {
	t.Parallel()
	serverDelay := 50 * time.Millisecond
	b, _ := proto.Marshal(&edpb.RetryInfo{
		RetryDelay: ptypes.DurationProto(serverDelay),
	})
	trailers := map[string]string{
		retryInfoKey: string(b),
	}
	retryer := onCodes(gax.Backoff{}, codes.Aborted)
	err := toSpannerErrorWithMetadata(status.Errorf(codes.Aborted, "transaction was aborted"), metadata.New(trailers), true)
	maxSeenDelay, shouldRetry := retryer.Retry(err)
	if !shouldRetry {
		t.Fatalf("expected shouldRetry to be true")
	}
	if maxSeenDelay != serverDelay {
		t.Fatalf("Retry delay mismatch:\ngot: %v\nwant: %v", maxSeenDelay, serverDelay)
	}
}
