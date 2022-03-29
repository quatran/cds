package api

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http/httptest"
	"strings"
	"testing"

	yaml "github.com/ghodss/yaml"
	"github.com/ovh/cds/engine/api/rbac"
	"github.com/ovh/cds/engine/api/test"
	"github.com/ovh/cds/engine/api/test/assets"
	"github.com/ovh/cds/sdk"
	"github.com/stretchr/testify/require"
)

func Test_crudVCSOnProjectLambdaUserForbidden(t *testing.T) {
	api, db, _ := newTestAPI(t)

	u, pass := assets.InsertLambdaUser(t, db)
	proj := assets.InsertTestProject(t, db, api.Cache, sdk.RandomString(10), sdk.RandomString(10))

	vars := map[string]string{
		"projectKey": proj.Key,
	}
	uriPost := api.Router.GetRouteV2("POST", api.postVCSProjectHandler, vars)
	test.NotEmpty(t, uriPost)

	req := assets.NewAuthentifiedRequest(t, u, pass, "POST", uriPost, nil)

	w := httptest.NewRecorder()
	api.Router.Mux.ServeHTTP(w, req)
	require.Equal(t, 403, w.Code)

	uriGet := api.Router.GetRouteV2("GET", api.getVCSProjectHandler, vars)
	test.NotEmpty(t, uriGet)

	reqGet := assets.NewAuthentifiedRequest(t, u, pass, "GET", uriGet, nil)
	w2 := httptest.NewRecorder()
	api.Router.Mux.ServeHTTP(w2, reqGet)
	require.Equal(t, 403, w2.Code)
}

func Test_crudVCSOnProjectLambdaUserOK(t *testing.T) {
	api, db, _ := newTestAPI(t)

	proj := assets.InsertTestProject(t, db, api.Cache, sdk.RandomString(10), sdk.RandomString(10))
	user1, pass := assets.InsertLambdaUser(t, db)

	perm := fmt.Sprintf(`name: perm-test-%s
projects:
  - role: manage
    projects: [%s]
    users: [%s]
`, proj.Key, proj.Key, user1.Username)

	var rb sdk.RBAC
	require.NoError(t, yaml.Unmarshal([]byte(perm), &rb))

	rb.Projects[0].RBACProjectsIDs = []int64{proj.ID}
	rb.Projects[0].RBACUsersIDs = []string{user1.ID}

	require.NoError(t, rbac.Insert(context.Background(), db, &rb))

	vars := map[string]string{
		"projectKey": proj.Key,
	}
	uri := api.Router.GetRouteV2("POST", api.postVCSProjectHandler, vars)
	test.NotEmpty(t, uri)
	req := assets.NewAuthentifiedRequest(t, user1, pass, "POST", uri, nil)

	body := `version: v1.0
name: my_vcs_server
type: bitbucketserver
description: "it's the test vcs server on project"
url: "http://my-vcs-server.localhost"
auth:
    user: the-username
    password: the-password
`

	// Here, we insert the vcs server as a CDS user (not administrator)
	req.Body = io.NopCloser(strings.NewReader(body))
	req.Header.Set("Content-Type", "application/x-yaml")

	w := httptest.NewRecorder()
	api.Router.Mux.ServeHTTP(w, req)
	require.Equal(t, 201, w.Code)

	// Then, get the vcs server
	uriGet := api.Router.GetRouteV2("GET", api.getVCSProjectHandler, vars)
	test.NotEmpty(t, uriGet)

	reqGet := assets.NewAuthentifiedRequest(t, user1, pass, "GET", uriGet, nil)
	w2 := httptest.NewRecorder()
	api.Router.Mux.ServeHTTP(w2, reqGet)
	require.Equal(t, 200, w2.Code)

	vcsProjects := []sdk.VCSProject{}
	require.NoError(t, json.Unmarshal(w2.Body.Bytes(), &vcsProjects))
	require.Len(t, vcsProjects, 1)
}

func Test_crudVCSOnProjectAdminOk(t *testing.T) {
	api, db, _ := newTestAPI(t)

	u, pass := assets.InsertAdminUser(t, db)
	proj := assets.InsertTestProject(t, db, api.Cache, sdk.RandomString(10), sdk.RandomString(10))

	vars := map[string]string{
		"projectKey": proj.Key,
	}
	uri := api.Router.GetRouteV2("POST", api.postVCSProjectHandler, vars)
	test.NotEmpty(t, uri)
	req := assets.NewAuthentifiedRequest(t, u, pass, "POST", uri, nil)

	body := `version: v1.0
name: my_vcs_server
type: bitbucketserver
description: "it's the test vcs server on project"
url: "http://my-vcs-server.localhost"
auth:
    user: the-username
    password: the-password
`

	// Here, we insert the vcs server as a CDS administrator
	req.Body = io.NopCloser(strings.NewReader(body))
	req.Header.Set("Content-Type", "application/x-yaml")

	w := httptest.NewRecorder()
	api.Router.Mux.ServeHTTP(w, req)
	require.Equal(t, 201, w.Code)

	// Then, get the vcs server
	uriGet := api.Router.GetRouteV2("GET", api.getVCSProjectHandler, vars)
	test.NotEmpty(t, uriGet)

	reqGet := assets.NewAuthentifiedRequest(t, u, pass, "GET", uriGet, nil)
	w2 := httptest.NewRecorder()
	api.Router.Mux.ServeHTTP(w2, reqGet)
	require.Equal(t, 200, w2.Code)

	vcsProjects := []sdk.VCSProject{}
	require.NoError(t, json.Unmarshal(w2.Body.Bytes(), &vcsProjects))
	require.Len(t, vcsProjects, 1)
}
