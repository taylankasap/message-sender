package main

import (
	"testing"
	"time"

	"github.com/taylankasap/message-sender/model"
	somethirdparty "github.com/taylankasap/message-sender/some_third_party"
	"go.uber.org/mock/gomock"
)

func TestMessageDispatcher_Start(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockDB := NewMockDBInterface(ctrl)
	mockClient := somethirdparty.NewMockClientWithResponsesInterface(ctrl)

	t.Run("it should start dispatching the first batch of messages without waiting the period", func(tt *testing.T) {
		fakeMsg := model.Message{
			ID:        1,
			Content:   "Hello",
			Recipient: "+1234567890",
			Status:    model.StatusUnsent,
		}

		// these should be called once
		mockDB.EXPECT().FetchUnsentMessages(1).Return([]model.Message{fakeMsg}, nil).Times(1)
		mockClient.EXPECT().SendMessageWithResponse(gomock.Any(), gomock.Any(), gomock.Any()).Return(
			&somethirdparty.SendMessageResponse{JSON202: &somethirdparty.APIResponse{MessageId: "dummy-message-id"}},
			nil,
		).Times(1)
		mockDB.EXPECT().MarkMessageAsSent(fakeMsg.ID, gomock.Any()).Return(nil).Times(1)

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
