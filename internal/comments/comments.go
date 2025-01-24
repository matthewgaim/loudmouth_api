package comments

import (
	"database/sql"
	"log"
	"time"
)

func MakeComment(message string, poster string, time_of_media int, media_id string, db *sql.DB) (int, time.Time, error) {
	var db_media_id int
	res := db.QueryRow(`SELECT id FROM media WHERE media_host_id = $1`, media_id)
	if err := res.Scan(&db_media_id); err != nil {
		log.Printf("Error finding Media ID on DB: %s\n", err.Error())
		return 0, time.Time{}, err
	}

	var id int
	var created_at time.Time
	row := db.QueryRow(
		`INSERT INTO comments
                (media_id, time_of_media, poster, message) 
            VALUES
                ($1, $2, $3, $4)
			RETURNING id, created_at`, db_media_id, time_of_media, poster, message)
	if err := row.Scan(&id, &created_at); err != nil {
		log.Printf("Error scanning new comment ID: %s\n", err.Error())
		return 0, time.Time{}, err
	}
	return id, created_at, nil
}

func GetComments(mediaId string, minTimeOfMedia int, maxTimeOfMedia int, db *sql.DB) ([]OutgoingComments, error) {
	var id int
	log.Printf("MediaID: %s, MinTime: %d, MaxTime: %d\n", mediaId, minTimeOfMedia, maxTimeOfMedia)
	res := db.QueryRow(`SELECT id FROM media WHERE media_host_id = $1`, mediaId)
	if err := res.Scan(&id); err != nil {
		log.Printf("Error finding Media ID on DB: %s\n", err.Error())
		return nil, err
	}
	var results []OutgoingComments
	rows, err := db.Query(
		`SELECT * FROM comments
		WHERE media_id = $1
		AND time_of_media BETWEEN $2 AND $3`,
		id, minTimeOfMedia, maxTimeOfMedia)
	if err != nil {
		log.Printf("Error querying comments: %s\n", err.Error())
		return nil, err
	}

	defer rows.Close()
	for rows.Next() {
		row := new(OutgoingComments)
		if err := rows.Scan(&row.Id, &row.TimeOfMedia, &row.MediaId, &row.Poster, &row.Message, &row.CreatedAt); err != nil {
			log.Println(err)
			return nil, err
		}
		results = append(results, *row)
	}
	return results, nil
}
