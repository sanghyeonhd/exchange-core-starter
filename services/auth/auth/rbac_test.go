package auth

import (
	"errors"
	"testing"
)

func TestAuthorizeRequiresMFAForAdminPermission(t *testing.T) {
	err := Authorize(Subject{ID: "admin-1", Roles: []Role{RoleKYCAnalyst}}, PermissionReviewKYC)
	if !errors.Is(err, ErrMFARequired) {
		t.Fatalf("err = %v, want %v", err, ErrMFARequired)
	}
}

func TestAuthorizeAllowsRolePermission(t *testing.T) {
	err := Authorize(Subject{ID: "admin-1", Roles: []Role{RoleKYCAnalyst}, MFAEnabled: true}, PermissionReviewKYC)
	if err != nil {
		t.Fatalf("Authorize: %v", err)
	}
}

func TestAuthorizeRejectsMissingPermission(t *testing.T) {
	err := Authorize(Subject{ID: "support-1", Roles: []Role{RoleSupport}, MFAEnabled: true}, PermissionManageMarkets)
	if !errors.Is(err, ErrPermissionDenied) {
		t.Fatalf("err = %v, want %v", err, ErrPermissionDenied)
	}
}
