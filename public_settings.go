package disk

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"time"
)

// Rights a public link can grant, for [PublicAccess.Rights].
const (
	RightRead                            = "read"
	RightReadWithoutDownload             = "read_without_download"
	RightReadWithPassword                = "read_with_password"
	RightReadWithPasswordWithoutDownload = "read_with_password_without_download"
	RightWrite                           = "write"
)

// Audiences of a public link, for [PublicAccess.Macros].
const (
	MacroAll       = "all"       // anyone with the link
	MacroEmployees = "employees" // only people in the organization OrgID
)

// PublicSettings protect a public link. Pass them to
// [Client.PublishResourceWithSettings] or [Client.UpdatePublicSettings].
//
// Password and ExpiresAt work on personal accounts too; this was checked
// against the live API, where the official reference differs in places.
// Forbidding downloads is not possible through the public API: the web
// interface uses an internal call for it.
type PublicSettings struct {
	// Password, when set, must be entered to open the link.
	Password string
	// ExpiresAt, when set, is the moment the link stops working.
	ExpiresAt time.Time
	// Accesses grant address access in Yandex 360 for business. Leave them
	// empty on a personal account: there they close the link to everyone.
	Accesses []PublicAccess
}

// PublicAccess grants rights over a public link to an audience in Yandex 360
// for business.
type PublicAccess struct {
	// Macros name the audience: [MacroAll] or [MacroEmployees].
	Macros []string `json:"macros,omitempty"`
	// OrgID is the organization meant by [MacroEmployees].
	OrgID int64 `json:"org_id,omitempty"`
	// Type is "macro" for an audience given by Macros, which is filled in
	// when left empty, or "user", "group", "department" or "outer_user"
	// together with ID.
	Type string `json:"type"`
	ID   int64  `json:"id,omitempty"`
	// Rights are what the audience may do, such as [RightRead].
	Rights []string `json:"rights"`
}

func (s *PublicSettings) empty() bool {
	return s == nil || (s.Password == "" && s.ExpiresAt.IsZero() && len(s.Accesses) == 0)
}

// body is the PATCH payload. The API takes the expiry as a Unix time in
// seconds, not as a duration as the reference says, and wants "type" on
// every access although the reference examples leave it out.
func (s *PublicSettings) body() ([]byte, error) {
	payload := map[string]any{}
	if s.Password != "" {
		payload["password"] = s.Password
	}
	if !s.ExpiresAt.IsZero() {
		payload["available_until"] = s.ExpiresAt.Unix()
	}
	if len(s.Accesses) > 0 {
		accesses := make([]PublicAccess, len(s.Accesses))
		copy(accesses, s.Accesses)
		for i := range accesses {
			if accesses[i].Type == "" && len(accesses[i].Macros) > 0 {
				accesses[i].Type = "macro"
			}
		}
		payload["accesses"] = accesses
	}
	return json.Marshal(payload)
}

// PublishResourceWithSettings publishes the file or folder at path and
// protects its link with settings. The API ignores settings sent along with
// publishing, so this publishes first and then applies them. If applying
// them fails, the resource is unpublished again rather than left open, and
// the error is returned. With empty settings it is the same as
// [Client.PublishResource].
func (c *Client) PublishResourceWithSettings(ctx context.Context, path string, settings *PublicSettings) (*Link, *ErrorResponse) {
	link, errResp := c.PublishResource(ctx, path)
	if errResp != nil || settings.empty() {
		return link, errResp
	}
	if errResp := c.UpdatePublicSettings(ctx, path, settings); errResp != nil {
		_, _ = c.UnpublishResource(ctx, path)
		return nil, errResp
	}
	return link, nil
}

// UpdatePublicSettings changes the protection of the already published file
// or folder at path. Settings left empty are not changed.
func (c *Client) UpdatePublicSettings(ctx context.Context, path string, settings *PublicSettings) *ErrorResponse {
	if len(path) < 1 {
		return &ErrorResponse{Error: "path cannot be empty"}
	}
	if settings.empty() {
		return &ErrorResponse{Error: "settings cannot be empty"}
	}

	body, err := settings.body()
	if err != nil {
		return &ErrorResponse{Error: fmt.Sprintf("failed to encode settings: %v", err)}
	}

	query := url.Values{}
	query.Set("path", path)

	resp, err := c.doRequest(ctx, PATCH, "public/resources/public-settings?"+query.Encode(), bytes.NewReader(body))
	if err != nil {
		return &ErrorResponse{Error: fmt.Sprintf("request failed: %v", err)}
	}
	defer func() { _ = resp.Body.Close() }()

	errResp, err := c.handleResponse(resp, []int{http.StatusOK, http.StatusNoContent})
	if errResp != nil {
		return errResp
	}
	if err != nil {
		return &ErrorResponse{Error: err.Error(), StatusCode: resp.StatusCode}
	}
	return nil
}

// PublicSettingsInfo is what the API reports about a link's protection.
type PublicSettingsInfo struct {
	// ExpiresAt is when the link stops working, or zero for never.
	ExpiresAt time.Time
}

// GetPublicSettings reads the protection of the published file or folder at
// path. The API reports only the expiry; whether a password is set cannot
// be read back.
func (c *Client) GetPublicSettings(ctx context.Context, path string) (*PublicSettingsInfo, *ErrorResponse) {
	if len(path) < 1 {
		return nil, &ErrorResponse{Error: "path cannot be empty"}
	}
	query := url.Values{}
	query.Set("path", path)

	raw, errResp := requestJSON[struct {
		AvailableUntil *int64 `json:"available_until"`
	}](ctx, c, GET, "public/resources/public-settings?"+query.Encode(), nil)
	if errResp != nil {
		return nil, errResp
	}
	info := &PublicSettingsInfo{}
	if raw.AvailableUntil != nil && *raw.AvailableUntil > 0 {
		info.ExpiresAt = time.Unix(*raw.AvailableUntil, 0)
	}
	return info, nil
}
