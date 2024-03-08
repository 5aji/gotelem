package gotelem

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"reflect"
	"testing"
	"time"

	"github.com/kschamplin/gotelem/skylab"
)

func Test_extractBusEventFilter(t *testing.T) {
	makeReq := func(path string) *http.Request {
		return httptest.NewRequest(http.MethodGet, path, nil)
	}
	tests := []struct {
		name    string
		req     *http.Request
		want    *BusEventFilter
		wantErr bool
	}{
		{
			name:    "test no extractions",
			req:     makeReq("http://localhost/"),
			want:    &BusEventFilter{},
			wantErr: false,
		},
		{
			name: "test single name extract",
			req:  makeReq("http://localhost/?name=hi"),
			want: &BusEventFilter{
				Names: []string{"hi"},
			},
			wantErr: false,
		},
		{
			name: "test multi name extract",
			req:  makeReq("http://localhost/?name=hi1&name=hi2"),
			want: &BusEventFilter{
				Names: []string{"hi1", "hi2"},
			},
			wantErr: false,
		},
		{
			name: "test start time valid extract",
			req:  makeReq(fmt.Sprintf("http://localhost/?start=%s", url.QueryEscape(time.Unix(160000000, 0).Format(time.RFC3339)))),
			want: &BusEventFilter{
				StartTime: time.Unix(160000000, 0),
			},
			wantErr: false,
		},
		// {
		// 	name: "test start time invalid extract",
		// 	req:  makeReq(fmt.Sprintf("http://localhost/?start=%s", url.QueryEscape("ajlaskdj"))),
		// 	wantErr: true,
		// },
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Logf("Testing URL %s", tt.req.URL.String())
			got, err := extractBusEventFilter(tt.req)
			if (err != nil) != tt.wantErr {
				t.Errorf("extractBusEventFilter() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			// we have to manually compare fields because timestamps can't be deeply compared.
			if !reflect.DeepEqual(got.Names, tt.want.Names) {
				t.Errorf("extractBusEventFilter() Names bad = %v, want %v", got.Names, tt.want.Names)
			}
			if !reflect.DeepEqual(got.Indexes, tt.want.Indexes) {
				t.Errorf("extractBusEventFilter() Indexes bad = %v, want %v", got.Indexes, tt.want.Indexes)
			}
			if !got.StartTime.Equal(tt.want.StartTime) {
				t.Errorf("extractBusEventFilter() StartTime mismatch = %v, want %v", got.StartTime, tt.want.StartTime)
			}
			if !got.EndTime.Equal(tt.want.EndTime) {
				t.Errorf("extractBusEventFilter() EndTime mismatch = %v, want %v", got.EndTime, tt.want.EndTime)
			}
		})
	}
}

func Test_extractLimitModifier(t *testing.T) {
	makeReq := func(path string) *http.Request {
		return httptest.NewRequest(http.MethodGet, path, nil)
	}
	tests := []struct {
		name    string
		req    *http.Request
		want    *LimitOffsetModifier
		wantErr bool
	}{
		{
			name: "test no limit/offset",
			req: makeReq("http://localhost/"),
			want: nil,
			wantErr: false,
		},
		{
			name: "test limit, no offset",
			req: makeReq("http://localhost/?limit=10"),
			want: &LimitOffsetModifier{Limit: 10},
			wantErr: false,
		},
		{
			name: "test limit and offset",
			req: makeReq("http://localhost/?limit=100&offset=200"),
			want: &LimitOffsetModifier{Limit: 100, Offset: 200},
			wantErr: false,
		},
		{
			name: "test only offset",
			req: makeReq("http://localhost/?&offset=200"),
			want: nil,
			wantErr: false,
		},
		{
			name: "test bad limit",
			req: makeReq("http://localhost/?limit=aaaa"),
			want: nil,
			wantErr: true,
		},
		{
			name: "test good limit, bad offset",
			req: makeReq("http://localhost/?limit=10&offset=jjjj"),
			want: nil,
			wantErr: true,
		},

	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := extractLimitModifier(tt.req)
			if (err != nil) != tt.wantErr {
				t.Errorf("extractLimitModifier() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("extractLimitModifier() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_ApiV1GetPackets(t *testing.T) {
	tdb := MakeMockDatabase(t.Name())
	SeedMockDatabase(tdb)
	evs := GetSeedEvents()
	handler := apiV1GetPackets(tdb)

	tests := []struct{
		name string
		req *http.Request
		statusCode int
		expectedResults []skylab.BusEvent
	}{
		{
			name: "get all packets test",
			req: httptest.NewRequest(http.MethodGet, "http://localhost/", nil),
			statusCode: http.StatusOK,
			expectedResults: evs,
		},
		{
			name: "filter name test",
			req: httptest.NewRequest(http.MethodGet, "http://localhost/?name=bms_module", nil),
			statusCode: http.StatusOK,
			expectedResults: func() []skylab.BusEvent {
				filtered := make([]skylab.BusEvent, 0)
				for _, pkt := range evs {
					if pkt.Name == "bms_module" {
						filtered = append(filtered, pkt)
					}
				}
				return filtered
			}(),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// construct the recorder
			w := httptest.NewRecorder()
			handler(w, tt.req)

			resp := w.Result()

			if tt.statusCode != resp.StatusCode {
				t.Errorf("incorrect status code: expected %d got %d", tt.statusCode, resp.StatusCode)
			}

			decoder := json.NewDecoder(resp.Body)
			var resultEvents []skylab.BusEvent
			err := decoder.Decode(&resultEvents)
			if err != nil {
				t.Fatalf("could not parse JSON response: %v", err)
			}

			if len(resultEvents) != len(tt.expectedResults) {
				t.Fatalf("response length did not match, want %d got %d", len(tt.expectedResults), len(resultEvents))
			}

			for idx := range tt.expectedResults {
				expected := tt.expectedResults[idx]
				actual := resultEvents[idx]
				if !expected.Equals(&actual) {
					t.Errorf("packet did not match, want %v got %v", expected, actual)
				}
			}

		})
	}
}
