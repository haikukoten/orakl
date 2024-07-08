//nolint:all
package test

import (
	"context"
	"os"
	"testing"
	"time"

	"bisonai.com/orakl/node/pkg/admin/tests"
	"bisonai.com/orakl/node/pkg/aggregator"
	"bisonai.com/orakl/node/pkg/dal/api"
	"bisonai.com/orakl/node/pkg/dal/common"
	"bisonai.com/orakl/node/pkg/utils/request"
	wsfcommon "bisonai.com/orakl/node/pkg/websocketfetcher/common"
	"bisonai.com/orakl/node/pkg/wss"
	"github.com/stretchr/testify/assert"
)

func TestApiControllerRun(t *testing.T) {
	ctx := context.Background()
	clean, testItems, err := setup(ctx)
	if err != nil {
		t.Fatalf("error setting up test: %v", err)
	}
	defer func() {
		if cleanupErr := clean(); cleanupErr != nil {
			t.Logf("Cleanup failed: %v", cleanupErr)
		}
	}()

	assert.False(t, testItems.Controller.Collector.IsRunning)

	testItems.Controller.Start(ctx)
	time.Sleep(10 * time.Millisecond)
	assert.True(t, testItems.Controller.Collector.IsRunning)
}

func TestApiGetLatestAll(t *testing.T) {
	ctx := context.Background()
	clean, testItems, err := setup(ctx)
	if err != nil {
		t.Fatalf("error setting up test: %v", err)
	}
	defer func() {
		if cleanupErr := clean(); cleanupErr != nil {
			t.Logf("Cleanup failed: %v", cleanupErr)
		}
	}()

	testItems.Controller.Start(ctx)

	sampleSubmissionData, err := generateSampleSubmissionData(
		testItems.TmpConfig.ID,
		int64(15),
		time.Now(),
		1,
		"test-aggregate",
	)
	if err != nil {
		t.Fatalf("error generating sample submission data: %v", err)
	}

	aggregator.SetLatestGlobalAggregateAndProof(ctx, testItems.TmpConfig.ID, sampleSubmissionData.GlobalAggregate, sampleSubmissionData.Proof)

	result, err := tests.GetRequest[[]common.OutgoingSubmissionData](testItems.App, "/api/v1/dal/latest-data-feeds/all", nil)
	if err != nil {
		t.Fatalf("error getting latest data: %v", err)
	}

	expected, err := testItems.Collector.IncomingDataToOutgoingData(ctx, *sampleSubmissionData)
	if err != nil {
		t.Fatalf("error converting sample submission data to outgoing data: %v", err)
	}

	assert.Greater(t, len(result), 0)
	if len(result) > 0 {
		assert.Equal(t, *expected, result[0])
	}
}

func TestShouldFailWithoutApiKey(t *testing.T) {
	ctx := context.Background()
	clean, testItems, err := setup(ctx)
	if err != nil {
		t.Fatalf("error setting up test: %v", err)
	}
	defer func() {
		if cleanupErr := clean(); cleanupErr != nil {
			t.Logf("Cleanup failed: %v", cleanupErr)
		}
	}()

	testItems.Controller.Start(ctx)
	go testItems.App.Listen(":8090")
	resp, err := request.RequestRaw(request.WithEndpoint("http://localhost:8090/api/v1"))
	if err != nil {
		t.Fatalf("error getting latest data: %v", err)
	}

	assert.Equal(t, 200, resp.StatusCode)

	sampleSubmissionData, err := generateSampleSubmissionData(
		testItems.TmpConfig.ID,
		int64(15),
		time.Now(),
		1,
		"test-aggregate",
	)
	if err != nil {
		t.Fatalf("error generating sample submission data: %v", err)
	}

	aggregator.SetLatestGlobalAggregateAndProof(ctx, testItems.TmpConfig.ID, sampleSubmissionData.GlobalAggregate, sampleSubmissionData.Proof)

	result, err := request.RequestRaw(request.WithEndpoint("http://localhost:8090/api/v1/dal/latest-data-feeds/test-aggregate"))

	if err != nil {
		t.Fatalf("error getting latest data: %v", err)
	}

	assert.Equal(t, 401, result.StatusCode)
}

func TestApiGetLatest(t *testing.T) {
	ctx := context.Background()
	clean, testItems, err := setup(ctx)
	if err != nil {
		t.Fatalf("error setting up test: %v", err)
	}
	defer func() {
		if cleanupErr := clean(); cleanupErr != nil {
			t.Logf("Cleanup failed: %v", cleanupErr)
		}
	}()

	testItems.Controller.Start(ctx)

	sampleSubmissionData, err := generateSampleSubmissionData(
		testItems.TmpConfig.ID,
		int64(15),
		time.Now(),
		1,
		"test-aggregate",
	)
	if err != nil {
		t.Fatalf("error generating sample submission data: %v", err)
	}

	aggregator.SetLatestGlobalAggregateAndProof(ctx, testItems.TmpConfig.ID, sampleSubmissionData.GlobalAggregate, sampleSubmissionData.Proof)

	result, err := tests.GetRequest[common.OutgoingSubmissionData](testItems.App, "/api/v1/dal/latest-data-feeds/test-aggregate", nil)
	if err != nil {
		t.Fatalf("error getting latest data: %v", err)
	}

	expected, err := testItems.Collector.IncomingDataToOutgoingData(ctx, *sampleSubmissionData)
	if err != nil {
		t.Fatalf("error converting sample submission data to outgoing data: %v", err)
	}
	assert.Equal(t, *expected, result)
}

func TestApiWebsocket(t *testing.T) {
	ctx := context.Background()
	clean, testItems, err := setup(ctx)
	if err != nil {
		t.Fatalf("error setting up test: %v", err)
	}
	defer func() {
		if cleanupErr := clean(); cleanupErr != nil {
			t.Logf("Cleanup failed: %v", cleanupErr)
		}
	}()

	apiKey := os.Getenv("API_KEY")
	if apiKey == "" {
		t.Fatalf("apiKey required")
	}
	headers := map[string]string{"X-API-Key": apiKey}

	testItems.Controller.Start(ctx)

	go testItems.App.Listen(":8090")

	conn, err := wss.NewWebsocketHelper(ctx, wss.WithEndpoint("ws://localhost:8090/api/v1/dal/ws"), wss.WithRequestHeaders(headers))
	if err != nil {
		t.Fatalf("error creating websocket helper: %v", err)
	}

	err = conn.Dial(ctx)
	if err != nil {
		t.Fatalf("error dialing websocket: %v", err)
	}

	err = conn.Write(ctx, api.Subscription{
		Method: "SUBSCRIBE",
		Params: []string{"submission@test-aggregate"},
	})
	if err != nil {
		t.Fatalf("error subscribing to websocket: %v", err)
	}

	sampleSubmissionData, err := generateSampleSubmissionData(
		testItems.TmpConfig.ID,
		int64(15),
		time.Now(),
		1,
		"test-aggregate",
	)
	if err != nil {
		t.Fatalf("error generating sample submission data: %v", err)
	}

	err = testPublishData(ctx, *sampleSubmissionData)
	if err != nil {
		t.Fatalf("error publishing sample submission data: %v", err)
	}

	ch := make(chan any)
	go conn.Read(ctx, ch)

	expected, err := testItems.Collector.IncomingDataToOutgoingData(ctx, *sampleSubmissionData)
	if err != nil {
		t.Fatalf("error converting sample submission data to outgoing data: %v", err)
	}

	sample := <-ch

	result, err := wsfcommon.MessageToStruct[common.OutgoingSubmissionData](sample.(map[string]any))
	if err != nil {
		t.Fatalf("error converting sample to struct: %v", err)
	}
	assert.Equal(t, *expected, result)

	err = conn.Close()
	if err != nil {
		t.Fatalf("error closing websocket: %v", err)
	}
}