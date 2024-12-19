package viewscount

import (
	"database/sql"
	"fmt"
	"github.com/jackc/pgx/v5/pgxpool"
	"net"
	"net/http"
	"strings"
	"sync"
	"time"
)

type DBExecutor interface {
	Exec(query string, args ...interface{}) (sql.Result, error)
}

type ViewTracker struct {
	mutex       sync.RWMutex
	viewHistory map[string][]ViewRequest
	db          DBExecutor
	minutesGap  time.Duration
}

type ViewRequest struct {
	IP        string
	UserAgent string
	ObjectID  int
	TableName string
	Timestamp time.Time
}

func NewViewTracker(db interface{}, minutesGap time.Duration) *ViewTracker {
	var executor DBExecutor

	switch d := db.(type) {
	case *sql.DB:
		executor = d
	case *pgxpool.Pool:
		executor = NewPgxAdapter(d)
	default:
		panic("unsupported database type")
	}

	return &ViewTracker{
		viewHistory: make(map[string][]ViewRequest),
		db:          executor,
		minutesGap:  minutesGap,
	}
}

func extractIP(r *http.Request) string {

	forwardedFor := r.Header.Get("X-Forwarded-For")
	if forwardedFor != "" {

		ips := strings.Split(forwardedFor, ",")
		return strings.TrimSpace(ips[0])
	}

	realIP := r.Header.Get("X-Real-IP")
	if realIP != "" {
		return realIP
	}

	ip, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}
	return ip
}

func (vt *ViewTracker) IncrementViewCount(tableName string, id int) error {
	query := fmt.Sprintf("UPDATE %s SET view_count = view_count + 1 WHERE id = $1", tableName)
	_, err := vt.db.Exec(query, id)
	return err
}

func (vt *ViewTracker) cleanOldRecords() {
	vt.mutex.Lock()
	defer vt.mutex.Unlock()

	cutoff := time.Now().Add(-vt.minutesGap * time.Minute)
	for key, requests := range vt.viewHistory {
		var newRequests []ViewRequest
		for _, req := range requests {
			if req.Timestamp.After(cutoff) {
				newRequests = append(newRequests, req)
			}
		}
		if len(newRequests) == 0 {
			delete(vt.viewHistory, key)
		} else {
			vt.viewHistory[key] = newRequests
		}
	}
}

func (vt *ViewTracker) IsOrganicView(req ViewRequest) bool {
	vt.mutex.RLock()
	defer vt.mutex.RUnlock()

	key := fmt.Sprintf("%s:%s:%d:%s", req.IP, req.UserAgent, req.ObjectID, req.TableName)
	requests := vt.viewHistory[key]

	if len(requests) == 0 {
		return true
	}

	lastRequest := requests[len(requests)-1]
	return time.Since(lastRequest.Timestamp) > vt.minutesGap*time.Minute
}

func (vt *ViewTracker) TrackView(req ViewRequest) {
	vt.mutex.Lock()
	defer vt.mutex.Unlock()

	key := fmt.Sprintf("%s:%s:%d:%s", req.IP, req.UserAgent, req.ObjectID, req.TableName)
	vt.viewHistory[key] = append(vt.viewHistory[key], req)
}

func (vt *ViewTracker) HandleView(r *http.Request, tableName string, id int) error {
	req := ViewRequest{
		IP:        extractIP(r),
		UserAgent: r.UserAgent(),
		ObjectID:  id,
		TableName: tableName,
		Timestamp: time.Now(),
	}

	go vt.cleanOldRecords()

	if vt.IsOrganicView(req) {
		if err := vt.IncrementViewCount(tableName, id); err != nil {
			return fmt.Errorf("failed to increment view count: %w", err)
		}
	}

	vt.TrackView(req)
	return nil
}
