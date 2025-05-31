package main

import (
	"encoding/json"
	"net/http"
)

type Server struct {
	resumePauser ResumePauser
}

//go:generate go tool mockgen --package=main --destination=mock_resume_pauser.go . ResumePauser
type ResumePauser interface {
	Resume()
	Pause()
}

func NewServer(resumePauser ResumePauser) Server {
	return Server{
		resumePauser: resumePauser,
	}
}

func (s Server) ChangeState(w http.ResponseWriter, r *http.Request, params ChangeStateParams) {
	switch params.Action {
	case Pause:
		s.resumePauser.Pause()
	case Resume:
		s.resumePauser.Resume()
	default:
		http.Error(w, "invalid action", http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(State{Running: params.Action == Resume})
}
