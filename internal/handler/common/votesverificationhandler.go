package handler

import (
	"net/http"

	"datacenter/internal/svc"
)

func VotesVerificationHandler(ctx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		w.Write([]byte("NT04cqknJe0em3mT"))
		return
	}
}
