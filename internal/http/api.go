package http

import (
	"net/http"
)

func GetUserFeeds() {

}

func getUserFeedsByTimeRange(w http.ResponseWriter, r *http.Request) {
	//cutoff := time.Now().UTC().Add(-24 * time.Hour)
	//const q = `
	//    SELECT id, created_at
	//    FROM   records
	//    WHERE  created_at >= $1
	//    ORDER BY created_at DESC;
	//`
	//rows, err := db.QueryContext(ctx, q, cutoff)
	//if err != nil {
	//	return nil, err
	//}
	//defer rows.Close()
	//
	//var out []Record
	//for rows.Next() {
	//	var r Record
	//	if err := rows.Scan(&r.ID, &r.CreatedAt); err != nil {
	//		return nil, err
	//	}
	//	out = append(out, r)
	//}
	//return out, rows.Err()
}
