package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"
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
