package disk_test

import (
	"reflect"
	"testing"

	"github.com/ilyabrin/disk"
	"gopkg.in/dnaeon/go-vcr.v3/cassette"
	"gopkg.in/dnaeon/go-vcr.v3/recorder"
)

const (
	TEST_DATA_DIR = "vcr/cassettes/"

	TEST_ACCESS_TOKEN    = "test"
	TEST_DIR_NAME        = "test_dir"
	TEST_DIR_NAME_COPY   = "test_dir_copy"
	TEST_PUBLIC_RESOURCE = "https://disk.yandex.ru/d/tCgV7GyS3QAYvg"
	TEST_TRASH_FILE_PATH = "trash:/___golang_API_dir_2_ddf8722d0aec88bfeb94a45a155511dbe151b764"
)

var client *disk.Client

func useCassette(path string) *recorder.Recorder {
	vcr, err := recorder.NewWithOptions(&recorder.Options{
		CassetteName:       TEST_DATA_DIR + path,
		Mode:               recorder.ModeRecordOnce,
		SkipRequestLatency: true,
	})
	if err != nil {
		panic(err)
	}

	vcr.AddHook(func(i *cassette.Interaction) error {
		delete(i.Request.Headers, "Authorization")
		return nil
	}, recorder.AfterCaptureHook)

	client = disk.New(TEST_ACCESS_TOKEN)
	client.HTTPClient.Transport = vcr

	return vcr
}

// TODO: use another method for testing got == expect
// TODO: change reflect to cmp package
func checkTypes(got, expect any, t *testing.T) {
	if reflect.TypeOf(got).Kind() != reflect.TypeOf(expect).Kind() {
		t.Fatalf("error: expect %v, got %v", expect, got)
	}
}
