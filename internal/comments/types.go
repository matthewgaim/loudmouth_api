package comments

import "time"

type GetCommentsRequest struct {
	MediaId     string `json:"media_id"`
	TimeOfMedia int    `json:"time_of_media"`
}

type IncomingCommentRequest struct {
	MediaId     string `json:"media_id"`
	TimeOfMedia int    `json:"time_of_media"`
	Message     string `json:"message"`
	Poster      int    `json:"poster"`
}

type OutgoingComments struct {
	Id          int       `json:"id"`
	TimeOfMedia int       `json:"time_of_media"`
	MediaId     int       `json:"media_id"`
	Poster      int       `json:"poster"`
	Message     string    `json:"message"`
	CreatedAt   time.Time `json:"created_at"`
}
