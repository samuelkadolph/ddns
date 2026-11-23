BUILD_DATE = ENV.fetch("BUILD_DATE", Time.now.utc.iso8601)
COMPOSE_DEV = ENV.fetch("COMPOSE_DEV", "docker-compose.dev.yml")
IMAGE_NAME = ENV.fetch("IMAGE_NAME", "samuelkadolph/ddns")
PLATFORMS = ENV.fetch("PLATFORMS", "linux/amd64,linux/arm64")
VCS_REF = ENV.fetch("VCS_REF", `git rev-parse --short HEAD`)
VERSION = ENV.fetch("VERSION", File.read("./VERSION").chomp)

task default: "run"

task "down" do
  sh %Q[docker compose -f "#{COMPOSE_DEV}" down]
end

task "push" do
  sh %Q[docker buildx create --name ddns-builder --driver docker-container --bootstrap --use]
  sh %Q[docker buildx build --platform "#{PLATFORMS}" --build-arg "BUILD_DATE=#{BUILD_DATE}" --build-arg "VCS_REF=#{VCS_REF}" --build-arg "VERSION=#{VERSION}" --tag "#{IMAGE_NAME}:latest" --push .]
  sh %Q[docker buildx build --platform "#{PLATFORMS}" --build-arg "BUILD_DATE=#{BUILD_DATE}" --build-arg "VCS_REF=#{VCS_REF}" --build-arg "VERSION=#{VERSION}" --tag "#{IMAGE_NAME}:#{VERSION}" --push .]
  sh %Q[docker buildx rm ddns-builder]
end

task "run" do
  sh %Q[docker compose -f "#{COMPOSE_DEV}" up --build]
end

task "up" do
  sh %Q[docker compose -f "#{COMPOSE_DEV}" up --build --detach]
end
