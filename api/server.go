package api

import (
	"encoding/json"
	"net/http"
)

type Server struct {
	DB           DBInterface
	ResumePauser ResumePauser
}

//go:generate go tool mockgen --package=api --destination=mock_resume_pauser.go . ResumePauser
type ResumePauser interface {
	Resume()
	Pause()
}

//go:generate go tool mockgen --package=api --destination=mock_db_interface.go . DBInterface
type DBInterface interface {
	GetSentMessages() ([]Message, error)
}

func NewServer(database DBInterface, resumePauser ResumePauser) Server {
	return Server{
		DB:           database,
		ResumePauser: resumePauser,
	}
}

func (s Server) ChangeState(w http.ResponseWriter, r *http.Request, params ChangeStateParams) {
	switch params.Action {
	case Pause:
		s.ResumePauser.Pause()
	case Resume:
		s.ResumePauser.Resume()
	default:
		http.Error(w, "invalid action", http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(State{Running: params.Action == Resume})
}

// GetSentMessages returns all sent messages
func (s Server) GetSentMessages(w http.ResponseWriter, r *http.Request) {
	msgs, err := s.DB.GetSentMessages()
	if err != nil {
		http.Error(w, "failed to fetch sent messages", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(msgs)
}
