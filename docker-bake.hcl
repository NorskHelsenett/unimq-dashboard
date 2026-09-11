group "default" {
  targets = ["frontend", "backend"]
}

group "release" {
  targets = ["frontend_release", "backend_release"]
}

# Comma separated list of platforms to build. CI overrides this so every
# platform is built natively on a runner of the matching architecture.
variable "PLATFORMS" {
  default = "linux/amd64,linux/arm64"
}

# Keeps the GitHub Actions cache entries of parallel per-platform builds apart.
variable "CACHE_SCOPE" {
  default = "local"
}

target "_common" {
  context = "."

  platforms = split(",", PLATFORMS)


  labels = {
    "org.opencontainers.image.source" = "https://github.com/NorskHelsenett/unimq-dashboard"
    "org.opencontainers.image.licenses" = "Apache-2.0"
  }
}

target "docker-metadata-action" {}


target "frontend" {
  inherits = ["_common"]

  dockerfile = "dockerfiles/Dockerfile.frontend"

  args = {
    NODE_VERSION = "22"
  }

  cache-from = [
    "type=gha,scope=frontend-${CACHE_SCOPE}"
  ]

  cache-to = [
    "type=gha,mode=max,scope=frontend-${CACHE_SCOPE}"
  ]
}

target "frontend_release" {
  inherits = ["docker-metadata-action", "frontend"]
}

target "backend" {
  inherits = ["_common"]

  dockerfile = "dockerfiles/Dockerfile.backend"

  args = {
    GO_VERSION = "1.27.0"
  }

  cache-from = [
    "type=gha,scope=backend-${CACHE_SCOPE}"
  ]

  cache-to = [
    "type=gha,mode=max,scope=backend-${CACHE_SCOPE}"
  ]
}

target "backend_release" {
  inherits = ["docker-metadata-action", "backend"]
}
