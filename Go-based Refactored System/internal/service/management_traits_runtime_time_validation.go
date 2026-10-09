package service

import "time"

// Validate the admitted participant credential against the server clock after
// all write locks are acquired. This does not renew credentials or deadlines.
func managementTraitsRuntimeCredentialTime(claims ManagementTraitsRuntimeClaims, now time.Time) error {
	if claims.ExpiresAt <= now.Unix() {
		return ErrManagementTraitsRuntimeToken
	}
	return nil
}
