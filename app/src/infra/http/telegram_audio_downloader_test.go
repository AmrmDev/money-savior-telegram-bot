package http

import (
	"context"
	"errors"
	nethttp "net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"
)

const fakeToken = "123456:SECRET-TOKEN"

type fakeLinker struct {
	link   string
	err    error
	fileID string
}

func (f *fakeLinker) GetFileDirectURL(fileID string) (string, error) {
	f.fileID = fileID
	return f.link, f.err
}

func newTestDownloader(link string) (*TelegramAudioDownloader, *fakeLinker) {
	linker := &fakeLinker{link: link}
	return newAudioDownloader(linker, &nethttp.Client{Timeout: 2 * time.Second}), linker
}

func serve(handler nethttp.HandlerFunc) *httptest.Server {
	return httptest.NewServer(handler)
}

func TestDownload_ReturnsTheFileBytes(t *testing.T) {
	server := serve(func(w nethttp.ResponseWriter, r *nethttp.Request) {
		if r.Method != nethttp.MethodGet {
			t.Errorf("unexpected method %s", r.Method)
		}
		_, _ = w.Write([]byte("OggS-voice"))
	})
	defer server.Close()

	downloader, linker := newTestDownloader(server.URL + "/file/bot" + fakeToken + "/voice/file_1.oga")

	audio, err := downloader.Download(context.Background(), "file-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if string(audio) != "OggS-voice" || linker.fileID != "file-1" {
		t.Errorf("unexpected result: audio=%q fileID=%q", audio, linker.fileID)
	}
}

func TestDownload_LinkErrorDoesNotLeakTheToken(t *testing.T) {
	downloader, linker := newTestDownloader("")
	linker.err = &url.Error{
		Op:  "Post",
		URL: "https://api.telegram.org/bot" + fakeToken + "/getFile",
		Err: errors.New("connection refused"),
	}

	_, err := downloader.Download(context.Background(), "file-1")

	if err == nil {
		t.Fatal("expected an error")
	}
	if strings.Contains(err.Error(), "SECRET-TOKEN") || !strings.Contains(err.Error(), "connection refused") {
		t.Errorf("expected the cause without the URL, got %v", err)
	}
}

func TestDownload_ConnectionErrorDoesNotLeakTheToken(t *testing.T) {
	server := serve(func(nethttp.ResponseWriter, *nethttp.Request) {})
	link := server.URL + "/file/bot" + fakeToken + "/voice/file_1.oga"
	server.Close() // nothing listens on this address anymore

	downloader, _ := newTestDownloader(link)

	_, err := downloader.Download(context.Background(), "file-1")

	if err == nil {
		t.Fatal("expected an error")
	}
	if strings.Contains(err.Error(), "SECRET-TOKEN") {
		t.Errorf("the error leaks the token: %v", err)
	}
}

func TestDownload_FailsOnErrorStatus(t *testing.T) {
	server := serve(func(w nethttp.ResponseWriter, _ *nethttp.Request) {
		w.WriteHeader(nethttp.StatusNotFound)
		_, _ = w.Write([]byte("body that must not leak"))
	})
	defer server.Close()

	downloader, _ := newTestDownloader(server.URL + "/file/bot" + fakeToken + "/x.oga")

	_, err := downloader.Download(context.Background(), "file-1")

	if err == nil || !strings.Contains(err.Error(), "404") {
		t.Fatalf("expected a 404 error, got %v", err)
	}
	if strings.Contains(err.Error(), "must not leak") || strings.Contains(err.Error(), "SECRET-TOKEN") {
		t.Errorf("the error leaks data: %v", err)
	}
}

func TestDownload_RejectsEmptyAndHugeFiles(t *testing.T) {
	empty := serve(func(nethttp.ResponseWriter, *nethttp.Request) {})
	defer empty.Close()

	huge := serve(func(w nethttp.ResponseWriter, _ *nethttp.Request) {
		_, _ = w.Write(make([]byte, maxDownloadBytes+10))
	})
	defer huge.Close()

	for name, link := range map[string]string{"vazio": empty.URL, "grande demais": huge.URL} {
		downloader, _ := newTestDownloader(link)

		if _, err := downloader.Download(context.Background(), "file-1"); err == nil {
			t.Errorf("%s: expected an error", name)
		}
	}
}

func TestDownload_StopsWhenTheContextExpires(t *testing.T) {
	server := serve(func(_ nethttp.ResponseWriter, r *nethttp.Request) {
		select {
		case <-r.Context().Done():
		case <-time.After(2 * time.Second):
		}
	})
	defer server.Close()

	downloader, _ := newTestDownloader(server.URL + "/file/bot" + fakeToken + "/x.oga")

	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()

	_, err := downloader.Download(ctx, "file-1")

	if err == nil {
		t.Fatal("expected an error when the context expires")
	}
	if strings.Contains(err.Error(), "SECRET-TOKEN") {
		t.Errorf("the error leaks the token: %v", err)
	}
}

func TestDownload_FileIDErrorIsReturned(t *testing.T) {
	downloader, linker := newTestDownloader("")
	linker.err = errors.New("file not found")

	if _, err := downloader.Download(context.Background(), "nope"); err == nil {
		t.Error("expected an error when the file id cannot be resolved")
	}
}
