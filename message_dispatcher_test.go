package main

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/taylankasap/message-sender/api"

	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/require"
	somethirdparty "github.com/taylankasap/message-sender/some_third_party"
	"go.uber.org/mock/gomock"
)

func TestMessageDispatcher_Start(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockDB := NewMockDBInterface(ctrl)
	mockClient := somethirdparty.NewMockClientWithResponsesInterface(ctrl)

	t.Run("it should start dispatching the first batch of messages without waiting the period", func(tt *testing.T) {
		fakeMsg := api.Message{
			Id:        1,
			Content:   "Hello",
			Recipient: "+1234567890",
			Status:    api.Unsent,
		}

		// these should be called once
		mockDB.EXPECT().FetchUnsentMessages(1).Return([]api.Message{fakeMsg}, nil).Times(1)
		mockClient.EXPECT().SendMessageWithResponse(gomock.Any(), gomock.Any(), gomock.Any()).Return(
			&somethirdparty.SendMessageResponse{JSON202: &somethirdparty.APIResponse{MessageId: "dummy-message-id"}},
			nil,
		).Times(1)
		mockDB.EXPECT().MarkMessageAsSent(fakeMsg.Id, gomock.Any()).Return(nil).Times(1)

		dispatcher := &MessageDispatcher{
			DB:        mockDB,
			Client:    mockClient,
			BatchSize: 1,
			Period:    2 * time.Minute,
		}

		go dispatcher.Start()
		time.Sleep(1 * time.Millisecond)
	})
}

func TestMessageDispatcher_Pause(t *testing.T) {
	d := &MessageDispatcher{
		pauseCh:  make(chan struct{}),
		resumeCh: make(chan struct{}),
	}

	require.False(t, d.paused, "dispatcher should not be paused initially")

	d.Pause()
	require.True(t, d.paused, "dispatcher should be paused after Pause() call")

	// Calling Pause again should not panic or close an already closed channel
	d.Pause()
	require.True(t, d.paused, "dispatcher should remain paused after second Pause() call")
}

func TestMessageDispatcher_Resume(t *testing.T) {
	d := &MessageDispatcher{
		paused:   true,
		pauseCh:  make(chan struct{}),
		resumeCh: make(chan struct{}),
	}

	require.True(t, d.paused, "dispatcher should be paused initially")

	d.Resume()
	require.False(t, d.paused, "dispatcher should not be paused after Resume() call")

	// Calling Resume again should not panic or change state
	d.Resume()
	require.False(t, d.paused, "dispatcher should remain unpaused after second Resume() call")
}

func TestMessageDispatcher_processUnsentMessages(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	t.Run("success - should send the message, mark it as sent and cache to redis", func(tt *testing.T) {
		mockDB := NewMockDBInterface(ctrl)
		mockClient := somethirdparty.NewMockClientWithResponsesInterface(ctrl)
		mockRedis := NewMockRedisCache(ctrl)

		msg := api.Message{
			Id: 123,
		}

		mockDB.EXPECT().FetchUnsentMessages(gomock.Any()).Return([]api.Message{msg}, nil)
		mockClient.EXPECT().SendMessageWithResponse(gomock.Any(), somethirdparty.Message{}).Return(
			&somethirdparty.SendMessageResponse{
				JSON202: &somethirdparty.APIResponse{},
			},
			nil,
		)
		mockDB.EXPECT().MarkMessageAsSent(msg.Id, gomock.Any()).Return(nil)

		cmd := redis.NewStatusCmd(context.Background())
		cmd.SetVal("OK")
		mockRedis.EXPECT().Set(gomock.Any(), "sent_message:123", gomock.Any(), time.Duration(0)).Return(cmd)

		d := &MessageDispatcher{
			DB:     mockDB,
			Client: mockClient,
			Redis:  mockRedis,
		}
		d.processUnsentMessages()
	})

	t.Run("error - should mark message as invalid if message is too long", func(tt *testing.T) {
		mockDB := NewMockDBInterface(ctrl)

		msg := api.Message{
			Content: string(make([]byte, 161)),
		}
		mockDB.EXPECT().FetchUnsentMessages(gomock.Any()).Return([]api.Message{msg}, nil)
		mockDB.EXPECT().MarkMessageAsInvalid(msg.Id).Return(nil)

		d := &MessageDispatcher{
			DB: mockDB,
		}
		d.processUnsentMessages()
	})

	t.Run("error - should not call return if DB fetch fails", func(tt *testing.T) {
		mockDB := NewMockDBInterface(ctrl)
		mockClient := somethirdparty.NewMockClientWithResponsesInterface(ctrl)

		mockDB.EXPECT().FetchUnsentMessages(gomock.Any()).Return(nil, fmt.Errorf("dummy error"))
		mockClient.EXPECT().SendMessageWithResponse(gomock.Any(), gomock.Any()).Return(
			nil,
			nil,
		).Times(0) // should not be called

		d := &MessageDispatcher{
			DB: mockDB,
		}

		d.processUnsentMessages()
	})
}
