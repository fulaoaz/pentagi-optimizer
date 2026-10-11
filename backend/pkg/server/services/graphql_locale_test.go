package services

import (
	"context"
	"errors"
	"testing"
)

func TestLocalizedGraphQLErrorPresenter(t *testing.T) {
	t.Run("Chinese request", func(t *testing.T) {
		ctx := withGraphQLLanguage(context.Background(), "zh-CN")
		got := localizedGraphQLErrorPresenter(ctx, errors.New("model provider is required"))
		if got.Message != "请选择模型提供商" {
			t.Fatalf("message = %q", got.Message)
		}
	})

	t.Run("English request", func(t *testing.T) {
		ctx := withGraphQLLanguage(context.Background(), "en-US")
		got := localizedGraphQLErrorPresenter(ctx, errors.New("model provider is required"))
		if got.Message != "model provider is required" {
			t.Fatalf("message = %q", got.Message)
		}
	})
}
