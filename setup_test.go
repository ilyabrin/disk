package disk

import (
	"errors"

	"gopkg.in/dnaeon/go-vcr.v3/cassette"
	"gopkg.in/dnaeon/go-vcr.v3/recorder"
)

const (
	TEST_DATA_DIR = "testdata/responses/"

	TEST_ACCESS_TOKEN    = "test"
	TEST_DIR_NAME        = "test_dir"
	TEST_DIR_NAME_COPY   = "test_dir_copy"
	TEST_PUBLIC_RESOURCE = "https://disk.yandex.ru/d/tCgV7GyS3QAYvg"
	TEST_TRASH_FILE_PATH = "trash:/___golang_API_dir_2_ddf8722d0aec88bfeb94a45a155511dbe151b764"
)

var client *Client

// Runs before any test
func init() {
	client = New(TEST_ACCESS_TOKEN)
}

func useCassette(path string) error {
	r, err := recorder.New(TEST_DATA_DIR + path)
	if err != nil {
		panic(err)
	}
	defer r.Stop()

	client.httpClient = r.GetDefaultClient()
	defer r.Stop()

	hookDeleteToken := func(i *cassette.Interaction) error {
		delete(i.Request.Headers, "Authorization")
		return nil
	}
	r.AddHook(hookDeleteToken, recorder.AfterCaptureHook)

	if r.Mode() != recorder.ModeRecordOnce {
		return errors.New("Recorder should be in ModeRecordOnce")
	}

	return nil
}