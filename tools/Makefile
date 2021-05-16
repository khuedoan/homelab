.POSIX:

TAG = homelab-tools

default: check build run

check:
	docker info --format '{{ .Plugins.Network }}' | grep 'host'

build:
	docker build . --tag ${TAG}

run:
	docker run \
		--rm \
		--interactive \
		--tty \
		--env "TERM=${TERM}" \
		--env "HOME=${HOME}" \
		--volume "${HOME}:${HOME}" \
		--volume "/var/run/docker.sock:/var/run/docker.sock" \
		--workdir "$(shell pwd)/.." \
		${TAG}
