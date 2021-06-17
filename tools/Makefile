.POSIX:

WORKDIR = /usr/local/src
TAG = homelab-tools

default: build run

build:
	docker build . \
		--tag ${TAG}

run:
	docker run \
		--rm \
		--interactive \
		--tty \
		--network host \
		--env "TERM=${TERM}" \
		--volume "$(shell pwd)/..:${WORKDIR}" \
		--volume "${HOME}/.ssh/id_ed25519.pub:/root/.ssh/id_ed25519.pub" \
		--volume "/var/run/docker.sock:/var/run/docker.sock" \
		--workdir "${WORKDIR}" \
		${TAG}
