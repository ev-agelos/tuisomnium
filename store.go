package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

type NameValue struct {
	Name  string
	Value string
}
type dbAuthType struct {
	Type   string
	Fields []*NameValue
}
type dbAuth struct {
	Types    []dbAuthType
	Selected int
}

type dbBody struct {
	Types    []*NameValue
	Selected int
}

type DBRequest struct {
	ID        int64
	Name      string
	Method    string
	Url       string
	Body      dbBody
	Auth      dbAuth
	Query     []NameValue
	Headers   []NameValue
	Responses []DBResponse
}
type DBResponse struct {
	ID            int64
	RequestID     int64
	RequestMethod string
	RequestUrl    string
	Body          string
	Status        string
	Headers       []NameValue
	Cookies       []NameValue
	Duration      string
	Size          string
	ResponseAt    time.Time
	TraceLogs     string
}

type Store struct {
	conn *sql.DB
}

func (s *Store) Init() error {
	var err error
	s.conn, err = sql.Open("sqlite3", "./requests.db")
	if err != nil {
		return err
	}

	requests := `CREATE TABLE IF NOT EXISTS requests (
		id integer not null primary key autoincrement,
        name text not null,
		method text not null,
		url text,
        body text,
		auth text,
		query text,
		headers text
	);`

	if _, err := s.conn.Exec(requests); err != nil {
		return err
	}

	responses := `CREATE TABLE IF NOT EXISTS responses (
        id integer not null primary key autoincrement,
        request_id integer,
        request_method text,
        request_url text,
		body text,
        status text,
        headers text,
        cookies text,
        duration text,
        size text,
        response_at integer,
        trace_logs text,
        FOREIGN KEY (request_id) REFERENCES names (id) ON DELETE CASCADE ON UPDATE CASCADE
    );`

	if _, err := s.conn.Exec(responses); err != nil {
		return err
	}

	return nil
}

func (s *Store) GetRequests() ([]DBRequest, error) {
	requestRows, err := s.conn.Query("SELECT * FROM requests")
	if err != nil {
		return nil, fmt.Errorf("failed to query requests: %w", err)
	}
	defer requestRows.Close()

	requestsMap := map[int64]*DBRequest{}
	for requestRows.Next() {
		var (
			r           DBRequest
			bodyJSON    string
			authJSON    string
			queryJSON   string
			headersJSON string
		)
		err := requestRows.Scan(&r.ID, &r.Name, &r.Method, &r.Url, &bodyJSON, &authJSON, &queryJSON, &headersJSON)
		if err != nil {
			return nil, fmt.Errorf("failed to scan requests: %w", err)
		}
		if err := json.Unmarshal([]byte(bodyJSON), &r.Body); err != nil {
			return nil, fmt.Errorf("failed to parse Body: %w", err)
		}
		if err := json.Unmarshal([]byte(authJSON), &r.Auth); err != nil {
			return nil, fmt.Errorf("failed to parse Auth: %w", err)
		}
		if err := json.Unmarshal([]byte(queryJSON), &r.Query); err != nil {
			return nil, fmt.Errorf("failed to parse Query: %w", err)
		}
		if err := json.Unmarshal([]byte(headersJSON), &r.Headers); err != nil {
			return nil, fmt.Errorf("failed to parse Headers: %w", err)
		}
		requestsMap[r.ID] = &r
	}

	responseRows, err := s.conn.Query("SELECT * FROM responses ORDER BY response_at DESC")
	if err != nil {
		return nil, fmt.Errorf("failed to query responses: %w", err)
	}
	defer responseRows.Close()

	for responseRows.Next() {
		var (
			r           DBResponse
			HeadersJSON string
			CookiesJSON string
			unixTime    int64
		)
		err := responseRows.Scan(
			&r.ID, &r.RequestID, &r.RequestMethod, &r.RequestUrl,
			&r.Body, &r.Status, &HeadersJSON,
			&CookiesJSON, &r.Duration, &r.Size,
			&unixTime, &r.TraceLogs,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan response: %w", err)
		}
		if err := json.Unmarshal([]byte(HeadersJSON), &r.Headers); err != nil {
			return nil, fmt.Errorf("failed to parse Response Headers: %w", err)
		}
		if err := json.Unmarshal([]byte(CookiesJSON), &r.Cookies); err != nil {
			return nil, fmt.Errorf("failed to parse Response Cookies: %w", err)
		}

		r.ResponseAt = time.UnixMilli(unixTime)
		request := requestsMap[r.RequestID]
		request.Responses = append(request.Responses, r)
	}

	result := []DBRequest{}
	for _, name := range requestsMap {
		result = append(result, *name)
	}

	return result, nil
}

func (s *Store) SaveRequest(r *DBRequest) error {
	bodyJSON, err := json.Marshal(r.Body)
	if err != nil {
		return fmt.Errorf("failed to serialize Body: %w", err)
	}
	authJSON, err := json.Marshal(r.Auth)
	if err != nil {
		return fmt.Errorf("failed to serialize Auth: %w", err)
	}
	queryJSON, err := json.Marshal(r.Query)
	if err != nil {
		return fmt.Errorf("failed to serialize Query: %w", err)
	}
	headersJSON, err := json.Marshal(r.Headers)
	if err != nil {
		return fmt.Errorf("failed to serialize Headers: %w", err)
	}
	if r.ID == 0 {
		requestQuery := `INSERT INTO requests (name, method, url, body, auth, query, headers) VALUES (?, ?, ?, ?, ?, ?, ?);`
		result, err := s.conn.Exec(requestQuery, r.Name, r.Method, r.Url, string(bodyJSON), string(authJSON), string(queryJSON), string(headersJSON))
		if err != nil {
			return err
		}
		r.ID, err = result.LastInsertId()
		if err != nil {
			return err
		}
	} else {
		requestQuery := `INSERT INTO requests (id, name, method, url, body, auth, query, headers)
        VALUES (?, ?, ?, ?, ?, ?, ?, ?)
        ON CONFLICT(id) DO UPDATE
        SET name=excluded.name, method=excluded.method, url=excluded.url, body=excluded.body, auth=excluded.auth,
        query=excluded.query, headers=excluded.headers;`

		if _, err := s.conn.Exec(requestQuery, r.ID, r.Name, r.Method, r.Url, string(bodyJSON), string(authJSON), string(queryJSON), string(headersJSON)); err != nil {
			return err
		}
	}
	return nil
}

func (s *Store) SaveResponse(r *DBResponse) error {
	headersJSON, err := json.Marshal(r.Headers)
	if err != nil {
		return fmt.Errorf("failed to serialize Response Headers: %w", err)
	}
	cookiesJSON, err := json.Marshal(r.Cookies)
	if err != nil {
		return fmt.Errorf("failed to serialize Response Cookies: %w", err)
	}

	responseQuery := `INSERT INTO responses (request_id, request_method, request_url, body, status, headers, cookies, duration, size, response_at, trace_logs)
    VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?);`

	result, err := s.conn.Exec(
		responseQuery, r.RequestID, r.RequestMethod, r.RequestUrl,
		r.Body, r.Status, string(headersJSON), string(cookiesJSON),
		r.Duration, r.Size, r.ResponseAt.UnixMilli(), r.TraceLogs,
	)
	if err != nil {
		return err
	}
	r.ID, err = result.LastInsertId()
	if err != nil {
		return err
	}

	return nil
}

func (s *Store) DeleteRequest(r *DBRequest) error {
	query := `DELETE FROM requests WHERE ID=$ID;`
	if _, err := s.conn.Exec(query, r.ID); err != nil {
		return err
	}
	return nil
}

// func jsonToMarshal(j NameValue) (string, error) {
// 	m, err := json.Marshal(j)
// 	if err != nil {
// 		return "", fmt.Errorf("failed to serialize: %w", err)
// 	}
// 	return string(m), nil
// }
