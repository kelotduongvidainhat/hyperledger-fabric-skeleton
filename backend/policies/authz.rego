package ownership.authz

import future.keywords.if
import future.keywords.in

# Default deny
default allow := false

# Helper: Is Admin
is_admin := role if {
    role := input.user.role
    role == "admin"
}

# Helper: Is Auditor
is_auditor := role if {
    role := input.user.role
    role == "auditor"
}

# 1. Admin allows everything
allow if is_admin

# 2. Public Assets (Read access for everyone)
# Assuming path is like /assets or /assets/:id
allow if {
    input.request.method == "GET"
    startswith(input.request.path, "/assets")
}

# 3. User Asset Management (Create)
allow if {
    input.request.method == "POST"
    input.request.path == "/assets"
    input.user.role == "user"
}

# 4. Auditor Access
# Auditor can read anything, including admin stats
allow if {
    is_auditor
    input.request.method == "GET"
}

# 5. IPFS Uploads (Authenticated users)
allow if {
    input.request.path == "/api/ipfs/upload"
    input.user.role in ["user", "admin"]
}

# 6. Specific Asset Modification (Transfer/Delete)
# For now, let's keep it simple: admin only via the /admin group
# User specific logic (Propose/Accept) can be added here once we extend the input with asset ownership data
