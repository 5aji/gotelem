package api
// This file contains common behaviors that are used across various requests
import (
	"net/http"
	"strconv"
	"time"

	"github.com/kschamplin/gotelem/internal/db"
)

func extractBusEventFilter(r *http.Request) (*db.BusEventFilter, error) {

	bef := &db.BusEventFilter{}

	v := r.URL.Query()
	bef.Names = v["name"] // put all the names in.
	if el := v.Get("start"); el != "" {
		// parse the start time query.
		t, err := time.Parse(time.RFC3339, el)
		if err != nil {
			return bef, err
		}
		bef.TimerangeStart = t
	}
	if el := v.Get("end"); el != "" {
		// parse the start time query.
		t, err := time.Parse(time.RFC3339, el)
		if err != nil {
			return bef, err
		}
		bef.TimerangeStart = t
	}
	return bef, nil
}

func extractLimitModifier(r *http.Request) (*db.LimitOffsetModifier, error) {
	lim := &db.LimitOffsetModifier{}
	v := r.URL.Query()
	if el := v.Get("limit"); el != "" {
		val, err := strconv.ParseInt(el, 10, 64)
		if err != nil {
			return nil, err
		}
		lim.Limit = int(val)
		// next, we check if we have an offset.
		// we only check offset if we also have a limit.
		// offset without limit isn't valid and is ignored.
		if el := v.Get("offset"); el != "" {
			val, err := strconv.ParseInt(el, 10, 64)
			if err != nil {
				return nil, err
			}
			lim.Offset = int(val)
		}
		return lim, nil
	}
	// we use the nil case to indicate that no limit was provided.
	return nil, nil
}
