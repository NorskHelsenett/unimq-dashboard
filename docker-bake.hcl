group "default" {
  targets = ["frontend", "backend"]
}

group "release" {
  targets = ["frontend_release", "backend_release"]
}

target "_common" {
  context = "."

  platforms = [
    "linux/amd64",
    "linux/arm64"
  ]


  labels = {
    "org.opencontainers.image.source" = "https://github.com/NorskHelsenett/unimq-dashboard"
    "org.opencontainers.image.licenses" = "Apache-2.0"
  }


  cache-from = [
    "type=gha"
  ]

  cache-to = [
    "type=gha,mode=max"
  ]
}

target "docker-metadata-action" {}


target "frontend" {
  inherits = ["_common"]

  dockerfile = "dockerfiles/Dockerfile.frontend"

  args = {
    NODE_VERSION = "22"
  }
}

target "frontend_release" {
  inherits = ["docker-metadata-action", "frontend"]
}

target "backend" {
  inherits = ["_common"]

  dockerfile = "dockerfiles/Dockerfile.backend"

  args = {
    GO_VERSION = "1.26.4"
  }
}

target "backend_release" {
  inherits = ["docker-metadata-action", "backend"]
}
