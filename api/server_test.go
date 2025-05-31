package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"github.com/taylankasap/message-sender/model"
	"go.uber.org/mock/gomock"
)

func TestServer_ChangeState(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	t.Run("success - should call Pause and return 200", func(t *testing.T) {
		mockResumePauser := NewMockResumePauser(ctrl)
		mockResumePauser.EXPECT().Pause()

		s := Server{resumePauser: mockResumePauser}
		r := httptest.NewRequest("GET", "/change-state?action=pause", nil)
		w := httptest.NewRecorder()
		s.ChangeState(w, r, ChangeStateParams{Action: Pause})
		require.Equal(t, http.StatusOK, w.Code)
		var resp State
		require.NoError(t, json.NewDecoder(w.Body).Decode(&resp))
		require.False(t, resp.Running)
	})

	t.Run("success - should call Resume and return 200", func(t *testing.T) {
		mockResumePauser := NewMockResumePauser(ctrl)
		mockResumePauser.EXPECT().Resume()
		s := Server{resumePauser: mockResumePauser}
		r := httptest.NewRequest("GET", "/change-state?action=resume", nil)
		w := httptest.NewRecorder()
		s.ChangeState(w, r, ChangeStateParams{Action: Resume})
		require.Equal(t, http.StatusOK, w.Code)
		var resp State
		require.NoError(t, json.NewDecoder(w.Body).Decode(&resp))
		require.True(t, resp.Running)
	})

	t.Run("error - should return 400 for invalid action", func(t *testing.T) {
		mockResumePauser := NewMockResumePauser(ctrl)
		s := Server{resumePauser: mockResumePauser}
		r := httptest.NewRequest("GET", "/change-state?action=invalid", nil)
		w := httptest.NewRecorder()
		s.ChangeState(w, r, ChangeStateParams{Action: "invalid"})
		require.Equal(t, http.StatusBadRequest, w.Code)
	})
}

func TestServer_GetSentMessages(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	t.Run("success - should return sent messages", func(tt *testing.T) {
		mockDB := NewMockDBInterface(ctrl)
		s := Server{DB: mockDB}

		t1 := "2025-05-31T10:00:00Z"
		t1Parsed, err := time.Parse(time.RFC3339, t1)
		require.NoError(tt, err)

		expectedMessages := []model.Message{
			{ID: 1, Content: "Hello!", Recipient: "+1234567890", Status: "sent", SentAt: &t1Parsed},
			{ID: 2, Content: "World!", Recipient: "+9876543210", Status: "sent", SentAt: &t1Parsed},
		}

		mockDB.EXPECT().FetchSentMessages().Return(expectedMessages, nil)

		r := httptest.NewRequest("GET", "/sent-messages", nil)
		w := httptest.NewRecorder()
		s.GetSentMessages(w, r)

		require.Equal(tt, http.StatusOK, w.Code)

		var actualMessages []model.Message
		require.NoError(tt, json.NewDecoder(w.Body).Decode(&actualMessages))
		require.Len(tt, actualMessages, 2)
		require.Equal(tt, expectedMessages[0].ID, actualMessages[0].ID)
		require.Equal(tt, expectedMessages[1].Content, actualMessages[1].Content)
	})

	t.Run("success - should return empty array if there are no messages", func(tt *testing.T) {
		mockDB := NewMockDBInterface(ctrl)
		s := Server{DB: mockDB}

		expectedMessages := []model.Message{}

		mockDB.EXPECT().FetchSentMessages().Return(expectedMessages, nil)

		r := httptest.NewRequest("GET", "/sent-messages", nil)
		w := httptest.NewRecorder()
		s.GetSentMessages(w, r)

		require.Equal(tt, http.StatusOK, w.Code)

		var actualMessages []model.Message
		require.NoError(tt, json.NewDecoder(w.Body).Decode(&actualMessages))
		require.Equal(tt, expectedMessages, actualMessages)
	})

	t.Run("error - should return 500 on DB error", func(tt *testing.T) {
		mockDB := NewMockDBInterface(ctrl)
		s := Server{DB: mockDB}

		mockDB.EXPECT().FetchSentMessages().Return(nil, fmt.Errorf("dummy error"))

		r := httptest.NewRequest("GET", "/sent-messages", nil)
		w := httptest.NewRecorder()
		s.GetSentMessages(w, r)

		require.Equal(tt, http.StatusInternalServerError, w.Code)
	})
}
