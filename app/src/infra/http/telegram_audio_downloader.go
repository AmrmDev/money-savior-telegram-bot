package http

import (
	"context"
	"errors"
	"fmt"
	"io"
	nethttp "net/http"
	"net/url"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"

	"money-savior-telegram-bot/app/src/adapter/controller"
)

const (
	audioDownloadTimeout = 15 * time.Second
	maxDownloadBytes = 5 << 20
)

var _ controller.AudioDownloader = (*TelegramAudioDownloader)(nil)

type fileLinker interface {
	GetFileDirectURL(fileID string) (string, error)
}

type TelegramAudioDownloader struct {
	files      fileLinker
	httpClient *nethttp.Client
}

func NewTelegramAudioDownloader(bot *tgbotapi.BotAPI) *TelegramAudioDownloader {
	return newAudioDownloader(bot, &nethttp.Client{Timeout: audioDownloadTimeout})
}

func newAudioDownloader(files fileLinker, httpClient *nethttp.Client) *TelegramAudioDownloader {
	return &TelegramAudioDownloader{files: files, httpClient: httpClient}
}

func (d *TelegramAudioDownloader) Download(ctx context.Context, fileID string) ([]byte, error) {
	link, err := d.files.GetFileDirectURL(fileID)
	if err != nil {
		return nil, fmt.Errorf("get telegram file: %w", withoutURL(err))
	}

	req, err := nethttp.NewRequestWithContext(ctx, nethttp.MethodGet, link, nil)
	if err != nil {
		return nil, errors.New("download telegram file: build request")
	}

	resp, err := d.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("download telegram file: %w", withoutURL(err))
	}
	defer resp.Body.Close()

	if resp.StatusCode != nethttp.StatusOK {
		return nil, fmt.Errorf("download telegram file: unexpected status %d", resp.StatusCode)
	}

	audio, err := io.ReadAll(io.LimitReader(resp.Body, maxDownloadBytes+1))
	if err != nil {
		return nil, fmt.Errorf("download telegram file: read body: %w", withoutURL(err))
	}
	if len(audio) == 0 {
		return nil, errors.New("download telegram file: empty file")
	}
	if len(audio) > maxDownloadBytes {
		return nil, errors.New("download telegram file: file too large")
	}

	return audio, nil
}

func withoutURL(err error) error {
	var urlErr *url.Error
	if errors.As(err, &urlErr) {
		return urlErr.Err
	}
	return err
}