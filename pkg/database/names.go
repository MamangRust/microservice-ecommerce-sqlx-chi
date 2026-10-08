// Package database constants for the six bounded contexts of the ecommerce
// platform. Each context runs on its own PostgreSQL instance; the database
// keeps the context name so the same identifier works in env keys, migrations
// and dashboards. A service selects its own through the cluster prefix (see the
// *Cluster constants) so that one table has exactly one owner.
package database

const (
	// IdentityDB hosts auth, user and role tables.
	IdentityDB = "ec_identity"
	// CatalogDB hosts category and product tables.
	CatalogDB = "ec_catalog"
	// MerchantDB hosts merchant and merchant sub-domain tables.
	MerchantDB = "ec_merchant"
	// SalesDB hosts order, order_item and transaction tables.
	SalesDB = "ec_sales"
	// ExperienceDB hosts cart, shipping_address, banner, slider and review tables.
	ExperienceDB = "ec_experience"
	// EmailDB hosts the email consumer inbox.
	EmailDB = "ec_email"
)

// Cluster* are the env prefixes passed to NewClientWithPrefix. Each prefix
// resolves its own <Prefix>_HOST, _PORT, _USERNAME, _NAME and _PASSWORD — one
// PostgreSQL instance per bounded context, fronted by its own PgBouncer. Only
// the username and password fall back to the generic DB_* keys.
const (
	IdentityCluster   = "DB_IDENTITY"
	MerchantCluster   = "DB_MERCHANT"
	CatalogCluster    = "DB_CATALOG"
	SalesCluster      = "DB_SALES"
	ExperienceCluster = "DB_EXPERIENCE"
	EmailCluster      = "DB_EMAIL"
)
