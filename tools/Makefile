.POSIX:

TAG = homelab-tools

default: build run

build:
	@docker build . --tag ${TAG}

run:
	@docker run \
		--rm \
		--interactive \
		--tty \
		--network host \
		--env "TERM=${TERM}" \
		--env "HOME=${HOME}" \
		--volume "${HOME}:${HOME}" \
		--volume "/var/run/docker.sock:/var/run/docker.sock" \
		--workdir "$(shell pwd)/.." \
		${TAG}
