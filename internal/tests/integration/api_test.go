package integration

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTeamAdd_Success(t *testing.T) {
	ts, err := SetupTestServer(t)
	require.NoError(t, err)
	defer ts.Close()

	reqBody := map[string]interface{}{
		"team": map[string]interface{}{
			"team_name": "backend",
			"members": []map[string]interface{}{
				{"user_id": "u1", "username": "Alice", "is_active": true},
				{"user_id": "u2", "username": "Bob", "is_active": true},
			},
		},
	}

	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest("POST", "/team/add", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	ts.Server.Handler.ServeHTTP(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)

	var response map[string]interface{}
	err = json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	team, ok := response["team"].(map[string]interface{})
	require.True(t, ok)
	assert.Equal(t, "backend", team["team_name"])
}

func TestTeamAdd_DuplicateTeam(t *testing.T) {
	ts, err := SetupTestServer(t)
	require.NoError(t, err)
	defer ts.Close()

	ctx := context.Background()
	_, err = ts.Storage.Db.Exec(ctx, `INSERT INTO team (name) VALUES ('backend')`)
	require.NoError(t, err)

	reqBody := map[string]interface{}{
		"team": map[string]interface{}{
			"team_name": "backend",
			"members": []map[string]interface{}{
				{"user_id": "u1", "username": "Alice", "is_active": true},
			},
		},
	}

	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest("POST", "/team/add", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	ts.Server.Handler.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)

	var response map[string]interface{}
	err = json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	errorObj, ok := response["error"].(map[string]interface{})
	require.True(t, ok)
	assert.Equal(t, "TEAM_EXISTS", errorObj["code"])
}

func TestTeamGet_Success(t *testing.T) {
	ts, err := SetupTestServer(t)
	require.NoError(t, err)
	defer ts.Close()

	ctx := context.Background()
	_, err = ts.Storage.Db.Exec(ctx, `
		INSERT INTO team (name) VALUES ('backend');
		INSERT INTO users (user_id, username, team_name, is_active) VALUES 
			('u1', 'Alice', 'backend', true),
			('u2', 'Bob', 'backend', true);
	`)
	require.NoError(t, err)

	req := httptest.NewRequest("GET", "/team/get?team_name=backend", nil)
	w := httptest.NewRecorder()

	ts.Server.Handler.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err = json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.Equal(t, "backend", response["team_name"])
	members, ok := response["members"].([]interface{})
	require.True(t, ok)
	assert.Len(t, members, 2)
}

func TestTeamGet_NotFound(t *testing.T) {
	ts, err := SetupTestServer(t)
	require.NoError(t, err)
	defer ts.Close()

	req := httptest.NewRequest("GET", "/team/get?team_name=nonexistent", nil)
	w := httptest.NewRecorder()

	ts.Server.Handler.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)

	var response map[string]interface{}
	err = json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	errorObj, ok := response["error"].(map[string]interface{})
	require.True(t, ok)
	assert.Equal(t, "NOT_FOUND", errorObj["code"])
}

func TestUsersSetIsActive_Success(t *testing.T) {
	ts, err := SetupTestServer(t)
	require.NoError(t, err)
	defer ts.Close()

	ctx := context.Background()
	_, err = ts.Storage.Db.Exec(ctx, `
		INSERT INTO team (name) VALUES ('backend');
		INSERT INTO users (user_id, username, team_name, is_active) VALUES 
			('u1', 'Alice', 'backend', true);
	`)
	require.NoError(t, err)

	reqBody := map[string]interface{}{
		"user_id":   "u1",
		"is_active": false,
	}

	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest("POST", "/users/setIsActive", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	ts.Server.Handler.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err = json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	user, ok := response["user"].(map[string]interface{})
	require.True(t, ok)
	assert.Equal(t, false, user["is_active"])
}

func TestUsersSetIsActive_NotFound(t *testing.T) {
	ts, err := SetupTestServer(t)
	require.NoError(t, err)
	defer ts.Close()

	reqBody := map[string]interface{}{
		"user_id":   "nonexistent",
		"is_active": false,
	}

	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest("POST", "/users/setIsActive", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	ts.Server.Handler.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)

	var response map[string]interface{}
	err = json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	errorObj, ok := response["error"].(map[string]interface{})
	require.True(t, ok)
	assert.Equal(t, "NOT_FOUND", errorObj["code"])
}

func TestUsersGetReview_Success(t *testing.T) {
	ts, err := SetupTestServer(t)
	require.NoError(t, err)
	defer ts.Close()

	ctx := context.Background()
	_, err = ts.Storage.Db.Exec(ctx, `
		INSERT INTO team (name) VALUES ('backend');
		INSERT INTO users (user_id, username, team_name, is_active) VALUES 
			('u1', 'Alice', 'backend', true),
			('u2', 'Bob', 'backend', true);
		INSERT INTO pull_requests (pull_request_id, pull_request_name, author_id, status) VALUES 
			('pr-1', 'PR 1', 'u1', 'OPEN');
		INSERT INTO pr_reviewers (pull_request_id, user_id) VALUES ('pr-1', 'u2');
	`)
	require.NoError(t, err)

	req := httptest.NewRequest("GET", "/users/getReview?user_id=u2", nil)
	w := httptest.NewRecorder()

	ts.Server.Handler.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err = json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.Equal(t, "u2", response["user_id"])
	pullRequests, ok := response["pull_requests"].([]interface{})
	require.True(t, ok)
	assert.Len(t, pullRequests, 1)
}

func TestUsersGetReview_UserNotFound(t *testing.T) {
	ts, err := SetupTestServer(t)
	require.NoError(t, err)
	defer ts.Close()

	req := httptest.NewRequest("GET", "/users/getReview?user_id=nonexistent", nil)
	w := httptest.NewRecorder()

	ts.Server.Handler.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestPullRequestCreate_Success(t *testing.T) {
	ts, err := SetupTestServer(t)
	require.NoError(t, err)
	defer ts.Close()

	ctx := context.Background()
	_, err = ts.Storage.Db.Exec(ctx, `
		INSERT INTO team (name) VALUES ('backend');
		INSERT INTO users (user_id, username, team_name, is_active) VALUES 
			('u1', 'Alice', 'backend', true),
			('u2', 'Bob', 'backend', true),
			('u3', 'Charlie', 'backend', true);
	`)
	require.NoError(t, err)

	reqBody := map[string]interface{}{
		"pull_request_id":   "pr-1",
		"pull_request_name": "Add feature",
		"author_id":         "u1",
	}

	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest("POST", "/pullRequest/create", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	ts.Server.Handler.ServeHTTP(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)

	var response map[string]interface{}
	err = json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	pr, ok := response["pr"].(map[string]interface{})
	require.True(t, ok)
	assert.Equal(t, "pr-1", pr["pull_request_id"])
	assert.Equal(t, "OPEN", pr["status"])

	reviewers, ok := pr["assigned_reviewers"].([]interface{})
	require.True(t, ok)
	assert.GreaterOrEqual(t, len(reviewers), 0)
	assert.LessOrEqual(t, len(reviewers), 2)
}

func TestPullRequestCreate_Duplicate(t *testing.T) {
	ts, err := SetupTestServer(t)
	require.NoError(t, err)
	defer ts.Close()

	ctx := context.Background()
	_, err = ts.Storage.Db.Exec(ctx, `
		INSERT INTO team (name) VALUES ('backend');
		INSERT INTO users (user_id, username, team_name, is_active) VALUES 
			('u1', 'Alice', 'backend', true);
		INSERT INTO pull_requests (pull_request_id, pull_request_name, author_id, status) VALUES 
			('pr-1', 'PR 1', 'u1', 'OPEN');
	`)
	require.NoError(t, err)

	reqBody := map[string]interface{}{
		"pull_request_id":   "pr-1",
		"pull_request_name": "Add feature",
		"author_id":         "u1",
	}

	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest("POST", "/pullRequest/create", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	ts.Server.Handler.ServeHTTP(w, req)

	assert.Equal(t, http.StatusConflict, w.Code)

	var response map[string]interface{}
	err = json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	errorObj, ok := response["error"].(map[string]interface{})
	require.True(t, ok)
	assert.Equal(t, "PR_EXISTS", errorObj["code"])
}

func TestPullRequestCreate_AuthorNotFound(t *testing.T) {
	ts, err := SetupTestServer(t)
	require.NoError(t, err)
	defer ts.Close()

	reqBody := map[string]interface{}{
		"pull_request_id":   "pr-1",
		"pull_request_name": "Add feature",
		"author_id":         "nonexistent",
	}

	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest("POST", "/pullRequest/create", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	ts.Server.Handler.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestPullRequestMerge_Success(t *testing.T) {
	ts, err := SetupTestServer(t)
	require.NoError(t, err)
	defer ts.Close()

	ctx := context.Background()
	_, err = ts.Storage.Db.Exec(ctx, `
		INSERT INTO team (name) VALUES ('backend');
		INSERT INTO users (user_id, username, team_name, is_active) VALUES 
			('u1', 'Alice', 'backend', true);
		INSERT INTO pull_requests (pull_request_id, pull_request_name, author_id, status) VALUES 
			('pr-1', 'PR 1', 'u1', 'OPEN');
	`)
	require.NoError(t, err)

	reqBody := map[string]interface{}{
		"pull_request_id": "pr-1",
	}

	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest("POST", "/pullRequest/merge", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	ts.Server.Handler.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err = json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	pr, ok := response["pr"].(map[string]interface{})
	require.True(t, ok)
	assert.Equal(t, "MERGED", pr["status"])
	assert.NotNil(t, pr["mergedAt"])
}

func TestPullRequestMerge_Idempotent(t *testing.T) {
	ts, err := SetupTestServer(t)
	require.NoError(t, err)
	defer ts.Close()

	ctx := context.Background()
	mergedAt := time.Now()
	_, err = ts.Storage.Db.Exec(ctx, `
		INSERT INTO team (name) VALUES ('backend');
	`)
	require.NoError(t, err)

	_, err = ts.Storage.Db.Exec(ctx, `
		INSERT INTO users (user_id, username, team_name, is_active) VALUES 
			('u1', 'Alice', 'backend', true);
	`)
	require.NoError(t, err)

	_, err = ts.Storage.Db.Exec(ctx, `
		INSERT INTO pull_requests (pull_request_id, pull_request_name, author_id, status, merged_at) VALUES 
			('pr-1', 'PR 1', 'u1', 'MERGED', $1);
	`, mergedAt)

	reqBody := map[string]interface{}{
		"pull_request_id": "pr-1",
	}

	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest("POST", "/pullRequest/merge", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	ts.Server.Handler.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err = json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	pr, ok := response["pr"].(map[string]interface{})
	require.True(t, ok)
	assert.Equal(t, "MERGED", pr["status"])
}

func TestPullRequestMerge_NotFound(t *testing.T) {
	ts, err := SetupTestServer(t)
	require.NoError(t, err)
	defer ts.Close()

	reqBody := map[string]interface{}{
		"pull_request_id": "nonexistent",
	}

	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest("POST", "/pullRequest/merge", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	ts.Server.Handler.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestPullRequestReassign_Success(t *testing.T) {
	ts, err := SetupTestServer(t)
	require.NoError(t, err)
	defer ts.Close()

	ctx := context.Background()
	_, err = ts.Storage.Db.Exec(ctx, `
		INSERT INTO team (name) VALUES ('backend');
		INSERT INTO users (user_id, username, team_name, is_active) VALUES 
			('u1', 'Alice', 'backend', true),
			('u2', 'Bob', 'backend', true),
			('u3', 'Charlie', 'backend', true);
		INSERT INTO pull_requests (pull_request_id, pull_request_name, author_id, status) VALUES 
			('pr-1', 'PR 1', 'u1', 'OPEN');
		INSERT INTO pr_reviewers (pull_request_id, user_id) VALUES ('pr-1', 'u2');
	`)
	require.NoError(t, err)

	reqBody := map[string]interface{}{
		"pull_request_id": "pr-1",
		"old_reviewer_id": "u2",
	}

	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest("POST", "/pullRequest/reassign", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	ts.Server.Handler.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err = json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	pr, ok := response["pr"].(map[string]interface{})
	require.True(t, ok)

	reviewers, ok := pr["assigned_reviewers"].([]interface{})
	require.True(t, ok)

	replacedBy, hasReplacedBy := response["replaced_by"]
	if len(reviewers) > 0 {
		assert.True(t, hasReplacedBy)
		assert.NotEqual(t, "u2", replacedBy)
	}
}

func TestPullRequestReassign_MergedPR(t *testing.T) {
	ts, err := SetupTestServer(t)
	require.NoError(t, err)
	defer ts.Close()

	ctx := context.Background()
	_, err = ts.Storage.Db.Exec(ctx, `
		INSERT INTO team (name) VALUES ('backend');
		INSERT INTO users (user_id, username, team_name, is_active) VALUES 
			('u1', 'Alice', 'backend', true),
			('u2', 'Bob', 'backend', true);
		INSERT INTO pull_requests (pull_request_id, pull_request_name, author_id, status) VALUES 
			('pr-1', 'PR 1', 'u1', 'MERGED');
		INSERT INTO pr_reviewers (pull_request_id, user_id) VALUES ('pr-1', 'u2');
	`)
	require.NoError(t, err)

	reqBody := map[string]interface{}{
		"pull_request_id": "pr-1",
		"old_reviewer_id": "u2",
	}

	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest("POST", "/pullRequest/reassign", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	ts.Server.Handler.ServeHTTP(w, req)

	assert.Equal(t, http.StatusConflict, w.Code)

	var response map[string]interface{}
	err = json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	errorObj, ok := response["error"].(map[string]interface{})
	require.True(t, ok)
	assert.Equal(t, "PR_MERGED", errorObj["code"])
}

func TestPullRequestReassign_NotAssigned(t *testing.T) {
	ts, err := SetupTestServer(t)
	require.NoError(t, err)
	defer ts.Close()

	ctx := context.Background()
	_, err = ts.Storage.Db.Exec(ctx, `
		INSERT INTO team (name) VALUES ('backend');
		INSERT INTO users (user_id, username, team_name, is_active) VALUES 
			('u1', 'Alice', 'backend', true),
			('u2', 'Bob', 'backend', true);
		INSERT INTO pull_requests (pull_request_id, pull_request_name, author_id, status) VALUES 
			('pr-1', 'PR 1', 'u1', 'OPEN');
		INSERT INTO pr_reviewers (pull_request_id, user_id) VALUES ('pr-1', 'u3');
	`)
	require.NoError(t, err)

	reqBody := map[string]interface{}{
		"pull_request_id": "pr-1",
		"old_reviewer_id": "u2",
	}

	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest("POST", "/pullRequest/reassign", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	ts.Server.Handler.ServeHTTP(w, req)

	assert.Equal(t, http.StatusConflict, w.Code)

	var response map[string]interface{}
	err = json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	errorObj, ok := response["error"].(map[string]interface{})
	require.True(t, ok)
	assert.Equal(t, "NOT_ASSIGNED", errorObj["code"])
}

func TestPullRequestReassign_NoCandidate(t *testing.T) {
	ts, err := SetupTestServer(t)
	require.NoError(t, err)
	defer ts.Close()

	ctx := context.Background()
	_, err = ts.Storage.Db.Exec(ctx, `
		INSERT INTO team (name) VALUES ('backend');
		INSERT INTO users (user_id, username, team_name, is_active) VALUES 
			('u1', 'Alice', 'backend', true),
			('u2', 'Bob', 'backend', true);
		INSERT INTO pull_requests (pull_request_id, pull_request_name, author_id, status) VALUES 
			('pr-1', 'PR 1', 'u1', 'OPEN');
		INSERT INTO pr_reviewers (pull_request_id, user_id) VALUES ('pr-1', 'u2');
	`)
	require.NoError(t, err)

	reqBody := map[string]interface{}{
		"pull_request_id": "pr-1",
		"old_reviewer_id": "u2",
	}

	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest("POST", "/pullRequest/reassign", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	ts.Server.Handler.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err = json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	pr, ok := response["pr"].(map[string]interface{})
	require.True(t, ok)

	reviewers, ok := pr["assigned_reviewers"].([]interface{})
	require.True(t, ok)

	_, hasReplacedBy := response["replaced_by"]
	assert.False(t, hasReplacedBy)
	assert.Len(t, reviewers, 0)
}

func TestPullRequestCreate_NoActiveReviewers(t *testing.T) {
	ts, err := SetupTestServer(t)
	require.NoError(t, err)
	defer ts.Close()

	ctx := context.Background()
	_, err = ts.Storage.Db.Exec(ctx, `
		INSERT INTO team (name) VALUES ('backend');
		INSERT INTO users (user_id, username, team_name, is_active) VALUES 
			('u1', 'Alice', 'backend', true);
	`)
	require.NoError(t, err)

	reqBody := map[string]interface{}{
		"pull_request_id":   "pr-1",
		"pull_request_name": "Add feature",
		"author_id":         "u1",
	}

	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest("POST", "/pullRequest/create", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	ts.Server.Handler.ServeHTTP(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)

	var response map[string]interface{}
	err = json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	pr, ok := response["pr"].(map[string]interface{})
	require.True(t, ok)

	reviewers, ok := pr["assigned_reviewers"].([]interface{})
	require.True(t, ok)
	assert.Len(t, reviewers, 0)
}

func TestPullRequestCreate_OnlyInactiveReviewers(t *testing.T) {
	ts, err := SetupTestServer(t)
	require.NoError(t, err)
	defer ts.Close()

	ctx := context.Background()
	_, err = ts.Storage.Db.Exec(ctx, `
		INSERT INTO team (name) VALUES ('backend');
		INSERT INTO users (user_id, username, team_name, is_active) VALUES 
			('u1', 'Alice', 'backend', true),
			('u2', 'Bob', 'backend', false);
	`)
	require.NoError(t, err)

	reqBody := map[string]interface{}{
		"pull_request_id":   "pr-1",
		"pull_request_name": "Add feature",
		"author_id":         "u1",
	}

	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest("POST", "/pullRequest/create", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	ts.Server.Handler.ServeHTTP(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)

	var response map[string]interface{}
	err = json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	pr, ok := response["pr"].(map[string]interface{})
	require.True(t, ok)

	reviewers, ok := pr["assigned_reviewers"].([]interface{})
	require.True(t, ok)
	assert.Len(t, reviewers, 0)
}
