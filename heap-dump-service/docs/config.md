# Setup

In order to deploy the Heap Dump Service Hasicorp Vault, AWS IAM and kubernetes RBAC configuration is needed.

## Vault

```hcl
# Enable Transit Engine in Vault
resource "vault_mount" "heap_dump_encryption_mount" {
  path        = "eaas-heap-dump-service"
  type        = "transit"
  description = "Transit engine for encrypting heap dump AES Keys"
}

# Create Encryption key in transit engine for each tenant
resource "vault_transit_secret_backend_key" "heap_dump_service_backend" {
  backend = vault_mount.heap_dump_encryption_mount.path
  name    = var.tenant
  type    = "aes256-gcm96"

  exportable             = false
  allow_plaintext_backup = false
}

# For Service
resource "vault_policy" "heap_dump_service" {
  name   = format("topic-%s-heap-dump-service", var.tenant)
  policy = <<EOT
path "eaas-heap-dump-service/encrypt/*" {
  capabilities = [ "update" ]
}
EOT
}

resource "vault_kubernetes_auth_backend_role" "heap_dump_service_sa" {
  bound_service_account_names      = var.heap_dump_service_config.service_account_name
  bound_service_account_namespaces = var.heap_dump_service_config.namespace
  role_name                        = "heap-dump-service"
  token_policies                   = vault_policy.heap_dump_service.name
  backend                          = var.heap_dump_service_config.kubernetes_auth_backend
}

# Create this role for each tenant
resource "vault_policy" "heap_dump_decryption" {
  name   = format("topic-%s-heap-dump-service", var.tenant)
  policy = <<EOT
path "eaas-heap-dump-service/decrypt/${var.tenant}" {
  capabilities = [ "update"]
}
EOT
}

variable "heap_dump_service_config" {
  default = {
    service_account_name    = "heap-dump-service"
    namespace               = "heap-dump-service"
    kubernetes_auth_backend = "define_it"
  }
}

variable "tenant" {
  type    = string
  default = "java-squad-1"
}

```

## AWS

As we are using primarily AWS, the following IAM policy can be taken as an example, but the service should work with any other S3 compatible storage backend as well.

```hcl
data "aws_iam_policy_document" "heap_dump_service_role_policy" {
  statement {
    effect  = "Allow"
    actions = ["s3:*"]
    resources = [
      "arn:aws:s3:::${aws_s3_bucket.heap_dump_service.id}/*"
    ]
  }
}
```

## Kubernetes

In order to make use of the Authentication validation functionality of Kubernetes, the ServiceAccount of the Heap Dump Service need to have the correct permissions. The RBAC configuration is included by default in the helm chart, but in case this needs to be done manually, the following permissions are needed:

```yaml
kind: ClusterRole
apiVersion: rbac.authorization.k8s.io/v1
metadata:
  name: heap-dump-service-role
rules:
- apiGroups: ["authentication.k8s.io"]
  resources: ["tokenreviews", "subjectaccessreviews"]
  verbs: ["create"]
```

# Configuration

The heap dump service can be configured with a json config and environment variables.  
Here is an example json configuration: 

```json
{
    "app": {
        "port": 8080,
        "bucket": "test-bucket"
    },
    "vault": {
        "vaultTransitMount": "eaas-heap-dump-service",
        "vaultRole": "heap-dump-service",
        "vaultAuthMountPath": "kubernetes-auth-mount-path"
    },
    "serviceAccount": {
        "jwtokenMountPoint": "/var/run/secrets/kubernetes.io/serviceaccount/token"
    },
    "metrics": {
        "port": 8081,
        "path": "/metrics"
    }
}
```

this `config.json` file can be referenced by the environment variable `APP_CONFIG_JSON`.  
Other environment variables include: 

```yaml
- name: APP_CONFIG_FILE
  value: "/opt/config.json"
- name: VAULT_ADDR
  value: "https://my-vault.svc.cluster.local"
- name: GIN_MODE
  value: release
```
