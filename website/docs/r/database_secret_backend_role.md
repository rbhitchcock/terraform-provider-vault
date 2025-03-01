---
layout: "vault"
page_title: "Vault: vault_database_secret_backend_role resource"
sidebar_current: "docs-vault-resource-database-secret-backend-role"
description: |-
  Configures a database secret backend role for Vault.
---

# vault\_database\_secret\_backend\_role

Creates a Database Secret Backend role in Vault. Database secret backend
roles can be used to generate dynamic credentials for the database.

~> **Important** All data provided in the resource configuration will be
written in cleartext to state and plan files generated by Terraform, and
will appear in the console output when Terraform runs. Protect these
artifacts accordingly. See
[the main provider documentation](../index.html)
for more details.

## Example Usage

```hcl
resource "vault_mount" "db" {
  path = "postgres"
  type = "database"
}

resource "vault_database_secret_backend_connection" "postgres" {
  backend       = vault_mount.db.path
  name          = "postgres"
  allowed_roles = ["dev", "prod"]

  postgresql {
    connection_url = "postgres://username:password@host:port/database"
  }
}

resource "vault_database_secret_backend_role" "role" {
  backend             = vault_mount.db.path
  name                = "dev"
  db_name             = vault_database_secret_backend_connection.postgres.name
  creation_statements = ["CREATE ROLE \"{{name}}\" WITH LOGIN PASSWORD '{{password}}' VALID UNTIL '{{expiration}}';"]
}
```

## Argument Reference

The following arguments are supported:

* `name` - (Required) A unique name to give the role.

* `backend` - (Required) The unique name of the Vault mount to configure.

* `db_name` - (Required) The unique name of the database connection to use for
  the role.

* `creation_statements` - (Required) The database statements to execute when
  creating a user.

* `revocation_statements` - (Optional) The database statements to execute when
  revoking a user.

* `rollback_statements` - (Optional) The database statements to execute when
  rolling back creation due to an error.

* `renew_statements` - (Optional) The database statements to execute when
  renewing a user.

* `default_ttl` - (Optional) The default number of seconds for leases for this
  role.

* `max_ttl` - (Optional) The maximum number of seconds for leases for this
  role.

## Attributes Reference

No additional attributes are exported by this resource.

## Import

Database secret backend roles can be imported using the `backend`, `/roles/`, and the `name` e.g.

```
$ terraform import vault_database_secret_backend_role.example postgres/roles/my-role
```
