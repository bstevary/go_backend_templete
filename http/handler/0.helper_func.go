package handler

import (
	"slices"
	"sort"
	"strings"

	"github.com/bstevary/hexagonal/database/model"
)

func getPermissionsForSelectedScope(
	rows []model.GetUserRolesWithPermissionsRow, selectedRef int64, system bool, typ string,
) (refIDs []int64, actualSelectedRef int64, flatPermissions []string) {

	scopeKey := "ORG"
	if typ == "INTERNAL" {
		scopeKey = "BRANCH"
	}

	refAllPerms := make(map[int64]map[string]struct{})
	allRefIDsSet := make(map[int64]struct{})
	minNonZeroRefID := int64(0)

	for _, row := range rows {
		if row.Scope != scopeKey {
			continue
		}
		allRefIDsSet[row.Reference] = struct{}{}
		if _, ok := refAllPerms[row.Reference]; !ok {
			refAllPerms[row.Reference] = make(map[string]struct{})
		}
		refAllPerms[row.Reference][row.Permission] = struct{}{}
		if row.Reference != 0 {
			if minNonZeroRefID == 0 || row.Reference < minNonZeroRefID {
				minNonZeroRefID = row.Reference
			}
		}
	}

	if selectedRef == 0 {
		if system {
			actualSelectedRef = 0
		} else {
			actualSelectedRef = minNonZeroRefID
		}
	} else {
		if _, exists := allRefIDsSet[selectedRef]; exists {
			actualSelectedRef = selectedRef
		} else {
			actualSelectedRef = minNonZeroRefID
		}
	}

	for id := range allRefIDsSet {
		refIDs = append(refIDs, id)
	}
	slices.Sort(refIDs)

	if permsSet, ok := refAllPerms[actualSelectedRef]; ok {
		for p := range permsSet {
			flatPermissions = append(flatPermissions, p)
		}
		sort.Strings(flatPermissions)
	}

	return refIDs, actualSelectedRef, flatPermissions
}

func wildcardString(s *string) {
	tokens := strings.Split(strings.TrimSpace(*s), " ")
	if len(tokens) > 0 {
		*s = "%" + tokens[0] + "%"
	}
}
