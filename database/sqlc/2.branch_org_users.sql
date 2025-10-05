-- name: CreateBranchUser :exec
INSERT INTO branch_users (
    branch,
    user_id
) VALUES ( $1, $2 );

-- name: DeleteBranchUser :exec
DELETE FROM branch_users 
WHERE branch = $1 AND user_id = $2;


-- name: CreateOrganizationUser :exec
INSERT INTO organization_users (
    organization,
    user_id
) VALUES ( $1, $2 );

-- name: DeleteOrganizationUser :exec
DELETE FROM organization_users
WHERE organization = $1 AND user_id = $2;