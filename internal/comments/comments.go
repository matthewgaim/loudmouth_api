package comments

import (
	"database/sql"
	"encoding/json"
	"log"
	"net/http"

	"github.com/matthewgaim/loudmouth_api/internal/errors"
)

func MakeComment(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req IncomingCommentRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			log.Println(err)
			errors.RespondWithError(w, http.StatusBadRequest, "Invalid body")
			return
		}

		var id int
		res := db.QueryRow(`SELECT id FROM media WHERE media_host_id = $1`, req.MediaId)
		if err := res.Scan(&id); err != nil {
			log.Println(err)
			errors.RespondWithError(w, http.StatusBadRequest, "Error finding media in database")
			return
		}

		_, err := db.Exec(
			`INSERT INTO comments
                (media_id, time_of_media, poster, message) 
            VALUES
                ($1, $2, $3, $4)`, id, req.TimeOfMedia, req.Poster, req.Message)
		if err != nil {
			log.Println(err)
			errors.RespondWithError(w, http.StatusInternalServerError, "Error querying database")
			return
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{"message": "ok"})
	}
}

func GetComments(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req GetCommentsRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			log.Printf("Error in GetComments():\n%s", err.Error())
			errors.RespondWithError(w, http.StatusBadRequest, "Invalid body")
			return
		}

		var id int
		res := db.QueryRow(`SELECT id FROM media WHERE media_host_id = $1`, req.MediaId)
		if err := res.Scan(&id); err != nil {
			log.Println(err)
			errors.RespondWithError(w, http.StatusBadRequest, "Error finding media in database")
			return
		}

		var results []OutgoingComments
		rows, err := db.Query(`SELECT * FROM comments WHERE (media_id, time_of_media) = ($1, $2)`, id, req.TimeOfMedia)
		if err != nil {
			log.Println(err)
			errors.RespondWithError(w, http.StatusInternalServerError, "Error querying database")
			return
		}

		defer rows.Close()
		for rows.Next() {
			row := new(OutgoingComments)
			if err := rows.Scan(&row.Id, &row.TimeOfMedia, &row.MediaId, &row.Poster, &row.Message, &row.CreatedAt); err != nil {
				log.Println(err)
				errors.RespondWithError(w, http.StatusInternalServerError, "Server error")
				return
			}
			results = append(results, *row)
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string][]OutgoingComments{"results": results})
	}
}
