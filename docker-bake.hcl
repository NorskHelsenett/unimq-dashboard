group "all" {
  targets = ["frontend", "backend"]
}

target "frontend" {
  context = "."
  dockerfile = "dockerfiles/Dockerfile.frontend"
  args = {
    NODE_VERSION = "22"
  }
  tags = ["nhn-unimq/frontend:latest"]
}

target "backend" {
  context = "."
  dockerfile = "dockerfiles/Dockerfile.api"
  args = {
    GO_VERSION = "1.25"
  }
  tags = ["nhn-unimq/backend:latest"]
}
