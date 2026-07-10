// Vikunja is a to-do list application to facilitate your life.
// Copyright 2018-present Vikunja and contributors. All rights reserved.
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU Affero General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU Affero General Public License for more details.
//
// You should have received a copy of the GNU Affero General Public License
// along with this program.  If not, see <https://www.gnu.org/licenses/>.

package caldavtests

import (
	"net/http"
	"strings"
	"testing"

	"code.vikunja.io/api/pkg/db"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Saved filter id 2 (owned by user15, filter "done = false") maps to the
// pseudo-project id -3. Task 40 (uid-caldav-test) lives in the real project 36
// and matches the filter, so it must be reachable through the filter collection.
const (
	savedFilterProjectPath = "/dav/projects/-3"
	savedFilterTaskHref    = "/dav/projects/-3/uid-caldav-test.ics"
	matchingTaskUID        = "uid-caldav-test"
	matchingTaskRealPath   = "/dav/projects/36/uid-caldav-test.ics"
)

// taskHrefs returns the response hrefs that point at a task resource (.ics).
func taskHrefs(ms Multistatus) []string {
	var hrefs []string
	for _, r := range ms.Responses {
		if strings.HasSuffix(r.Href, ".ics") {
			hrefs = append(hrefs, r.Href)
		}
	}
	return hrefs
}

func TestSavedFilterCaldavReport(t *testing.T) {
	t.Run("calendar-query returns matching tasks under the filter collection path", func(t *testing.T) {
		e := setupTestEnv(t)

		rec := caldavREPORT(t, e, savedFilterProjectPath, ReportCalendarQuery)

		assertResponseStatus(t, rec, 207)
		ms := parseMultistatus(t, rec)

		hrefs := taskHrefs(ms)
		require.NotEmpty(t, hrefs, "filter collection should report its matching tasks")
		for _, h := range hrefs {
			assert.Truef(t, strings.HasPrefix(h, savedFilterProjectPath+"/"),
				"task href %q should be addressed under the filter collection %q", h, savedFilterProjectPath)
		}
		assert.Contains(t, hrefs, savedFilterTaskHref,
			"the matching task should be reported under the filter collection path")
	})
}

func TestSavedFilterCaldavPropfind(t *testing.T) {
	t.Run("Depth 1 lists filter tasks under the filter collection path", func(t *testing.T) {
		e := setupTestEnv(t)

		rec := caldavPROPFIND(t, e, savedFilterProjectPath, "1", PropfindCalendarCollectionProperties)

		assertResponseStatus(t, rec, 207)
		ms := parseMultistatus(t, rec)

		hrefs := taskHrefs(ms)
		require.NotEmpty(t, hrefs, "Depth 1 PROPFIND on a filter should list its tasks")
		for _, h := range hrefs {
			assert.Truef(t, strings.HasPrefix(h, savedFilterProjectPath+"/"),
				"task href %q should be addressed under the filter collection", h)
		}
		assert.Contains(t, hrefs, savedFilterTaskHref)
	})
}

func TestSavedFilterCaldavMultiget(t *testing.T) {
	t.Run("calendar-multiget returns a task requested via the filter path", func(t *testing.T) {
		e := setupTestEnv(t)

		rec := caldavREPORT(t, e, savedFilterProjectPath, ReportCalendarMultiget(savedFilterTaskHref))

		assertResponseStatus(t, rec, 207)
		ms := parseMultistatus(t, rec)

		var found bool
		for _, r := range ms.Responses {
			if r.Href != savedFilterTaskHref {
				continue
			}
			found = true
			prop := getSuccessfulProp(t, r)
			assert.Contains(t, prop.CalendarData, matchingTaskUID,
				"multiget response should carry the task's calendar data")
		}
		assert.True(t, found, "multiget of a filter task should return it under the filter path")
	})
}

func TestSavedFilterCaldavGet(t *testing.T) {
	t.Run("GET a single task through the filter path", func(t *testing.T) {
		e := setupTestEnv(t)

		rec := caldavGET(t, e, savedFilterTaskHref)

		assert.Equal(t, http.StatusOK, rec.Code, "Body:\n%s", rec.Body.String())
		assert.Contains(t, rec.Body.String(), "BEGIN:VTODO")
		assert.Contains(t, rec.Body.String(), matchingTaskUID)
	})
}

func TestSavedFilterCaldavUpdate(t *testing.T) {
	t.Run("updating a task through the filter path keeps its real project", func(t *testing.T) {
		e := setupTestEnv(t)

		vtodo := NewVTodo(matchingTaskUID, "Updated via filter").Build()
		rec := caldavPUT(t, e, savedFilterTaskHref, vtodo)

		assert.Truef(t, rec.Code == http.StatusOK || rec.Code == http.StatusCreated || rec.Code == http.StatusNoContent,
			"update through filter path should succeed, got %d. Body:\n%s", rec.Code, rec.Body.String())

		// The task must stay in its real project, not be moved into the filter pseudo-project.
		db.AssertExists(t, "tasks", map[string]interface{}{
			"uid":        matchingTaskUID,
			"project_id": 36,
			"title":      "Updated via filter",
		}, false)
	})
}

func TestSavedFilterCaldavCreateRejected(t *testing.T) {
	t.Run("creating a new task in a filter collection is rejected", func(t *testing.T) {
		e := setupTestEnv(t)

		vtodo := NewVTodo("filter-create-uid", "Should not be created").Build()
		rec := caldavPUT(t, e, "/dav/projects/-3/filter-create-uid.ics", vtodo)

		assert.NotEqual(t, http.StatusCreated, rec.Code,
			"creating a task inside a filter collection must not succeed")
		db.AssertMissing(t, "tasks", map[string]interface{}{"uid": "filter-create-uid"})
	})
}

func TestSavedFilterCaldavForeignUser(t *testing.T) {
	t.Run("another user cannot read a filter they do not own", func(t *testing.T) {
		e := setupTestEnv(t)

		rec := caldavRequest(t, e, "REPORT", savedFilterProjectPath, ReportCalendarQuery, map[string]string{
			"Authorization": basicAuthHeader(testuser1.Username, fixturePassword),
		})

		assert.NotContains(t, rec.Body.String(), matchingTaskUID,
			"a non-owner must not receive tasks from someone else's filter")
	})
}
