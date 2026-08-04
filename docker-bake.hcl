group "default" {
  targets = ["frontend", "backend"]
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

target "frontend" {
  inherits = ["_common"]

  dockerfile = "dockerfiles/Dockerfile.frontend"

  args = {
    NODE_VERSION = "22"
  }
}

target "backend" {
  inherits = ["_common"]

  dockerfile = "dockerfiles/Dockerfile.api"

  args = {
    GO_VERSION = "1.26.4"
  }
}
