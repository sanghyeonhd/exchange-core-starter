package auth

import "errors"

var (
	ErrPermissionDenied = errors.New("permission denied")
	ErrMFARequired      = errors.New("mfa required")
)

type Role string

const (
	RoleUser        Role = "USER"
	RoleSuperAdmin  Role = "SUPER_ADMIN"
	RoleRiskAdmin   Role = "RISK_ADMIN"
	RoleWalletAdmin Role = "WALLET_ADMIN"
	RoleMarketAdmin Role = "MARKET_ADMIN"
	RoleSupport     Role = "SUPPORT"
	RoleAuditor     Role = "AUDITOR"
	RoleKYCAnalyst  Role = "KYC_ANALYST"
	RoleAMLAnalyst  Role = "AML_ANALYST"
)

type Permission string

const (
	PermissionTrade             Permission = "TRADE"
	PermissionWithdraw          Permission = "WITHDRAW"
	PermissionReadAccount       Permission = "READ_ACCOUNT"
	PermissionReviewKYC         Permission = "REVIEW_KYC"
	PermissionReviewAML         Permission = "REVIEW_AML"
	PermissionApproveWithdrawal Permission = "APPROVE_WITHDRAWAL"
	PermissionManageMarkets     Permission = "MANAGE_MARKETS"
	PermissionManageListings    Permission = "MANAGE_LISTINGS"
	PermissionViewAudit         Permission = "VIEW_AUDIT"
	PermissionManageRisk        Permission = "MANAGE_RISK"
)

type Subject struct {
	ID         string
	Roles      []Role
	MFAEnabled bool
}

var rolePermissions = map[Role][]Permission{
	RoleUser:        {PermissionTrade, PermissionWithdraw, PermissionReadAccount},
	RoleSuperAdmin:  {PermissionReviewKYC, PermissionReviewAML, PermissionApproveWithdrawal, PermissionManageMarkets, PermissionManageListings, PermissionViewAudit, PermissionManageRisk},
	RoleRiskAdmin:   {PermissionReviewAML, PermissionViewAudit, PermissionManageRisk},
	RoleWalletAdmin: {PermissionApproveWithdrawal, PermissionViewAudit},
	RoleMarketAdmin: {PermissionManageMarkets, PermissionManageListings, PermissionViewAudit},
	RoleSupport:     {PermissionReadAccount},
	RoleAuditor:     {PermissionViewAudit},
	RoleKYCAnalyst:  {PermissionReviewKYC, PermissionViewAudit},
	RoleAMLAnalyst:  {PermissionReviewAML, PermissionViewAudit},
}

func Authorize(subject Subject, permission Permission) error {
	if isAdminPermission(permission) && !subject.MFAEnabled {
		return ErrMFARequired
	}
	for _, role := range subject.Roles {
		for _, candidate := range rolePermissions[role] {
			if candidate == permission {
				return nil
			}
		}
	}
	return ErrPermissionDenied
}

func isAdminPermission(permission Permission) bool {
	switch permission {
	case PermissionReviewKYC, PermissionReviewAML, PermissionApproveWithdrawal, PermissionManageMarkets, PermissionManageListings, PermissionViewAudit, PermissionManageRisk:
		return true
	default:
		return false
	}
}
