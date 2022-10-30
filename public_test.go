package disk

import (
	"context"
	"testing"

	"gopkg.in/dnaeon/go-vcr.v3/cassette"
	"gopkg.in/dnaeon/go-vcr.v3/recorder"
)

const TEST_PUBLIC_RESOURCE = "https://disk.yandex.ru/d/tCgV7GyS3QAYvg"

func TestGetMetadataForPublicResource(t *testing.T) {
	rec, err := recorder.New(TEST_DATA_DIR + "/public/get_meta")
	if err != nil {
		t.Fatal(err)
	}
	defer rec.Stop()

	hookDeleteToken := func(i *cassette.Interaction) error {
		delete(i.Request.Headers, "Authorization")
		return nil
	}
	rec.AddHook(hookDeleteToken, recorder.AfterCaptureHook)

	if rec.Mode() != recorder.ModeRecordOnce {
		t.Fatal("Recorder should be in ModeRecordOnce")
	}

	client := New(TEST_ACCESS_TOKEN)
	client.HTTPClient = rec.GetDefaultClient()

	ctx := context.Background()
	resp, errorResponse := client.GetMetadataForPublicResource(ctx, TEST_PUBLIC_RESOURCE, nil)

	if errorResponse != nil {
		t.Fatal(errorResponse)
	}

	publicResource := new(PublicResource)
	if publicResource == resp {
		t.Fatalf("error: expect %v, got %v", publicResource, resp)
	}
}

func TestGetDownloadURLForPublicResource(t *testing.T) {
	rec, err := recorder.New(TEST_DATA_DIR + "/public/download_url")
	if err != nil {
		t.Fatal(err)
	}
	defer rec.Stop()

	hookDeleteToken := func(i *cassette.Interaction) error {
		delete(i.Request.Headers, "Authorization")
		return nil
	}
	rec.AddHook(hookDeleteToken, recorder.AfterCaptureHook)

	if rec.Mode() != recorder.ModeRecordOnce {
		t.Fatal("Recorder should be in ModeRecordOnce")
	}

	client := New(TEST_ACCESS_TOKEN)
	client.HTTPClient = rec.GetDefaultClient()

	ctx := context.Background()
	resp, errorResponse := client.GetDownloadURLForPublicResource(ctx, TEST_PUBLIC_RESOURCE, nil)

	if errorResponse != nil {
		t.Fatal(errorResponse)
	}

	link := new(Link)
	if link == resp {
		t.Fatalf("error: expect %v, got %v", link, resp)
	}
}

func TestSavePublicResource(t *testing.T) {
	rec, err := recorder.New(TEST_DATA_DIR + "/public/save")
	if err != nil {
		t.Fatal(err)
	}
	defer rec.Stop()

	hookDeleteToken := func(i *cassette.Interaction) error {
		delete(i.Request.Headers, "Authorization")
		return nil
	}
	rec.AddHook(hookDeleteToken, recorder.AfterCaptureHook)

	if rec.Mode() != recorder.ModeRecordOnce {
		t.Fatal("Recorder should be in ModeRecordOnce")
	}

	client := New(TEST_ACCESS_TOKEN)
	client.HTTPClient = rec.GetDefaultClient()

	ctx := context.Background()
	resp, errorResponse := client.SavePublicResource(ctx, TEST_PUBLIC_RESOURCE, nil)

	if errorResponse != nil {
		t.Fatal(errorResponse)
	}

	link := new(Link)
	if link == resp {
		t.Fatalf("error: expect %v, got %v", link, resp)
	}
}
