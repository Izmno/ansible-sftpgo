package sftpgo

import (
	"fmt"

	"github.com/sftpgo/sdk"
)

type UserInput struct {
	ClientConfig
	User   *User `json:"userdata"`
	Status int   `json:"status"`
}

// User is a partial SFTPGo user model that the client can
// interact with. It is a subset of the BaseUser model with
//
// - the ID field removed
// - all read-only fields removed
// - the password field removed
// - the UID, GID, and HomeDir fields removed
// - the transfer limits fields removed
type User struct {
	Status            int                 `json:"status"` // 1 enabled, 0 disabled (login is not allowed)
	Username          string              `json:"username"`
	Email             string              `json:"email,omitempty"`
	ExpirationDate    int64               `json:"expiration_date,omitempty"`
	PublicKeys        []string            `json:"public_keys,omitempty"`
	MaxSessions       int                 `json:"max_sessions"`
	QuotaSize         int64               `json:"quota_size"`  // bytes
	QuotaFiles        int                 `json:"quota_files"` // number of files
	Permissions       map[string][]string `json:"permissions"`
	UploadBandwidth   int64               `json:"upload_bandwidth,omitempty"`   // KB/s
	DownloadBandwidth int64               `json:"download_bandwidth,omitempty"` // KB/s
	Description       string              `json:"description,omitempty"`
	AdditionalInfo    string              `json:"additional_info,omitempty"`
	Groups            []sdk.GroupMapping  `json:"groups,omitempty"`
	Role              string              `json:"role,omitempty"`
}

func (u *User) Fix() {
	// Grant all permissions if they are not set
	if u.Permissions == nil {
		u.Permissions = map[string][]string{
			"/": {"*"},
		}
	}
}

func (u *User) AsSdkUser() *sdk.BaseUser {
	return &sdk.BaseUser{
		Status:            u.Status,
		Username:          u.Username,
		Email:             u.Email,
		ExpirationDate:    u.ExpirationDate,
		PublicKeys:        u.PublicKeys,
		MaxSessions:       u.MaxSessions,
		QuotaSize:         u.QuotaSize,
		QuotaFiles:        u.QuotaFiles,
		Permissions:       u.Permissions,
		UploadBandwidth:   u.UploadBandwidth,
		DownloadBandwidth: u.DownloadBandwidth,
		Description:       u.Description,
		AdditionalInfo:    u.AdditionalInfo,
		Groups:            u.Groups,
		Role:              u.Role,
	}

}

func (u *User) NeedsUpdate(target *User) bool {
	if u == nil {
		return true
	}

	if u.Status != target.Status ||
		(u.Email != target.Email && target.Email != "") ||
		(u.ExpirationDate != target.ExpirationDate && target.ExpirationDate != 0) ||
		u.MaxSessions != target.MaxSessions ||
		u.QuotaFiles != target.QuotaFiles ||
		u.QuotaSize != target.QuotaSize ||
		(u.UploadBandwidth != target.UploadBandwidth && target.UploadBandwidth != 0) ||
		(u.DownloadBandwidth != target.DownloadBandwidth && target.DownloadBandwidth != 0) ||
		(u.Description != target.Description && target.Description != "") ||
		(u.AdditionalInfo != target.AdditionalInfo && target.AdditionalInfo != "") ||
		(u.Role != target.Role && target.Role != "") {
		return true
	}

	if !equalStringList(u.PublicKeys, target.PublicKeys) {
		return true
	}

	if !equalStringListMaps(u.Permissions, target.Permissions) {
		return true
	}

	originalGroups := mapToStringListMap(
		u.Groups,
		func(g sdk.GroupMapping) string { return g.Name },
		func(g sdk.GroupMapping) string { return fmt.Sprintf("%d", g.Type) },
	)

	targetGroups := mapToStringListMap(
		target.Groups,
		func(g sdk.GroupMapping) string { return g.Name },
		func(g sdk.GroupMapping) string { return fmt.Sprintf("%d", g.Type) },
	)

	if !equalStringListMaps(originalGroups, targetGroups) {
		return true
	}

	return false
}

func NewUserFromSdkUser(sdkUser *sdk.BaseUser) *User {
	if sdkUser == nil {
		return nil
	}

	return &User{
		Status:            sdkUser.Status,
		Username:          sdkUser.Username,
		Email:             sdkUser.Email,
		ExpirationDate:    sdkUser.ExpirationDate,
		PublicKeys:        sdkUser.PublicKeys,
		MaxSessions:       sdkUser.MaxSessions,
		QuotaSize:         sdkUser.QuotaSize,
		QuotaFiles:        sdkUser.QuotaFiles,
		Permissions:       sdkUser.Permissions,
		UploadBandwidth:   sdkUser.UploadBandwidth,
		DownloadBandwidth: sdkUser.DownloadBandwidth,
		Description:       sdkUser.Description,
		AdditionalInfo:    sdkUser.AdditionalInfo,
		Groups:            sdkUser.Groups,
		Role:              sdkUser.Role,
	}
}

func equalStringList(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}

	m := make(map[string]struct{})
	for _, s := range a {
		m[s] = struct{}{}
	}

	for _, s := range b {
		if _, ok := m[s]; !ok {
			return false
		}
	}

	return true
}

func equalStringListMaps(a, b map[string][]string) bool {
	if len(a) != len(b) {
		return false
	}

	for k, v := range a {
		w, ok := b[k]
		if !ok {
			return false
		}

		if !equalStringList(v, w) {
			return false
		}
	}

	return true
}

func mapToStringListMap[T any](l []T, keyFunc func(T) string, valFunc func(T) string) map[string][]string {
	m := make(map[string][]string)
	for _, v := range l {
		key := keyFunc(v)

		if _, ok := m[key]; !ok {
			m[key] = []string{}
		}

		m[key] = append(m[key], valFunc(v))
	}

	return m
}
